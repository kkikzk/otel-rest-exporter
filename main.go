package main

import (
	restexporter "github.com/kkikzk/otel-rest-exporter/exporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"
	otlpexporter "go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	envprovider "go.opentelemetry.io/collector/confmap/provider/envprovider"
	fileprovider "go.opentelemetry.io/collector/confmap/provider/fileprovider"
	httpprovider "go.opentelemetry.io/collector/confmap/provider/httpprovider"
	httpsprovider "go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	yamlprovider "go.opentelemetry.io/collector/confmap/provider/yamlprovider"

	//prometheusexporter "go.opentelemetry.io/collector/exporter/prometheusexporter"
	prometheusexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"go.opentelemetry.io/collector/otelcol"
	batchprocessor "go.opentelemetry.io/collector/processor/batchprocessor"
	otlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func main() {
	otlpreceiver := otlpreceiver.NewFactory()
	batchprocessor := batchprocessor.NewFactory()
	restexporter := restexporter.NewFactory()
	otlpexporter := otlpexporter.NewFactory()
	prometheusexporter := prometheusexporter.NewFactory()

	factories := otelcol.Factories{
		Receivers: map[component.Type]receiver.Factory{
			otlpreceiver.Type(): otlpreceiver,
		},
		Processors: map[component.Type]processor.Factory{
			batchprocessor.Type(): batchprocessor,
		},
		Exporters: map[component.Type]exporter.Factory{
			restexporter.Type():       restexporter,
			otlpexporter.Type():       otlpexporter,
			prometheusexporter.Type(): prometheusexporter,
		},
	}

	info := component.BuildInfo{
		Command:     "otel-rest-exporter",
		Description: "OpenTelemetry REST Exporter",
		Version:     "1.0.0",
	}

	settings := otelcol.CollectorSettings{
		Factories: func() (otelcol.Factories, error) {
			return factories, nil
		},
		BuildInfo: info,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					httpprovider.NewFactory(),
					httpsprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
			},
		},
	}

	if err := otelcol.NewCommand(settings).Execute(); err != nil {
		panic(err)
	}
}
