package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	modem "pimetrics/pkg/pi-modem"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/modem/at"
	"github.com/warthog618/modem/gsm"
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
	output, err := gModem.Command(string(cmd)+modem.BREAKLINE+modem.CTRL_Z, []at.CommandOption{}...)
	if err != nil {
		log.WithError(err).Error("SendCommand failed")
		http.Error(w, fmt.Sprintf("SendCommand failed with %v", err),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, output)
}

func HandleCall(w http.ResponseWriter, r *http.Request) {
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
	}).Info("Calling the following number")

	rsp, err := gModem.Call(call.Number)
	if err != nil {
		log.WithError(err).Error("Call failed")
		http.Error(w, fmt.Sprintf("Call failed with %v", err), http.StatusInternalServerError)
		return
	}

	// Go routine to hangup the call
	go func() {
		for {
			select {
			case <-time.After(time.Second * 25):
				r, err := gModem.Handup()
				if err != nil {
					log.WithError(err).Error("Failed hanging up call")
				}
				log.WithField("response", r).Infof("Hanged up call with %s", call.Number)
				return
			}
		}
	}()

	fmt.Fprint(w, rsp)
}

func HandleSmsRecieveMode(w http.ResponseWriter, r *http.Request) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	err := gModem.StartMessageRx(
		func(msg gsm.Message) {
			log.WithField("message", msg.Message).Infof("Recieved SMS from %s", msg.Number)
		},
		func(err error) {
			log.WithError(err).Error("Failed reciving sms")
		})
	if err != nil {
		log.WithError(err).Error("StartMessageRx failed")
	}

	log.Info("Waiting for SMS")
	fmt.Fprint(w, "Waiting for SMS")
}

func HandleSmsStopRecieveMode(w http.ResponseWriter, r *http.Request) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	gModem.StopMessageRx()

	log.Info("Stopped waiting for SMS")
	fmt.Fprint(w, "Stopped waiting for SMS")
}
