package pi_modem

import "time"

type ModemConfig struct {
	Baud           int
	Device         string
	DefaultTimeout time.Duration
}

type SMS struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

type Call struct {
	Number string `json:"number"`
	Input  []byte `json:"input,omitempty"`
}
