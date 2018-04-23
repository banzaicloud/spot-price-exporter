## AWS Spot price exporter

Prometheus exporter for AWS spot prices.
The exporter is fetching the current spot price from the AWS API on every scrape from the Prometheus server.
Price info is queried in every available region for every available instance type and exposed via an HTTP metrics endpoint.

### Quick start

Building the project is as simple as running a go build command. The result is a statically linked executable binary.
```
go build .
```

### Configuration

```
Usage of ./spot-price-exporter:
  -listen-address string
        The address to listen on for HTTP requests. (default ":8080")
  -log-level string
        log level (default "info")
  -metrics-path string
        path to metrics endpoint (default "/metrics")
  -partitions string
        Comma separated list of AWS partitions. Accepted values: aws, aws-cn, aws-us-gov (default "aws")
  -product-descriptions string
        Comma separated list of product descriptions. Accepted values: Linux/UNIX, SUSE Linux, Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC), Windows (Amazon VPC) (default "Linux/UNIX")
```

### Example metrics

```
[...]
aws_spot_current_price{availability_zone="us-west-2b",instance_type="c5.xlarge",product_description="Linux/UNIX",region="us-west-2"} 0.0735
aws_spot_current_price{availability_zone="us-west-2b",instance_type="c5.xlarge",product_description="Windows",region="us-west-2"}
[...]
aws_spot_current_price{availability_zone="us-west-2c",instance_type="c5.xlarge",product_description="Linux/UNIX",region="us-west-2"} 0.0604
aws_spot_current_price{availability_zone="us-west-2c",instance_type="c5.xlarge",product_description="Windows",region="us-west-2"} 0.2442
[...]
aws_spot_current_price{availability_zone="eu-west-1c",instance_type="c5.xlarge",product_description="Linux/UNIX",region="eu-west-1"} 0.0754
aws_spot_current_price{availability_zone="eu-west-1c",instance_type="c5.xlarge",product_description="Windows",region="eu-west-1"} 0.2482
[...]
aws_spot_current_price{availability_zone="ap-southeast-1a",instance_type="i3.4xlarge",product_description="Linux/UNIX",region="ap-southeast-1"} 0.4488
aws_spot_current_price{availability_zone="ap-southeast-1a",instance_type="i3.4xlarge",product_description="Windows",region="ap-southeast-1"} 1.1848
aws_spot_current_price{availability_zone="ap-southeast-1a",instance_type="i3.8xlarge",product_description="Linux/UNIX",region="ap-southeast-1"} 1.1184
aws_spot_current_price{availability_zone="ap-southeast-1a",instance_type="i3.8xlarge",product_description="Windows",region="ap-southeast-1"} 2.3696
[...]
```