package main

import (
	"flag"
	"net/http"

	"github.com/banzaicloud/spot-price-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	addr        = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsPath = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
	rawLevel    = flag.String("log-level", "info", "log level")
)

func init() {
	flag.Parse()
	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.WithError(err).Warnf("Couldn't parse log level, using default: %s", log.GetLevel())
	} else {
		log.SetLevel(parsedLevel)
		log.Debugf("Set log level to %s", parsedLevel)
	}
}

func main() {
	log.Info("Starting AWS Spot Price exporter")
	log.Infof("Starting metric http endpoint on %s", *addr)

	exporter, err := exporter.NewExporter()
	if err != nil {
		log.Fatal(err)
	}

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html>
		<head><title>AWS Spot Price Exporter</title></head>
		<body>
		<h1>AWS Spot Price Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>
	`))

}
