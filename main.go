package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

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
	isUpMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pimetrics_is_up",
		Help: "Is pimetrics system is up",
	})

	isModemInitialised = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pimetrics_is_modem_initialised",
		Help: "Is the pi modem initialised successfully",
	})
)

func HandleSendCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST requests are valid", http.StatusBadRequest)
		return
	}

	// var cmd string

	// err := json.NewDecoder(r.Body).Decode(&cmd)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	cmd, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed getting command from body with %v", err),
			http.StatusBadRequest)
		return
	}

	output, err := modem.SendCommand(string(cmd))
	if err != nil {
		http.Error(w, fmt.Sprintf("SendCommand failed with %v", err),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, output)
}

func main() {
	log.Infoln(HEADER)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.Handle(METRICS_ENDPOINT, prom.Handler())
	http.HandleFunc("/send_command", HandleSendCommand)

	isUpMetric.Inc()

	if err := modem.InitModem(); err != nil {
		log.WithError(err).Error("Failed initialising modem with")
	} else {
		isModemInitialised.Inc()
		log.Info("Successfully initialised modem")
	}
	log.Info("Listening on" + string(HTTP_PORT))
	http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), nil)
}
