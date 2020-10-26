package pi_modem

import (
	"github.com/warthog618/modem/at"
	"github.com/warthog618/modem/gsm"
	"github.com/warthog618/modem/serial"
	"github.com/warthog618/modem/trace"
)

func InitModemV2(cfg *ModemConfig, opts []gsm.Option) (*gsm.GSM, error) {

	serial, err := serial.New(serial.WithPort(cfg.Device), serial.WithBaud(cfg.Baud))
	if err != nil {
		return nil, err
	}

	modem := gsm.New(at.New(trace.New(serial), at.WithTimeout(cfg.DefaultTimeout)), opts...)
	if err = modem.Init(); err != nil {
		return nil, err
	}

	return modem, nil
}
