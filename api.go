package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	modem "pimetrics/pkg/pi-modem"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/modem/at"
	"github.com/warthog618/modem/gsm"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

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

func HandleSmsReceiveMode(w http.ResponseWriter, r *http.Request) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	err := gModem.StartMessageRx(
		func(msg gsm.Message) {
			log.WithField("message", msg.Message).Infof("Recieved SMS from %s", msg.Number)
			for client := range wsClients {
				err := client.WriteJSON(modem.SMS{
					Text:   msg.Message,
					Number: msg.Number,
				})
				if err != nil {
					log.WithError(err).Error("Failed sending sms message to web socket")
					client.Close()
					delete(wsClients, client)
				}
			}
		},
		func(err error) {
			log.WithError(err).Error("Failed reciving sms")
		})
	if err != nil {
		log.WithError(err).Error("StartMessageRx failed")
		http.Error(w, fmt.Sprintf("StartMessageRx failed %v", err), http.StatusInternalServerError)
		return
	}

	log.Info("Waiting for SMS")
	fmt.Fprint(w, "Waiting for SMS")
}

func HandleSmsStopReceiveMode(w http.ResponseWriter, r *http.Request) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	gModem.StopMessageRx()

	log.Info("Stopped waiting for SMS")
	fmt.Fprint(w, "Stopped waiting for SMS")
}

func handleSignalStatus(w http.ResponseWriter, r *http.Request) {
	i, err := gModem.Command("+CSQ")
	if err != nil {
		log.WithField("signal_status", err)
	} else {
		squal := strings.Split(i[0], " ")
		if len(squal) == 2 {
			fQual, _ := strconv.ParseFloat(strings.Replace(squal[1], ",", ".", 1), 32)
			rQual := signalQualityReadable(int(fQual))
			w.Write([]byte(fmt.Sprintf("%d (%s)", int(fQual), rQual)))
		}
	}
}

func handleGetProvider(w http.ResponseWriter, r *http.Request) {
	i, err := gModem.Command("+COPS?")
	if err != nil {
		log.WithField("provider", err)
	} else {
		retCmd := strings.Split(i[0], ",")
		if len(retCmd) >= 2 {
			pString := strings.Split(i[0], ",")[2]
			w.Write([]byte(pString))
		} else {
			w.Write([]byte("Unknown"))
		}

	}
}

func handleTestRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var tr TestResult
	tr.Message = "Great SUCCESS!!!!"

	if vars["target"] == "" {
		tr.Error = errors.New("invalid request")
		tr.Message = "Could not parse target"
	}

	if tconfig, ok := LabConfig[vars["target"]]; ok {
		//Target config has been found, do something with it
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		// Send SMS with magic content
		_, err := gModem.SendShortMessage(tconfig.Msisdn, "OMG_MAGIC_STUFF_!!_one_eleven_!!!")
		if err != nil {
			tr.Error = errors.New("failed in SMS send operation")

			tr.Operations = append(tr.Operations, TestOperation{
				Type:    "SMSSend",
				Success: false,
				Error:   err,
			})
		} else {
			tr.Operations = append(tr.Operations, TestOperation{
				Type:    "SMSSend",
				Success: true,
				Error:   nil,
			})
		}

		// Reset receive mode because we need different MessageHandler functions
		gModem.StopMessageRx()

		// Now Wait for SMS returning
		_ = gModem.StartMessageRx(
			func(msg gsm.Message) {
				log.WithField("message", msg.Message).Infof("Received SMS from %s", msg.Number)
				if msg.Message == "NO_MAGIC_JUST_TURTLES" {
					tr.Operations = append(tr.Operations, TestOperation{
						Type:    "SMSReceive",
						Success: true,
						Error:   nil,
					})
				} else {
					tr.Operations = append(tr.Operations, TestOperation{
						Type:    "SMSReceive",
						Success: false,
						Error:   errors.New("no return SMS received"),
					})
				}
			},
			func(err error) {
				log.WithError(err).Error("Failed receiving sms")
			})

		// Sleep for 5 seconds to limit how long we are listening on SMS
		time.Sleep(5 * time.Duration(time.Second))

		gModem.StopMessageRx()

		// Start standard receive mode again
		gModem.ReceiveMode()

		// if no messages have been received the length of TestOperations will be less than 2, write error
		if len(tr.Operations) < 2 {
			tr.Operations = append(tr.Operations, TestOperation{
				Type:    "SMSReceive",
				Success: false,
				Error:   errors.New("no return SMS received"),
			})
		}

	} else {
		// We dont know that target, return error
		tr.Error = errors.New("unknown target")
		tr.Message = "Could not find target"
	}

	w.Header().Add("Content-Type", "application/json")
	trj, _ := json.Marshal(tr)
	w.Write(trj)
}

type TestResult struct {
	Message    string          `json:"Message"`
	Error      error           `json:"Error"`
	Operations []TestOperation `json:"Operations"`
}

type TestOperation struct {
	Type    string `json:"Type"`
	Success bool   `json:"Success"`
	Error   error  `json:"Error"`
}

func signalQualityReadable(iQual int) string {
	if iQual >= 2 && iQual <= 9 {
		return "Marginal"
	}
	if iQual >= 10 && iQual <= 14 {
		return "OK"
	}
	if iQual >= 15 && iQual <= 19 {
		return "Good"
	}
	if iQual >= 20 && iQual <= 30 {
		return "Excellent"
	}
	return "Unknown"
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("WebSocket Upgrade failed")
		http.Error(w, fmt.Sprintf("WebSocket Upgrade failed %v", err), http.StatusInternalServerError)
		return
	}

	wsClients[ws] = true
	log.WithField("Address", ws.RemoteAddr().String()).Info("Registered WebSocket client")
}
