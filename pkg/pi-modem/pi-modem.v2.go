package pi_modem

import (
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/modem/at"
	"github.com/warthog618/modem/gsm"
	"github.com/warthog618/modem/serial"
)

const (
	CTRL_Z    = string(26)
	BREAKLINE = string(13)
)

func InitModemV2(cfg *ModemConfig, opts []gsm.Option) (*gsm.GSM, error) {

	log.WithFields(log.Fields{
		"modem_config": cfg,
	}).Info("Modem being initialised")

	serial, err := serial.New(serial.WithPort(cfg.Device), serial.WithBaud(cfg.Baud))
	if err != nil {
		return nil, err
	}

	// modem := gsm.New(at.New(trace.New(serial), at.WithTimeout(cfg.DefaultTimeout)), opts...)
	modem := gsm.New(at.New(serial, at.WithTimeout(cfg.DefaultTimeout)), opts...)
	if err = modem.Init(); err != nil {
		return nil, err
	}

	return modem, nil
}
