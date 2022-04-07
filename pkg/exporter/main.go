// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"fmt"

	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"

	"github.com/ishantanu/gcp-status-exporter/pkg/common"
	"github.com/ishantanu/gcp-status-exporter/pkg/gcpstatus"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func initMeter() {
	config := prometheus.Config{
		DefaultHistogramBoundaries: []float64{1, 2, 5, 10, 20, 50},
	}
	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	)
	exporter, err := prometheus.New(config, c)
	if err != nil {
		log.Panicf("failed to initialize prometheus exporter %v", err)
	}

	global.SetMeterProvider(exporter.MeterProvider())

	http.HandleFunc("/", exporter.ServeHTTP)
	go func() {
		_ = http.ListenAndServe(":8888", nil)
	}()

	fmt.Println("Prometheus server running on :8888")
}

func StartExporter(gcpStatusEndpoint string, metricPrefix string) {
	initMeter()

	meter := global.MeterProvider().Meter("otel/gcp_status_exporter")

	cLoggerEntry := common.SetLogging()

	if err := getGCPMetrics(gcpStatusEndpoint, cLoggerEntry, meter); err != nil {
		log.Fatal(err)
		cLoggerEntry.WithFields(log.Fields{
			"gcpStatusEndpoint": gcpStatusEndpoint,
			"metricPrefix":      metricPrefix,
		}).Errorf("getGCPMetrics: %v", err)
	}

}

func getGCPMetrics(gcpStatusEndpoint string, cLoggerEntry *log.Entry, meter metric.Meter) error {

	mGCPServiceStatus, _ := meter.AsyncInt64().Gauge("monitoring/gcp_service_status")
	mGCPIncidentBegin, _ := meter.AsyncFloat64().Gauge("monitoring/gcp_incident_begin")
	mGCPIncidentCreated, _ := meter.AsyncFloat64().Gauge("monitoring/gcp_incident_created")
	mGCPMostRecentUpdate, _ := meter.AsyncFloat64().Gauge("monitoring/gcp_most_recent_update")
	mGCPIncidentResolutionTotal, _ := meter.AsyncFloat64().Gauge("monitoring/gcp_incident_resolution_seconds_total")

	for {

		cLoggerEntry.WithFields(log.Fields{
			"gcpStatusEndpoint": gcpStatusEndpoint,
		}).Infof("Fetching GCP service statuses from: %v", gcpStatusEndpoint)

		// Record metrics every 10 seconds
		for range time.Tick(time.Second * 10) {
			statusData, err := gcpstatus.GetGcpStatus(gcpStatusEndpoint)
			if err != nil {
				cLoggerEntry.WithFields(log.Fields{
					"gcpStatusEndpoint": gcpStatusEndpoint,
				}).Errorf("Error fetching GCP status details from: %v", gcpStatusEndpoint)
			}

			cLoggerEntry.WithFields(log.Fields{
				"gcpStatusEndpoint": gcpStatusEndpoint,
			}).Info("Starting recording metrics.")

			for _, v := range statusData {

				/*
					Record current status for the service
					Conditions when status_impact is set to none:
					1. Most recent status is "AVAILABLE" since 6000 ms from current time.
					2. Most recent stauts is not "AVAILABLE" but the the last most recent update time was more than 2 days ago. This is done because in the historic data, some entries were found where the most recent status was not "AVAILABLE" but the incident was in resolved state. For now, the arbitrary value of 2 days is set. This can, however, change based on further discussion.
				*/
				if (v.MostRecentUpdate.Status == "AVAILABLE" && time.Since(v.MostRecentUpdate.When).Milliseconds() > 6000) || (v.MostRecentUpdate.Status != "AVAILABLE" && time.Now().Sub(v.MostRecentUpdate.When).Seconds() > 172800) {

					commonLabels := []attribute.KeyValue{attribute.String("gcp_service_name", v.ServiceName), attribute.String("status_impact", "none")}
					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Recording metric monitoring/gcp_service_up")

					err = meter.RegisterCallback([]instrument.Asynchronous{
						mGCPServiceStatus,
					},

						func(ctx context.Context) {
							mGCPServiceStatus.Observe(ctx, int64(1), commonLabels...)
						},
					)

				} else {
					if err != nil {
						cLoggerEntry.WithFields(log.Fields{
							"gcpStatusEndpoint": gcpStatusEndpoint,
						}).Errorf("Error adding KeyStatusImpact tag to metrics: %v", err)
					}

					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Recording metric monitoring/gcp_service_up")

					commonLabels := []attribute.KeyValue{attribute.String("gcp_service_name", v.ServiceName), attribute.String("status_impact", v.StatusImpact)}
					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Recording metric monitoring/gcp_service_up")

					err = meter.RegisterCallback([]instrument.Asynchronous{
						mGCPServiceStatus,
					},

						func(ctx context.Context) {
							mGCPServiceStatus.Observe(ctx, int64(0), commonLabels...)
						},
					)

				}

				if time.Since(v.MostRecentUpdate.When).Seconds() > 2592000 {
					fmt.Println("Time diff", time.Since(v.MostRecentUpdate.When).Seconds())

					incident_uri := "https://status.cloud.google.com" + "/" + v.URI
					keyLabels := []attribute.KeyValue{
						attribute.String("gcp_service_name", v.ServiceName),
						attribute.String("severity", v.Severity),
						attribute.String("most_recent_status", v.MostRecentUpdate.Status),
						attribute.String("status_impact", v.StatusImpact),
						attribute.String("id", v.ID),
						attribute.String("number", v.Number),
						attribute.String("uri", incident_uri),
					}

					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Recording metric monitoring/gcp_incident_begin")

					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Recording metric monitoring/gcp_service_up")

					err = meter.RegisterCallback([]instrument.Asynchronous{
						mGCPIncidentBegin,
					},

						func(ctx context.Context) {
							mGCPIncidentBegin.Observe(ctx, float64(v.Begin.Unix()), keyLabels...)
						},
					)

					if err != nil {
						cLoggerEntry.WithFields(log.Fields{
							"gcpStatusEndpoint": gcpStatusEndpoint,
						}).Errorf("Error recording metric mGCPIncidentBegin: %v", err)

					}

					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Recording metric monitoring/gcp_incident_created")

					err = meter.RegisterCallback([]instrument.Asynchronous{
						mGCPIncidentCreated,
					},

						func(ctx context.Context) {
							mGCPIncidentCreated.Observe(ctx, float64(v.Created.Unix()), keyLabels...)
						},
					)

					if err != nil {
						cLoggerEntry.WithFields(log.Fields{
							"gcpStatusEndpoint": gcpStatusEndpoint,
						}).Errorf("Error recording metric mGCPIncidentCreated: %v", err)

					}

					if v.MostRecentUpdate.Status != "" {
						keyLabels := []attribute.KeyValue{
							attribute.String("gcp_service_name", v.ServiceName),
							attribute.String("severity", v.Severity),
							attribute.String("most_recent_status", v.MostRecentUpdate.Status),
							attribute.String("status_impact", v.StatusImpact),
							attribute.String("id", v.ID),
							attribute.String("number", v.Number),
						}

						cLoggerEntry.WithFields(log.Fields{
							"gcpStatusEndpoint": gcpStatusEndpoint,
						}).Info("Recording metric monitoring/gcp_most_recent_update")

						err = meter.RegisterCallback([]instrument.Asynchronous{
							mGCPMostRecentUpdate,
						},

							func(ctx context.Context) {
								mGCPMostRecentUpdate.Observe(ctx, float64(v.MostRecentUpdate.When.Unix()), keyLabels...)
							},
						)

						if err != nil {
							cLoggerEntry.WithFields(log.Fields{
								"gcpStatusEndpoint": gcpStatusEndpoint,
							}).Errorf("Error recording metric mGCPMostRecentUpdate: %v", err)

						}

						// Record seconds incidents lasted
						cLoggerEntry.WithFields(log.Fields{
							"gcpStatusEndpoint": gcpStatusEndpoint,
						}).Info("Recording metric monitoring/gcp_incident_resolution_seconds")

						err = meter.RegisterCallback([]instrument.Asynchronous{
							mGCPIncidentResolutionTotal,
						},

							func(ctx context.Context) {
								mGCPIncidentResolutionTotal.Observe(ctx, float64(v.MostRecentUpdate.When.Sub(v.Created).Seconds()), keyLabels...)
							},
						)

						if err != nil {
							cLoggerEntry.WithFields(log.Fields{
								"gcpStatusEndpoint": gcpStatusEndpoint,
							}).Errorf("Error recording metric mGCPIncidentResolutionTotal: %v", err)

						}

					}

				} else {
					cLoggerEntry.WithFields(log.Fields{
						"gcpStatusEndpoint": gcpStatusEndpoint,
					}).Info("Skipped metric recording because the last update was more than 30 days ago")
				}
			}
		}
	}

}
