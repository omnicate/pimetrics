package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	modem "pimetrics/pkg/pi-modem"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/modem/at"
)

func HandleSendSMSV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST requests are valid", http.StatusBadRequest)
		return
	}

	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	var sms modem.SMS

	err := json.NewDecoder(r.Body).Decode(&sms)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"number": sms.Number,
		"text":   sms.Text,
	}).Info("Sending sms with the following details")

	res, err := gModem.SendShortMessage(sms.Number, sms.Text)

	if err != nil {
		log.WithError(err).Error("SendSMS failed")
		http.Error(w, fmt.Sprintf("SendSMS failed with %v", err),
			http.StatusInternalServerError)
		SMSSentError.With(prometheus.Labels{"number": sms.Number}).Inc()
		return
	}

	SMSSentSuccess.With(prometheus.Labels{"number": sms.Number}).Inc()

	fmt.Fprint(w, res)
}

func HandleSendCommandV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST requests are valid", http.StatusBadRequest)
		return
	}

	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	cmd, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed getting command from body with %v", err),
			http.StatusBadRequest)
		return
	}

	// output, err := modem.SendCommand(string(cmd) + modem.BREAKLINE + modem.CTRL_Z)
	output, err := gModem.Command(string(cmd), []at.CommandOption{}...)
	if err != nil {
		log.WithError(err).Error("SendCommand failed")
		http.Error(w, fmt.Sprintf("SendCommand failed with %v", err),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, output)
}

func registerV2Api() {
	http.HandleFunc("/v2/send_sms", HandleSendSMSV2)
	http.HandleFunc("/v2/send_command", HandleSendCommandV2)
}
