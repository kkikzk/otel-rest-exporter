package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	ctx := context.Background()

	// Configure connection to OTel Collector
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("otel-collector:4317"),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Set up resource information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("my-service"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	// Configure MeterProvider
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)
	otel.SetMeterProvider(mp)

	// Create a meter
	meter := otel.Meter("my-meter")
	meter2 := otel.Meter("my-meter2")

	// Create a counter
	counter, err := meter.Int64Counter("my_counter")
	if err != nil {
		log.Fatalf("Failed to create counter: %v", err)
	}
	counter2, err := meter2.Int64Counter("my_counter2")
	if err != nil {
		log.Fatalf("Failed to create counter2: %v", err)
	}

	// Send metrics
	for {
		counter.Add(ctx, 1)
		counter2.Add(ctx, 5)
		log.Println("Metric sent")
		time.Sleep(5 * time.Second)
	}
}
