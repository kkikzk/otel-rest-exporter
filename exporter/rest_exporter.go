package restexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type restExporter struct {
	config        *Config
	metricsMux    sync.RWMutex
	latestMetrics pmetric.Metrics
	server        *http.Server
}

func newRestExporter(_ context.Context, _ exporter.Settings, cfg *Config) (exporter.Metrics, error) {
	fmt.Println("newRestExporter Called.")
	return &restExporter{
		config:        cfg,
		latestMetrics: pmetric.NewMetrics(),
	}, nil
}

func (e *restExporter) Capabilities() consumer.Capabilities {
	fmt.Println("Capabilities Called.")
	return consumer.Capabilities{MutatesData: false}
}

func (e *restExporter) Start(_ context.Context, _ component.Host) error {
	fmt.Println("Starting REST exporter on", e.config.Endpoint)
	e.server = &http.Server{Addr: e.config.Endpoint}

	http.HandleFunc("/metrics", e.handleMetrics)

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Test endpoint is working")
	})

	go func() {
		fmt.Println("Server is starting...")
		if err := e.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				fmt.Println("Server closed")
			} else {
				fmt.Printf("Error starting server: %v\n", err)
			}
		}
	}()

	return nil
}

func (e *restExporter) Shutdown(ctx context.Context) error {
	fmt.Println("Shutdown Called.")
	if e.server != nil {
		return e.server.Shutdown(ctx)
	}
	return nil
}

func (e *restExporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	e.metricsMux.Lock()
	defer e.metricsMux.Unlock()
	e.latestMetrics = md
	fmt.Println("Received metrics:", md.MetricCount())
	return nil
}

func (e *restExporter) handleMetrics(w http.ResponseWriter, r *http.Request) {
	e.metricsMux.RLock()
	defer e.metricsMux.RUnlock()
	fmt.Println("Received request for /metrics")

	response := struct {
		MetricsCount int            `json:"metrics_count"`
		Metrics      []MetricDetail `json:"metrics"`
	}{
		MetricsCount: e.latestMetrics.MetricCount(),
		Metrics:      []MetricDetail{},
	}

	rms := e.latestMetrics.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		sms := rm.ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			sm := sms.At(j)
			metrics := sm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				response.Metrics = append(response.Metrics, MetricDetail{
					Name: metric.Name(),
					Type: metric.Type().String(),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type MetricDetail struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Config holds the configuration for the REST exporter.
type Config struct {
	// 設定フィールドをここに記述
	Endpoint string `mapstructure:"endpoint"`
}

func NewFactory() exporter.Factory {
	fmt.Println("NewFactory Called.")
	restType, err := component.NewType("rest")
	if err != nil {
		panic(fmt.Sprintf("failed to create rest type: %v", err))
	}
	return exporter.NewFactory(
		restType,
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelAlpha),
	)
}

func createDefaultConfig() component.Config {
	fmt.Println("createDefaultConfig called.")
	return &Config{
		Endpoint: ":8890",
	}
}

func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	fmt.Println("endpoint:", cfg.(*Config).Endpoint)
	return newRestExporter(ctx, set, cfg.(*Config))
}
