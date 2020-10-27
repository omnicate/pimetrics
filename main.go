package main

import (
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	modem "pimetrics/pkg/pi-modem"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/warthog618/modem/gsm"

	prom "github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

const (
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

type RenderData struct {
	IP   string
	Port uint
}

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

	gModem *modem.PiModem

	configPath string

	renderData RenderData
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

	flag.Parse()
}

func init() {
	readFlags()
	registerMetrics()

	readConfig(configPath)

	renderData.IP = getOwnIP()
	renderData.Port = CurrentConfig.AppConfig.Port

	var err error
	gModem, err = modem.InitModemV2(&CurrentConfig.AppConfig.ModemConfig,
		[]gsm.Option{})
	if err != nil {
		log.WithError(err).Error("Failed initialising modem with")
	} else {
		isModemInitialised.Inc()
		log.Info("Successfully initialised modem")
	}
}

func main() {
	log.Infoln(HEADER)

	fs := http.FileServer(http.Dir("./web/static"))

	mux := mux.NewRouter()
	mux.HandleFunc("/", handleIndex).Methods("GET")
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs)).Methods("GET")
	mux.Handle(METRICS_ENDPOINT, prom.Handler()).Methods("GET")
	mux.HandleFunc("/v2/send_sms", HandleSendSMSV2).Methods("POST")
	mux.HandleFunc("/v2/send_command", HandleSendCommandV2).Methods("POST")
	mux.HandleFunc("/v2/call", HandleCall).Methods("POST")
	mux.HandleFunc("/v2/sms_receive", HandleSmsRecieveMode).Methods("POST")
	mux.HandleFunc("/v2/stop_sms_receive", HandleSmsStopRecieveMode).Methods("POST")

	webServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", CurrentConfig.AppConfig.Port),
		Handler: mux,
	}

	log.Info("Starting web Server now on " + renderData.IP + ":" + strconv.Itoa(int(renderData.Port)))
	err := webServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error(err, "Could not start web server")
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := template.Must(template.ParseFiles("web/index.gohtml")).Execute(w, renderData)
	if err != nil {
		log.Error(err)
	}
}

func getOwnIP() string {
	ifaces, _ := net.Interfaces()

	for _, i := range ifaces {
		addrs, _ := i.Addrs()

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}
