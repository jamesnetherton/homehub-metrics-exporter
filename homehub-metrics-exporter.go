package main

import (
	"flag"
	"homehub-metrics-exporter/pkg/client"
	"homehub-metrics-exporter/pkg/exporter"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		listenAddress string
		hubAddress    string
		username      string
		password      string
	)

	flag.StringVar(&listenAddress, "listen-address", envOrDefault("HUB_EXPORTER_LISTEN_ADDRESS", ":19092"), "Address that the metrics HTTP server will listen on")
	flag.StringVar(&hubAddress, "hub-address", envOrDefault("HUB_ADDRESS", "192.168.1.254"), "Address for the Home Hub router")
	flag.StringVar(&username, "hub-username", envOrDefault("HUB_USERNAME", "admin"), "Username for the Home Hub router")
	flag.StringVar(&password, "hub-password", envOrDefault("HUB_PASSWORD", ""), "Password for the Home Hub router")
	flag.Parse()

	homehub := client.New("http://"+hubAddress, username, password)
	response := homehub.Login()

	if response.Error != nil {
		log.Fatalln("Home Hub login failed. Unable to collect metrics.")
	}

	exporter := exporter.New(homehub)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Home Hub Exporter")

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		                <head><title>Home Hub Exporter</title></head>
		                <body>
		                   <h1>Home Hub Exporter</h1>
		                   <p><a href="/metrics">Metrics</a></p>
		                   </body>
		                </html>
		              `))
	})
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func envOrDefault(env string, defaultValue string) string {
	if value, present := os.LookupEnv(env); present == true {
		return value
	}
	return defaultValue
}
