package exporter

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Exporter implements the prometheus.Exporter interface, and exports AWS Spot Price metrics.
type Exporter struct {
	session             *session.Session
	partitions          []string
	productDescriptions []string
	duration            prometheus.Gauge
	scrapeErrors        prometheus.Gauge
	totalScrapes        prometheus.Counter
	spotMetrics         map[string]*prometheus.GaugeVec
	metricsMtx          sync.RWMutex
	sync.RWMutex
}

type scrapeResult struct {
	Name               string
	Value              float64
	Region             string
	AvailabilityZone   string
	InstanceType       string
	ProductDescription string
}

// NewExporter returns a new exporter of AWS Autoscaling group metrics.
func NewExporter(p []string, pds []string) (*Exporter, error) {

	session, err := session.NewSession()
	if err != nil {
		log.WithError(err).Error("Error creating AWS session")
		return nil, err
	}

	e := Exporter{
		session:             session,
		partitions:          p,
		productDescriptions: pds,
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
	}, []string{"instance_type", "region", "availability_zone", "product_description"})
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

	spotScrapes := make(chan scrapeResult)

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

func (e *Exporter) scrape(scrapes chan<- scrapeResult) {

	defer close(scrapes)
	now := time.Now().UnixNano()
	e.totalScrapes.Inc()

	var errorCount uint64

	dp := endpoints.DefaultPartitions()
	for _, p := range dp {
		if !e.inPartitions(p.ID()) {
			continue
		}
		log.Infof("querying spot prices in all regions [partition=%s]", p.ID())
		var wg sync.WaitGroup
		for _, r := range p.Regions() {
			log.Debugf("querying spot prices [region=%s]", r.ID())
			wg.Add(1)
			go func(r string) {
				defer wg.Done()
				ec2Svc := ec2.New(e.session, &aws.Config{Region: aws.String(r)})
				var pds []*string
				if len(e.productDescriptions) > 0 {
					pds = aws.StringSlice(e.productDescriptions)
				} else {
					pds = nil
				}
				err := ec2Svc.DescribeSpotPriceHistoryPages(&ec2.DescribeSpotPriceHistoryInput{
					StartTime:           aws.Time(time.Now()),
					ProductDescriptions: pds,
				}, func(history *ec2.DescribeSpotPriceHistoryOutput, lastPage bool) bool {
					for _, pe := range history.SpotPriceHistory {
						price, err := strconv.ParseFloat(*pe.SpotPrice, 64)
						if err != nil {
							log.WithError(err).Errorf("error while parsing spot price value from API response [region=%s, az=%s, type=%s]", r, *pe.AvailabilityZone, *pe.InstanceType)
							atomic.AddUint64(&errorCount, 1)
						}
						log.Debugf("Creating new metric: current_price{region=%s, az=%s, instance_type=%s, product_description=%s} = %v.", r, *pe.AvailabilityZone, *pe.InstanceType, *pe.ProductDescription, price)
						scrapes <- scrapeResult{
							Name:               "current_price",
							Value:              price,
							Region:             r,
							AvailabilityZone:   *pe.AvailabilityZone,
							InstanceType:       *pe.InstanceType,
							ProductDescription: *pe.ProductDescription,
						}

					}
					return true
				})
				if err != nil {
					log.WithError(err).Errorf("error while fetching spot price history [region=%s]", r)
					atomic.AddUint64(&errorCount, 1)
				}
			}(r.ID())
		}
		wg.Wait()
	}

	e.scrapeErrors.Set(float64(atomic.LoadUint64(&errorCount)))
	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)
}

func (e *Exporter) setSpotMetrics(scrapes <-chan scrapeResult) {
	log.Debug("set spot metrics")
	for scr := range scrapes {
		name := scr.Name
		if _, ok := e.spotMetrics[name]; !ok {
			e.metricsMtx.Lock()
			e.spotMetrics[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: "aws_spot",
				Name:      name,
			}, []string{"instance_type", "region", "availability_zone", "product_description"})
			e.metricsMtx.Unlock()
		}
		var labels prometheus.Labels = map[string]string{
			"instance_type":       scr.InstanceType,
			"region":              scr.Region,
			"availability_zone":   scr.AvailabilityZone,
			"product_description": scr.ProductDescription,
		}
		e.spotMetrics[name].With(labels).Set(float64(scr.Value))
	}
}

func (e *Exporter) inPartitions(p string) bool {
	for _, partition := range e.partitions {
		if p == partition {
			return true
		}
	}
	return false
}
