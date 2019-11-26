package pi_modem

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
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

	n, err := s.Write([]byte(command+"\r\n"))
	if err != nil {
		return errors.New("SendCommandWrite: " + err.Error()), ""
	}

	read := 0
	var res []byte
	for {
		buf := make([]byte, 512)
		n, err = s.Read(buf)
		if err != nil {
			return errors.New("SendCommandRead: " + err.Error()), ""
		}

		read += n
		res = append(res, buf[:n]...)
		if read >= 4 && reflect.DeepEqual(res[read-4:read], []byte("OK\r\n")) {
			break
		}
		if read >= 7 && reflect.DeepEqual(res[read-7:read], []byte("ERROR\r\n")) {
			return errors.New("SendCommandRead: AT " +   command + "failed:" + string(res)), ""
		}
	}

	return nil, string(res[:read])
}

func SendSMS(sms SMS) (error, string) {
	return nil, "dummy"
}

func MakeCall(call Call) (error, string) {
	return nil, "dummy"
}
