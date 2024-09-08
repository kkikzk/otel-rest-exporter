package restexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
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
	log.Println("Received request for /metrics")

	response := struct {
		MetricsCount int            `json:"metrics_count"`
		Metrics      []MetricDetail `json:"metrics"`
	}{
		MetricsCount: e.latestMetrics.MetricCount(),
		Metrics:      []MetricDetail{},
	}

	// メトリクスの詳細をログに出力
	log.Printf("Total metric count: %d\n", e.latestMetrics.MetricCount())

	rms := e.latestMetrics.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		log.Printf("Resource Metrics #%d:\n", i)

		// リソース属性のログ出力
		attrs := rm.Resource().Attributes()
		attrs.Range(func(k string, v pcommon.Value) bool {
			log.Printf("  Resource Attribute: %s = %v\n", k, v.AsString())
			return true
		})

		sms := rm.ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			sm := sms.At(j)
			log.Printf("  Scope Metrics #%d:\n", j)
			log.Printf("    Scope Name: %s\n", sm.Scope().Name())

			metrics := sm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				log.Printf("    Metric #%d:\n", k)
				log.Printf("      Name: %s\n", metric.Name())
				log.Printf("      Description: %s\n", metric.Description())
				log.Printf("      Unit: %s\n", metric.Unit())
				log.Printf("      Type: %s\n", metric.Type().String())

				// メトリックタイプに応じたデータポイントのログ出力
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					dp := metric.Gauge().DataPoints().At(0)
					log.Printf("      Value: %v\n", dp.DoubleValue())
				case pmetric.MetricTypeSum:
					dp := metric.Sum().DataPoints().At(0)
					log.Printf("      Value: %v\n", dp.DoubleValue())
					// 他のメトリックタイプも必要に応じて追加
				}

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
