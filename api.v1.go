package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	modem "pimetrics/pkg/pi-modem"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func HandleSendCommand(w http.ResponseWriter, r *http.Request) {
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

	output, err := modem.SendCommand(string(cmd) + modem.BREAKLINE + modem.CTRL_Z)
	if err != nil {
		log.WithError(err).Error("SendCommand failed")
		http.Error(w, fmt.Sprintf("SendCommand failed with %v", err),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, output)
}

func HandleSendSMS(w http.ResponseWriter, r *http.Request) {
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

	output, err := modem.SendSMS(sms)
	if err != nil {
		log.WithError(err).Error("SendSMS failed")
		http.Error(w, fmt.Sprintf("SendSMS failed with %v", err),
			http.StatusInternalServerError)
		SMSSentError.With(prometheus.Labels{"number": sms.Number}).Inc()
		return
	}

	SMSSentSuccess.With(prometheus.Labels{"number": sms.Number}).Inc()

	fmt.Fprint(w, output)
}

func HandleMakeCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST requests are valid", http.StatusBadRequest)
		return
	}

	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	var call modem.Call

	err := json.NewDecoder(r.Body).Decode(&call)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"number": call.Number,
		"input":  call.Input,
	}).Info("Making Call with the following details")

	output, err := modem.MakeCall(call)
	if err != nil {
		log.WithError(err).Error("MakeCall failed")
		http.Error(w, fmt.Sprintf("MakeCall failed with %v", err),
			http.StatusInternalServerError)
		CallError.With(prometheus.Labels{"number": call.Number}).Inc()
		return
	}

	CallSuccess.With(prometheus.Labels{"number": call.Number}).Inc()

	fmt.Fprint(w, output)
}

func registerApiV1() {
	http.HandleFunc("/send_command", HandleSendCommand)
	http.HandleFunc("/send_sms", HandleSendSMS)
	http.HandleFunc("/make_call", HandleMakeCall)
}
