package pi_modem

import (
	"fmt"

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

type PiModem struct {
	*gsm.GSM
}

func InitModemV2(cfg *ModemConfig, opts []gsm.Option) (*PiModem, error) {

	log.WithFields(log.Fields{
		"modem_config": cfg,
	}).Info("Modem being initialised")

	serial, err := serial.New(serial.WithPort(cfg.Device), serial.WithBaud(cfg.Baud))
	if err != nil {
		return nil, err
	}

	// modem := gsm.New(at.New(trace.New(serial), at.WithTimeout(cfg.DefaultTimeout)), opts...)
	modem := &PiModem{
		gsm.New(at.New(serial, at.WithTimeout(cfg.DefaultTimeout)), opts...),
	}
	if err = modem.Init(); err != nil {
		return nil, err
	}

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

func (g *PiModem) Handup() (rsp []string, err error) {
	cmd := "H" + MODEM_BREAK
	r, err := g.Command(cmd, []at.CommandOption{}...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed hanging up call")
	}
	return r, nil
}
