package exporter

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Exporter) getSpotPricing(region string, scrapes chan<- scrapeResult) {
	ec2Svc := ec2.NewFromConfig(e.awsCfg)
	pag := ec2.NewDescribeSpotPriceHistoryPaginator(
		ec2Svc,
		&ec2.DescribeSpotPriceHistoryInput{
			StartTime:           aws.Time(time.Now()),
			ProductDescriptions: e.productDescriptions,
		})
	for pag.HasMorePages() {
		history, err := pag.NextPage(context.TODO())
		if err != nil {
			log.WithError(err).Errorf("error while fetching spot price history [region=%s]", region)
			atomic.AddUint64(&e.errorCount, 1)
			break
		}
		for _, price := range history.SpotPriceHistory {
			value, err := strconv.ParseFloat(*price.SpotPrice, 64)
			if err != nil {
				log.WithError(err).Errorf("error while parsing spot price value from API response [region=%s, az=%s, type=%s]", region, *price.AvailabilityZone, price.InstanceType)
				atomic.AddUint64(&e.errorCount, 1)
			}
			log.Debugf("Creating new metric: current_price{region=%s, az=%s, instance_type=%s, product_description=%s} = %v.", region, *price.AvailabilityZone, price.InstanceType, price.ProductDescription, value)
			scrapes <- scrapeResult{
				Name:               "current_price",
				Value:              value,
				Region:             region,
				AvailabilityZone:   *price.AvailabilityZone,
				InstanceType:       string(price.InstanceType),
				InstanceLifecycle:  "spot",
				ProductDescription: string(price.ProductDescription),
			}
		}
	}
}
