package pi_modem

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/modem/at"
	"github.com/warthog618/modem/gsm"
	"github.com/warthog618/modem/serial"
)

const (
	CTRL_Z    = string(26)
	BREAKLINE = string(13)
	NEWLINE   = string(10)

	MODEM_BREAK = BREAKLINE + NEWLINE
)

var (
	wsClients map[*websocket.Conn]bool
)

type PiModem struct {
	*gsm.GSM
}

func InitModem(
	clients map[*websocket.Conn]bool,
	cfg *ModemConfig,
	opts []gsm.Option) (*PiModem, error) {

	log.WithFields(log.Fields{
		"modem_config": cfg,
	}).Info("Modem being initialised")

	serial, err := serial.New(serial.WithPort(cfg.Device), serial.WithBaud(cfg.Baud))
	if err != nil {
		return nil, err
	}

	modem := &PiModem{
		gsm.New(at.New(serial, at.WithTimeout(cfg.DefaultTimeout)), opts...),
	}
	if err = modem.Init(); err != nil {
		return nil, err
	}

	wsClients = clients

	modem.ReceiveMode()
	modem.HandleIncomingCall()
	modem.HandleNoCarrier()

	return modem, nil
}

func (g *PiModem) Call(number string, options ...at.CommandOption) (rsp []string, err error) {
	cmd := fmt.Sprintf("D%s;", number) + MODEM_BREAK
	r, err := g.Command(cmd, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed executing call command")
	}

	return r, nil
}

func (g *PiModem) Hangup() (rsp []string, err error) {
	cmd := "H0" + MODEM_BREAK
	r, err := g.Command(cmd, []at.CommandOption{}...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed hanging up call")
	}
	return r, nil
}

func (g *PiModem) ReceiveMode() {
	_ = g.StartMessageRx(
		func(msg gsm.Message) {
			log.WithField("message", msg.Message).Infof("Received SMS from %s", msg.Number)
			if msg.Message == "OMG_MAGIC_STUFF_!!_one_eleven_!!!" {
				_, _ = g.SendShortMessage(msg.Number, "NO_MAGIC_JUST_TURTLES")
			}
			for client := range wsClients {
				err := client.WriteJSON(SMS{
					Text:   msg.Message,
					Number: msg.Number,
				})
				if err != nil {
					log.WithError(err).Error("Failed sending sms message to web socket")
					client.Close()
					delete(wsClients, client)
				}
			}
		},
		func(err error) {
			log.WithError(err).Error("Failed reciving sms")
		})
}

func (g *PiModem) HandleIncomingCall() {
	g.AddIndication("RING", func(info []string) {
		// Answer call
		res, err := g.Command("A")
		if err != nil {
			log.WithError(err).Error("Failed answering call")
		}
		log.WithField("result", res).Info("Answered call")

		for {
			select {
			case <-time.After(time.Second * 5):
				res, err := g.Command("H0")
				if err != nil {
					log.WithError(err).Error("Failed hangingup call")
				}
				log.WithField("result", res).Info("Hanged-up call")
				return
			}
		}
	})
}

func (g *PiModem) HandleNoCarrier() {
	g.AddIndication("NO CARRIER", func(info []string) {

		log.WithField("result", info).Info("Got a no carrier")
	})
}

func (g *PiModem) SendSMS(number string, message string, options ...at.CommandOption) (rsp string, err error) {

	rsp, err = g.SendShortMessage(number, message, options...)

	return rsp, err
}
