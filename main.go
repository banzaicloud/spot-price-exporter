package main

import (
	"context"
	"flag"
	"net/http"
	"strings"

	"github.com/AndreZiviani/ec2-price-exporter/exporter"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	addr                = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsPath         = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
	rawLevel            = flag.String("log-level", "info", "log level")
	productDescriptions = flag.String("product-descriptions", "Linux/UNIX", "Comma separated list of product descriptions, used to filter spot instances. Accepted values: Linux/UNIX, SUSE Linux, Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC), Windows (Amazon VPC)")
	operatingSystems    = flag.String("operating-systems", "Linux", "Comma separated list of operating systems, used to filter ondemand instances. Accepted values: Linux, RHEL, SUSE, Windows")
	regions             = flag.String("regions", "", "Comma separated list of AWS regions to get pricing for (defaults to *all*)")
	lifecycle           = flag.String("lifecycle", "", "Comma separated list of Lifecycles (spot or ondemand) to get pricing for (defaults to *all*)")
	cache               = flag.Int("cache", 0, "How long should the results be cached, in seconds (defaults to *0*)")
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
	log.Infof("Starting AWS EC2 Price exporter. [log-level=%s, product-descriptions=%s, operating-systems=%s, cache=%d]", *rawLevel, *productDescriptions, *operatingSystems, *cache)

	var reg []string
	if len(*regions) == 0 {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.WithError(err).Errorf("error while initializing aws client to list available regions")
			return
		}

		ec2Svc := ec2.NewFromConfig(cfg)
		r, err := ec2Svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)})
		if err != nil {
			log.Fatal(err)
			return
		}

		for _, region := range r.Regions {
			reg = append(reg, *region.RegionName)
		}
	} else {
		reg = splitAndTrim(*regions)
	}

	pds := splitAndTrim(*productDescriptions)
	oss := splitAndTrim(*operatingSystems)
	validateProductDesc(pds)
	validateOperatingSystems(oss)
	exporter, err := exporter.NewExporter(pds, oss, reg, *cache)
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

func validateProductDesc(pds []string) {
	for _, desc := range pds {
		if desc != "Linux/UNIX" && desc != "Linux/UNIX (Amazon VPC)" &&
			desc != "SUSE Linux" && desc != "SUSE Linux (Amazon VPC)" &&
			desc != "Windows" && desc != "Windows (Amazon VPC)" {
			log.Fatalf("product description '%s' is not recognized. Available product descriptions: Linux/UNIX, SUSE Linux, Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC), Windows (Amazon VPC)", desc)
		}
	}
}

func validateOperatingSystems(oss []string) {
	for _, os := range oss {
		if os != "Linux" &&
			os != "RHEL" &&
			os != "SUSE" &&
			os != "Windows" {
			log.Fatalf("Operating System '%s' is not recognized. Available operating system: Linux, RHEL, SUSE, Windows", os)
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html>
		<head><title>AWS EC2 Price Exporter</title></head>
		<body>
		<h1>AWS EC2 Price Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>
	`))

}
