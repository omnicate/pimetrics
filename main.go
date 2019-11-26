package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	modem "pimetrics/pkg/pi-modem"

	prom "github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	HTTP_PORT        = 8080
	METRICS_ENDPOINT = "/metrics"

	HEADER = `
	_______ _________   _______  _______ _________ _______ _________ _______  _______
   (  ____ )\__   __/  (       )(  ____ \\__   __/(  ____ )\__   __/(  ____ \(  ____ \
   | (    )|   ) (     | () () || (    \/   ) (   | (    )|   ) (   | (    \/| (    \/
   | (____)|   | |     | || || || (__       | |   | (____)|   | |   | |      | (_____
   |  _____)   | |     | |(_)| ||  __)      | |   |     __)   | |   | |      (_____  )
   | (         | |     | |   | || (         | |   | (\ (      | |   | |            ) |
   | )      ___) (___  | )   ( || (____/\   | |   | ) \ \_____) (___| (____/\/\____) |
   |/       \_______/  |/     \|(_______/   )_(   |/   \__/\_______/(_______/\_______)

`
)

var (
	isUpMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pimetrics_is_up",
		Help: "Is pimetrics system is up",
	}, []string{"roaming", "network"})

	isModemInitialised = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pimetrics_is_modem_initialised",
		Help: "Is the pi modem initialised successfully",
	})
	SMSSentSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pimetrics_sms_sent_success",
		Help: "SMS successfully sent from number",
	}, []string{"number"})
	SMSSentError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pimetrics_sms_sent_error",
		Help: "SMS successfully sent from number",
	}, []string{"number"})
	CallSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pimetrics_call_success",
		Help: "Call has been established",
	}, []string{"number"})
	CallError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pimetrics_call_error",
		Help: "Call has failed",
	}, []string{"number"})

	handlerMutex = &sync.Mutex{}
)

func registerMetrics() {
	if err := prometheus.Register(isUpMetric); err != nil {
		log.WithError(err).Error("Failed register isUpMetric")
	}
	if err := prometheus.Register(isModemInitialised); err != nil {
		log.WithError(err).Error("Failed register isModemInitialised")
	}
	if err := prometheus.Register(SMSSentSuccess); err != nil {
		log.WithError(err).Error("Failed register SMSSentSuccess")
	}
	if err := prometheus.Register(SMSSentError); err != nil {
		log.WithError(err).Error("Failed register SMSSentError")
	}
	if err := prometheus.Register(CallSuccess); err != nil {
		log.WithError(err).Error("Failed register CallSuccess")
	}
	if err := prometheus.Register(CallError); err != nil {
		log.WithError(err).Error("Failed register CallError")
	}
}

func setUpStatus() {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	roamingCmd := "AT+CGREG?" + modem.BREAKLINE
	networkCmd := "AT+COPS?" + modem.BREAKLINE

	isRoaming := "0"
	network := ""

	output, err := modem.SendCommand(roamingCmd)
	if err != nil {
		log.WithError(err).Errorf("SendCommand %s failed", roamingCmd)
	} else {
		log.WithFields(log.Fields{
			"command": roamingCmd,
			"output":  output,
		}).Info("Roaming command output")
		if strings.Contains(output, "0,5") {
			isRoaming = "1"
		}
	}

	output, err = modem.SendCommand(networkCmd)
	if err != nil {
		log.WithError(err).Errorf("SendCommand %s failed", networkCmd)
	} else {
		log.WithFields(log.Fields{
			"command": networkCmd,
			"output":  output,
		}).Info("Roaming command output")

		// example output: "T+COPS?\r\r\n+COPS: 0,0,\"N Telenor LOLTEL\",2\r\n\r\nOK\r\n"
		network = strings.Split(output, "\"")[1]
	}

	isUpMetric.With(prometheus.Labels{
		"roaming": isRoaming,
		"network": network}).Inc()
}

func init() {
	registerMetrics()

	if !modem.ModemInitialized() {
		if err := modem.InitModem(); err != nil {
			log.WithError(err).Error("Failed initialising modem with")
		} else {
			isModemInitialised.Inc()
			log.Info("Successfully initialised modem")
		}
	} else {
		isModemInitialised.Inc()
	}

	setUpStatus()
}

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

	var call modem.Call

	err := json.NewDecoder(r.Body).Decode(&call)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"number": call.Number,
		"input":   call.Input,
	}).Info("Making Call with the following details")

	output, err := modem.MakeCall(call)
	if err != nil {
		log.WithError(err).Error("MakeCall failed")
		http.Error(w, fmt.Sprintf("MakeCall failed with %v", err),
			http.StatusInternalServerError)
		CallError.With(prometheus.Labels{"number":call.Number}).Inc()
		return
	}

	CallSuccess.With(prometheus.Labels{"number":call.Number}).Inc()

	fmt.Fprint(w,output)
}

func main() {
	log.Infoln(HEADER)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.Handle(METRICS_ENDPOINT, prom.Handler())
	http.HandleFunc("/send_command", HandleSendCommand)
	http.HandleFunc("/send_sms", HandleSendSMS)
	http.HandleFunc("/make_call", HandleMakeCall)

	log.WithFields(log.Fields{
		"port": HTTP_PORT,
	}).Info("Listening on ")

	http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), nil)
}
