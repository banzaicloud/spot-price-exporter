package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/banzaicloud/spot-price-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	addr                = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsPath         = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
	rawLevel            = flag.String("log-level", "info", "log level")
	partitions          = flag.String("partitions", "aws", "Comma separated list of AWS partitions. Accepted values: aws, aws-cn, aws-us-gov")
	productDescriptions = flag.String("product-descriptions", "Linux/UNIX", "Comma separated list of product descriptions. Accepted values: Linux/UNIX, SUSE Linux, Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC), Windows (Amazon VPC)")
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
	log.Infof("Starting AWS Spot Price exporter. [log-level=%s, partitions=%s, product-descriptions=%s]", *rawLevel, *partitions, *productDescriptions)
	parts := splitAndTrim(*partitions)
	pds := splitAndTrim(*productDescriptions)
	validatePartitions(parts)
	validateProductDesc(pds)
	exporter, err := exporter.NewExporter(parts, pds)
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(exporter)

	log.Infof("Starting metric http endpoint [address=%s, path=%s]", *addr, *metricsPath)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
func splitAndTrim(str string) []string {
	if str == "" {
		return []string{}
	}
	parts := strings.Split(str, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func validatePartitions(parts []string) {
	for _, p := range parts {
		if p != "aws" && p != "aws-cn" && p != "aws-us-gov" {
			log.Fatalf("partition '%s' is not recognized. Available partitions: aws, aws-cn, aws-us-gov", p)
		}
	}
}

func validateProductDesc(pds []string) {
	for _, desc := range pds {
		if desc != "Linux/UNIX" && desc != "Linux/UNIX (Amazon VPC)" &&
			desc != "SUSE Linux" && desc != "SUSE Linux (Amazon VPC)" &&
			desc != "Windows" && desc != "Windows (Amazon VPC)" {
			log.Fatalf("product description '%s' is not recognized. Available product descriptions: Linux/UNIX, SUSE Linux, Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC), Windows (Amazon VPC)", desc)
		}
	}
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
