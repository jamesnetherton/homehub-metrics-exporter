FROM golang:alpine as compile

RUN apk --no-cache add make git

# Download go modules before copying in the source code, to improve caching
COPY go.* src/homehub-metrics-exporter/
RUN cd src/homehub-metrics-exporter && go mod download

COPY . src/homehub-metrics-exporter

ENV CGO_ENABLED=0
RUN cd src/homehub-metrics-exporter && \
    make build && \
    cp build/homehub-metrics-exporter /go/bin/homehub-metrics-exporter

FROM scratch

COPY --from=compile /go/bin/homehub-metrics-exporter /homehub-metrics-exporter

ENTRYPOINT ["/homehub-metrics-exporter"]
