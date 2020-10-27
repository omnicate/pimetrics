package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"

	modem "pimetrics/pkg/pi-modem"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/warthog618/modem/gsm"

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

	DEFAULT_CONFIG_PATH = "/home/ubuntu/config.yaml"
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

	gModem *gsm.GSM

	configPath string
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

func readFlags() {
	flag.StringVar(&configPath, "config-path", DEFAULT_CONFIG_PATH, "path to pimetrics config file")
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
	readFlags()
	registerMetrics()

	readConfig(configPath)

	var err error
	gModem, err = modem.InitModemV2(&CurrentConfig.AppConfig.ModemConfig, []gsm.Option{})
	if err != nil {
		log.WithError(err).Error("Failed initialising modem with")
	} else {
		isModemInitialised.Inc()
		log.Info("Successfully initialised modem")
	}
	// if !modem.ModemInitialized() {
	// 	if err := modem.InitModem(); err != nil {
	// 		log.WithError(err).Error("Failed initialising modem with")
	// 	} else {
	// 		isModemInitialised.Inc()
	// 		log.Info("Successfully initialised modem")
	// 	}
	// } else {
	// 	isModemInitialised.Inc()
	// }

	//setUpStatus()
}

func main() {
	log.Infoln(HEADER)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.Handle(METRICS_ENDPOINT, prom.Handler())

	registerApiV1()
	registerV2Api()

	log.WithFields(log.Fields{
		"port": CurrentConfig.AppConfig.Port,
	}).Info("Listening on ")

	http.ListenAndServe(fmt.Sprintf(":%d", CurrentConfig.AppConfig.Port), nil)
}
