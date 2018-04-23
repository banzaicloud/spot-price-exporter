package exporter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Exporter implements the prometheus.Exporter interface, and exports AWS AutoScaling metrics.
type Exporter struct {
	session      *session.Session
	duration     prometheus.Gauge
	scrapeErrors prometheus.Gauge
	totalScrapes prometheus.Counter
	spotMetrics  map[string]*prometheus.GaugeVec
	metricsMtx   sync.RWMutex
	sync.RWMutex
}

type ScrapeResult struct {
	Name             string
	Value            float64
	Region           string
	AvailabilityZone string
	InstanceType     string
}

type scrapeError struct {
	count uint64
}

func (e *scrapeError) Error() string {
	return fmt.Sprintf("Error count: %d", e.count)
}

// NewExporter returns a new exporter of AWS Autoscaling group metrics.
func NewExporter() (*Exporter, error) {

	session, err := session.NewSession()
	if err != nil {
		log.WithError(err).Error("Error creating AWS session")
		return nil, err
	}

	e := Exporter{
		session: session,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "aws_spot",
			Name:      "scrape_duration_seconds",
			Help:      "The scrape duration.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "aws_spot",
			Name:      "scrapes_total",
			Help:      "Total AWS autoscaling group scrapes.",
		}),
		scrapeErrors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "aws_spot",
			Name:      "scrape_error",
			Help:      "The scrape error status.",
		}),
	}

	e.initGauges()
	return &e, nil
}

func (e *Exporter) initGauges() {
	e.spotMetrics = map[string]*prometheus.GaugeVec{}

	e.spotMetrics["current_spot_price"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "aws_spot",
		Name:      "current_price",
		Help:      "Current spot price of the instance type.",
	}, []string{"instance_type", "region", "availability_zone"})
}

// Describe outputs metric descriptions.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.spotMetrics {
		m.Describe(ch)
	}
	ch <- e.duration.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.scrapeErrors.Desc()
}

// Collect fetches info from the AWS API and the BanzaiCloud recommendation API
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	spotScrapes := make(chan ScrapeResult)

	e.Lock()
	defer e.Unlock()

	e.initGauges()
	go e.scrape(spotScrapes)
	e.setSpotMetrics(spotScrapes)

	e.duration.Collect(ch)
	e.totalScrapes.Collect(ch)
	e.scrapeErrors.Collect(ch)

	for _, m := range e.spotMetrics {
		m.Collect(ch)
	}
}

func (e *Exporter) scrape(scrapes chan<- ScrapeResult) {

	defer close(scrapes)
	now := time.Now().UnixNano()
	e.totalScrapes.Inc()

	var errorCount uint64 = 0

	//ec2Svc := ec2.New(e.session, &aws.Config{Region:aws.String("eu-west-1")})

	scrapes <- ScrapeResult{
		Name:             "current_price",
		Value:            0.12,
		Region:           "eu-west-1",
		AvailabilityZone: "eu-west-1a",
		InstanceType:     "m5.xlarge",
	}

	e.scrapeErrors.Set(float64(atomic.LoadUint64(&errorCount)))
	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)
}

func (e *Exporter) setSpotMetrics(scrapes <-chan ScrapeResult) {
	log.Debug("set spot metrics")
	for scr := range scrapes {
		name := scr.Name
		if _, ok := e.spotMetrics[name]; !ok {
			e.metricsMtx.Lock()
			e.spotMetrics[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: "aws_spot",
				Name:      name,
			}, []string{"instance_type", "region", "availability_zone"})
			e.metricsMtx.Unlock()
		}
		var labels prometheus.Labels = map[string]string{
			"instance_type":     scr.InstanceType,
			"region":            scr.Region,
			"availability_zone": scr.AvailabilityZone,
		}
		e.spotMetrics[name].With(labels).Set(float64(scr.Value))
	}
}