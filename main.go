package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	HTTP_PORT    = 8080
	METRICS_PORT = 8081
)

func main() {
	fmt.Println("Pi imports")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Listening on %d...", HTTP_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), nil)
}
