FROM golang:1.9.3-alpine3.7 as backend
RUN apk update && apk add ca-certificates curl git make tzdata

RUN mkdir -p /go/src/github.com/banzaicloud/spot-price-exporter
ADD Gopkg.* Makefile /go/src/github.com/banzaicloud/spot-price-exporter/
WORKDIR /go/src/github.com/banzaicloud/spot-price-exporter

RUN make vendor
ADD . /go/src/github.com/banzaicloud/spot-price-exporter

RUN go build -o /bin/spot-price-exporter .


FROM alpine:3.7
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /bin/spot-price-exporter /bin
ENTRYPOINT ["/bin/spot-price-exporter"]
