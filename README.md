## AWS EC2 price exporter

Prometheus exporter for AWS EC2 prices.
The exporter is fetching the current spot and ondemand prices from the AWS API when scraped from the Prometheus server, with an option to cache results.
Price info is queried in every available region for every available instance type and exposed via an HTTP metrics endpoint.

### Quick start

Building the project is as simple as running a go build command. The result is a statically linked executable binary.
```
make build
```

#### Docker

An auto-built image is available at https://hub.docker.com/r/AndreZiviani/ec2-price-exporter/

```
docker run -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY AndreZiviani/ec2-price-exporter
```

### Run as dev
```
export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""

make run-dev
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
  -product-descriptions string
        Comma separated list of product descriptions, used to filter spot instances. Accepted values: Linux/UNIX, SUSE Linux, Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC), Windows (Amazon VPC) (default "Linux/UNIX")
  -operating-systems string
        Comma separated list of operating systems, used to filter ondemand instances. Accepted values: Linux, RHEL, SUSE, Windows (default "Linux")
  -regions string
        Comma separated list of AWS regions to get pricing for (defaults to *all*)
  -cache int
        How long should the results be cached, in seconds (defaults to *0*)
  -lifecycle string
        Comma separated list of Lifecycles (spot or ondemand) to get pricing for (defaults to *all*)
```

### Example metrics

```
[...]
ec2_current_price{availability_zone="us-west-2b",instance_lifecycle="spot",instance_type="c5.xlarge",operating_system="",product_description="Linux/UNIX",region="us-west-2"} 0.0735
ec2_current_price{availability_zone="us-west-2b",instance_lifecycle="ondemand",instance_type="c5.xlarge",operating_system="Linux",product_description="",region="us-west-2"} 0.1700
ec2_current_price{availability_zone="us-west-2b",instance_lifecycle="spot",instance_type="c5.xlarge",operating_system="",product_description="Windows",region="us-west-2"} 0.2484
ec2_current_price{availability_zone="us-west-2b",instance_lifecycle="ondemand",instance_type="c5.xlarge",operating_system="Windows",product_description="",region="us-west-2"} 0.3540
[...]
ec2_current_price{availability_zone="us-west-2c",instance_lifecycle="spot",instance_type="c5.xlarge",operating_system="",product_description="Linux/UNIX",region="us-west-2"} 0.0604
ec2_current_price{availability_zone="us-west-2c",instance_lifecycle="ondemand",instance_type="c5.xlarge",operating_system="Linux",product_description="",region="us-west-2"} 0.1700
ec2_current_price{availability_zone="us-west-2c",instance_lifecycle="spot",instance_type="c5.xlarge",operating_system="",product_description="Windows",region="us-west-2"} 0.2442
ec2_current_price{availability_zone="us-west-2c",instance_lifecycle="ondemand",instance_type="c5.xlarge",operating_system="Windows",product_description="",region="us-west-2"} 0.3540
[...]
ec2_current_price{availability_zone="eu-west-1c",instance_lifecycle="spot",instance_type="c5.xlarge",operating_system="",product_description="Linux/UNIX",region="eu-west-1"} 0.0754
ec2_current_price{availability_zone="eu-west-1c",instance_lifecycle="ondemand",instance_type="c5.xlarge",operating_system="Linux",product_description="",region="eu-west-1"} 0.1920
ec2_current_price{availability_zone="eu-west-1c",instance_lifecycle="spot",instance_type="c5.xlarge",operating_system="",product_description="Windows",region="eu-west-1"} 0.2482
ec2_current_price{availability_zone="eu-west-1c",instance_lifecycle="ondemand",instance_type="c5.xlarge",operating_system="Windows",product_description="",region="eu-west-1"} 0.3760
[...]
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="spot",instance_type="i3.4xlarge",operating_system="",product_description="Linux/UNIX",region="ap-southeast-1"} 0.4488
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="ondemand",instance_type="i3.4xlarge",operating_system="Linux",product_description="",region="ap-southeast-1"} 1.4960
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="spot",instance_type="i3.4xlarge",operating_system="",product_description="Windows",region="ap-southeast-1"} 1.1848
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="ondemand",instance_type="i3.4xlarge",operating_system="Windows",product_description="",region="ap-southeast-1"} 2.2320
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="spot",instance_type="i3.8xlarge",operating_system="",product_description="Linux/UNIX",region="ap-southeast-1"} 1.1184
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="ondemand",instance_type="i3.8xlarge",operating_system="Linux",product_description="",region="ap-southeast-1"} 2.9920
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="spot",instance_type="i3.8xlarge",operating_system="",product_description="Windows",region="ap-southeast-1"} 2.3696
ec2_current_price{availability_zone="ap-southeast-1a",instance_lifecycle="ondemand",instance_type="i3.8xlarge",operating_system="Windows",product_description="",region="ap-southeast-1"} 4.4640
[...]
```
