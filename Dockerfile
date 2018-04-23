# build stage
FROM golang:1.9.3-alpine3.7

ADD . /go/src/github.com/banzaicloud/spot-price-exporter
WORKDIR /go/src/github.com/banzaicloud/spot-price-exporter
RUN go build -o /bin/spot-price-exporter .

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=0 /bin/spot-price-exporter /bin
ENTRYPOINT ["/bin/spot-price-exporter"]
