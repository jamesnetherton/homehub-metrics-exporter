# Home Hub Metrics Exporter

[![CircleCI](https://img.shields.io/circleci/project/github/jamesnetherton/homehub-metrics-exporter/master.svg)](https://circleci.com/gh/jamesnetherton/homehub-metrics-exporter/tree/master)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=600)](https://opensource.org/licenses/MIT)

A [Prometheus](https://prometheus.io) exporter for BT Home Hub routers.

## Running

```
./homehub-metrics-exporter --listen-address=0.0.0.0:19092 --hub-address=192.168.1.254 --hub-username=admin --hub-password=secret
```

Some of the arguments can be ommitted if the default values are acceptable.

| Name           | Default Value   |
|----------------|-----------------|
| listen-address | 0.0.0.0:19092 |
| hub-address    | 192.168.1.254   |
| hub-username   | admin           |

With the exporter running, hit the /metrics endpoint to collect metrics from the Home Hub. Here's a breakdown of available metrics.

| Metrics Name           | Description   |
|----------------|-----------------|
| bt_homehub_build_info | Home Hub build information. Currently only a label for the firmware version. |
| bt_homehub_device_downloaded_megabytes | Total megabytes downloaded by each active device. |
| bt_homehub_device_uploaded_megabytes | Total megabytes uploaded by each active device. |
| bt_homehub_download_rate_mbps | The download rate of the Home Hub router. |
| bt_homehub_upload_rate_mbps | The upload rate of the Home Hub router |
| bt_homehub_up | Whether the Home Hub is 'Up'. Will be 0 if the exporter failed to collect metrics. |
| bt_homehub_uptime_seconds | The amount of time in seconds that the Home Hub has been running. |

## Docker image

You can run the exporter within a Docker container:

```
docker run -ti --rm jamesnetherton/homehub-exporter \ 
    --listen-address=0.0.0.0:19092 \
    --hub-address=192.168.1.254 \
    --hub-username=admin \
    --hub-password=secret
```

## Docker Compose

Getting started is simple with [Docker Compose](https://docs.docker.com/compose/).

You'll need to edit the [docker-compose.yml](docker-compose.yml) file to add your Home Hub password to the `--hub-password` argument. Then Simply run `docker-compose up`.

The Home Hub Grafana dashboard can be accessed at http://localhost:3000. Prometheus is available at http://localhost:9090.

## Building

    git clone git@github.com:jamesnetherton/homehub-metrics-exporter.git $GOPATH/src/homehub-metrics-exporter
    make build

Generated binaries are output to the `build` directory.
