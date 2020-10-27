package pi_modem

import "time"

type ModemConfig struct {
	Baud           int           `yaml:"baud"`
	Device         string        `yaml:"device"`
	DefaultTimeout time.Duration `yaml:"timeout,omitempty"`
	InitCmd        string        `yaml:"init_command,omitempty"`
}

type SMS struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

type Call struct {
	Number string `json:"number"`
	Input  []byte `json:"input,omitempty"`
}
