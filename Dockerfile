FROM golang:1-alpine3.15 as backend

COPY . /src/spot-price-exporter
WORKDIR /src/spot-price-exporter

RUN go build -o /bin/spot-price-exporter -ldflags="-s -w" .



FROM alpine:3.15
COPY --from=backend /bin/spot-price-exporter /bin
ENTRYPOINT ["/bin/spot-price-exporter"]
