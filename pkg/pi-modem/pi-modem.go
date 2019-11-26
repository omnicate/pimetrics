package pi_modem

import (
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"os"
	"reflect"
	"time"

	"github.com/tarm/serial"
)

const (
	SLEEP_TIME_MS   = 100 * time.Millisecond
	EXPORT_GPIO     = "/sys/class/gpio/export"
	DIRECTION_GPIO4 = "/sys/class/gpio/gpio4/direction"
	VALUE_GPIO4     = "/sys/class/gpio/gpio4/value"

	DIRECTION_GPIO6 = "/sys/class/gpio/gpio6/direction"
	VALUE_GPIO6     = "/sys/class/gpio/gpio6/value"

	MODE = 0644

	CTRL_Z    = string(26)
	BREAKLINE = string(13)
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func ModemInitialized() bool {
	if fileExists("/sys/class/gpio/gpio4") && fileExists("/sys/class/gpio/gpio6") {
		return true
	}
	return false
}

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

func SendNonBlockCommand(command string) error {
	c := &serial.Config{Name: "/dev/serial0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		return errors.New("SendCommandInit: " + err.Error())
	}

	_, err = s.Write([]byte(command))
	if err != nil {
		return errors.New("SendCommandWrite: " + err.Error())
	}

	return err
}

func SendCommand(command string) (string, error) {
	c := &serial.Config{Name: "/dev/serial0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		return "", errors.New("SendCommandInit: " + err.Error())
	}

	n, err := s.Write([]byte(command))
	if err != nil {
		return "", errors.New("SendCommandWrite: " + err.Error())
	}

	read := 0
	var res []byte
	for {
		buf := make([]byte, 512)
		n, err = s.Read(buf)
		if err != nil {
			return "", errors.New("SendCommandRead: " + err.Error())
		}

		read += n
		res = append(res, buf[:n]...)
		if read >= 4 && reflect.DeepEqual(res[read-4:read], []byte("OK\r\n")) {
			break
		}
		if read >= 7 && reflect.DeepEqual(res[read-7:read], []byte("ERROR\r\n")) {
			return "", errors.New("SendCommandRead: AT " + command + "failed:" + string(res))
		}
	}

	return string(res[:read]), nil
}

func SendSMS(sms SMS) (string, error) {
	if sms.Number == "" || sms.Text == "" {
		return "", errors.New("SendSMS: Number or Message is empty")
	}

	if len(sms.Text) >= 140 {
		return "", errors.New("SendSMS: Message is too long")
	}

	if setTextMode() {

		// AT+CMGS=<number><CR><message><CTRL-Z>
		cmd := "AT+CMGS=\"" + sms.Number + "\""+ BREAKLINE

		fmt.Printf("===========%s\n", cmd)

		err := SendNonBlockCommand(cmd)
		if err != nil {
			return "", errors.New("SendSMS: Failed to send SMS Part1. Reason: " + err.Error())
		}

		// Mui importante
		time.Sleep(1 * time.Second)

		cmd = sms.Text + CTRL_Z

		fmt.Printf("===========%s\n", cmd)

		rv, err := SendCommand(cmd)
		if err != nil {
			return "", errors.New("SendSMS: Failed to send SMS Part2. Reason: " + err.Error())
		}


		return rv, nil
	}

	return "", errors.New("SendSMS: Could not set TextMode")
}

func setTextMode() bool {
	_, err := SendCommand("AT+CMGF=1" + BREAKLINE)
	if err != nil {
		log.Info("SetTextMode: Failed to set Text Mode. Reason: " + err.Error())
		return false
	}
	return true
}

func MakeCall(call Call) (error, string) {
	return nil, "dummy"
}
