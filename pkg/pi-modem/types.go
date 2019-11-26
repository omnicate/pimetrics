package pi_modem

type SMS struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

type Call struct {
	Number string `json:"number"`
	Input  []byte `json:"input,omitempty"`
}
