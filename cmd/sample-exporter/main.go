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
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
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

	// Create a counter
	counter, err := meter.Int64Counter("my_counter")
	if err != nil {
		log.Fatalf("Failed to create counter: %v", err)
	}

	// Send metrics
	for {
		counter.Add(ctx, 1)
		log.Println("Metric sent")
		time.Sleep(5 * time.Second)
	}
}
