# GCP Status Exporter

An exporter to get metrics for availability status of GCP services and observe the incident impacts in GCP environments.

**Note**: This is a work in progress. Contributions are welcome.

## Approach

* GCP incidents older than 30 days are discarded to keep the metrics relevant.
* GCP service status is collected from [Google Cloud Status Dashboard's incident history](https://status.cloud.google.com/incidents.json)

## Components

* [Opentelemetry](https://github.com/open-telemetry/opentelemetry-go) library is used for creating metrics.
* [logrus](https://github.com/sirupsen/logrus) for structured logging.
* [cobra](https://github.com/spf13/cobra) for creating a CLI application.

## Metrics

List of metrics created by this exporter is listed [here](./metrics.md) along with the description.

## Building the Exporter 

```bash
$ git clone https://github.com/ishantanu/gcp-status-exporter
$ go build -o gcp-status-exporter
$ ./gcp-status-exporter start -e https://status.cloud.google.com/incidents.json
```

## Running the Exporter

```bash
$ go run main.go start -e https://status.cloud.google.com/incidents.json

```

Visit http://localhost:8888 for metrics.

## Adding gcp-status-exporter config to Prometheus

Add the below block to prometheus config to start scraping the metrics

```yaml
scrape_configs:
    - job_name: 'gcp-status-exporter'

    static_configs:
    - targets: ['<gcp-status-exporter>:8888']
```
