package pi_modem


type SMS struct {
	number	string
	text	string
}

type Call struct {
	number	string
	input	[]byte
}