package pi_modem

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/tarm/serial"

	log "github.com/sirupsen/logrus"
)

const (
	SLEEP_TIME_MS   = 100 * time.Millisecond
	EXPORT_GPIO     = "/sys/class/gpio/export"
	DIRECTION_GPIO4 = "/sys/class/gpio/gpio4/direction"
	VALUE_GPIO4     = "/sys/class/gpio/gpio4/value"

	DIRECTION_GPIO6 = "/sys/class/gpio/gpio6/direction"
	VALUE_GPIO6     = "/sys/class/gpio/gpio6/value"

	MODE = 0644
)

func InitModem() error {
	if err := ioutil.WriteFile(EXPORT_GPIO, []byte("4"), MODE); err != nil {
		return fmt.Errorf("Failed Writing '4' to %s with %v", EXPORT_GPIO, err)
	}
	time.Sleep(SLEEP_TIME_MS)

	if err := ioutil.WriteFile(DIRECTION_GPIO4, []byte("out"), MODE); err != nil {
		return fmt.Errorf("Failed writing 'out' to %s with %v", DIRECTION_GPIO4, err)
	}

	if err := ioutil.WriteFile(VALUE_GPIO4, []byte("0"), MODE); err != nil {
		return fmt.Errorf("Failed writing '0' to %s with %v", VALUE_GPIO4, err)
	}
	if err := ioutil.WriteFile(EXPORT_GPIO, []byte("6"), MODE); err != nil {
		return fmt.Errorf("Failed writing '6' to %s with %v", EXPORT_GPIO, err)
	}
	time.Sleep(SLEEP_TIME_MS)

	if err := ioutil.WriteFile(DIRECTION_GPIO6, []byte("out"), MODE); err != nil {
		return fmt.Errorf("Failed writing 'out' to %s with %v", DIRECTION_GPIO6, err)
	}
	if err := ioutil.WriteFile(VALUE_GPIO6, []byte("0"), MODE); err != nil {
		return fmt.Errorf("Failed writing '0' to %s with %v", VALUE_GPIO6, err)
	}

	return nil
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
