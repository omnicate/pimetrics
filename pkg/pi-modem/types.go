package pi_modem


type SMS struct {
	number	string `json:"number"`
	text	string `json:"text"`
}

type Call struct {
	number	string
	input	[]byte
}