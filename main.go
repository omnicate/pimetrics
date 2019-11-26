package main

import (
	"fmt"
	"log"
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus/promhttp"
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

func main() {
	log.Println(HEADER)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.Handle(METRICS_ENDPOINT, prom.Handler())

	isUpMetric.Inc()
	log.Printf("Listening on %d...\n", HTTP_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), nil)
}
