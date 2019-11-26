package pi_modem

import (
	"errors"
	"github.com/tarm/serial"
	"log"
)


func Dummy() {
	println("Some stuff")
}


func SendCommand(command string) (error, string) {
	c := &serial.Config{Name: "/dev/serial0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		return errors.New("SendCommandInit: " + err.Error()), ""
	}

	n, err := s.Write([]byte(command))
	if err != nil {
		return errors.New("SendCommandWrite: " + err.Error()), ""
	}

	buf := make([]byte, 128)
	n, err = s.Read(buf)
	if err != nil {
		return errors.New("SendCommandRead: " + err.Error()), ""
	}
	log.Printf("%q", buf[:n])

	return nil, string(buf[:n])
}