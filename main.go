package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus/promhttp"
	modem "pimetrics/pkg/pi-modem"
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
)


func main() {
	log.Println(HEADER)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.Handle(METRICS_ENDPOINT, prom.Handler())

	isUpMetric.Inc()

	modem.Dummy()
	log.Printf("Listening on %d...\n", HTTP_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), nil)
}
