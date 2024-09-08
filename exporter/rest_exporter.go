package restexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricDetail struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	HostName    string `json:"hostname"`
	ServiceName string `json:"sercicename"`
}

// Config holds the configuration for the REST exporter.
type Config struct {
	// 設定フィールドをここに記述
	Endpoint string `mapstructure:"endpoint"`
}

type MetricKey struct {
	ServiceName string
	HostName    string
	MetricName  string
}

type MetricValue struct {
	ResourceMetrics pmetric.Metric
	Timestamp       time.Time
}

type restExporter struct {
	config        *Config
	metricsMux    sync.RWMutex
	latestMetrics map[MetricKey]MetricValue
	server        *http.Server
}

func newRestExporter(_ context.Context, _ exporter.Settings, cfg *Config) (exporter.Metrics, error) {
	fmt.Println("newRestExporter Called.")
	return &restExporter{
		config:        cfg,
		latestMetrics: make(map[MetricKey]MetricValue),
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
	http.HandleFunc("/metrics/", e.handleSpecificMetrics)
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

func getAttributeValue(attrs pcommon.Map, key string) string {
	val, ok := attrs.Get(key)
	if !ok {
		return "unknown"
	}
	return val.AsString()
}

func (e *restExporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	e.metricsMux.Lock()
	defer e.metricsMux.Unlock()
	fmt.Println("Received metrics:", md.MetricCount())

	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)

		// サービス名とホスト名を取得
		serviceName := getAttributeValue(rm.Resource().Attributes(), "service.name")
		hostName := getAttributeValue(rm.Resource().Attributes(), "host.name")

		sms := rm.ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			sm := sms.At(j)
			metrics := sm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)

				key := MetricKey{
					ServiceName: serviceName,
					HostName:    hostName,
					MetricName:  metric.Name(),
				}

				value := MetricValue{
					ResourceMetrics: metric,
					Timestamp:       time.Now(),
				}

				e.latestMetrics[key] = value
			}
		}
	}
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
		MetricsCount: len(e.latestMetrics),
		Metrics:      []MetricDetail{},
	}

	// メトリクスの詳細をログに出力
	log.Printf("Total metric count: %d\n", len(e.latestMetrics))

	for key, value := range e.latestMetrics {
		log.Printf("Metric: %s/%s/%s\n", key.HostName, key.ServiceName, key.MetricName)

		metric := value.ResourceMetrics
		log.Printf("  Name: %s\n", metric.Name())
		log.Printf("  Description: %s\n", metric.Description())
		log.Printf("  Unit: %s\n", metric.Unit())
		log.Printf("  Type: %s\n", metric.Type().String())

		// メトリックタイプに応じたデータポイントのログ出力
		switch metric.Type() {
		case pmetric.MetricTypeGauge:
			dp := metric.Gauge().DataPoints().At(0)
			log.Printf("  Value: %v\n", dp.DoubleValue())
		case pmetric.MetricTypeSum:
			dp := metric.Sum().DataPoints().At(0)
			log.Printf("  Value: %v\n", dp.DoubleValue())
			// 他のメトリックタイプも必要に応じて追加
		}

		response.Metrics = append(response.Metrics, MetricDetail{
			Name:        metric.Name(),
			Type:        metric.Type().String(),
			HostName:    key.HostName,
			ServiceName: key.ServiceName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (e *restExporter) handleSpecificMetrics(w http.ResponseWriter, r *http.Request) {
	e.metricsMux.RLock()
	defer e.metricsMux.RUnlock()

	// URIからパラメータを抽出
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid URI format. Expected /metrics/{service.host}/{service.name}/{Name}", http.StatusBadRequest)
		return
	}

	hostName := parts[2]
	serviceName := parts[3]
	metricName := parts[4]

	key := MetricKey{
		ServiceName: serviceName,
		HostName:    hostName,
		MetricName:  metricName,
	}

	metricValue, exists := e.latestMetrics[key]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metric not found",
			"key":   key,
		})
		return
	}

	// メトリクスデータを直接取得
	metric := metricValue.ResourceMetrics
	var metricData interface{}

	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		metricData = metric.Gauge().DataPoints().At(0).DoubleValue()
	case pmetric.MetricTypeSum:
		metricData = metric.Sum().DataPoints().At(0).DoubleValue()
	// 他のメトリックタイプも必要に応じて追加
	default:
		metricData = fmt.Sprintf("Unsupported metric type: %s", metric.Type().String())
	}

	response := map[string]interface{}{
		"service_host": hostName,
		"service_name": serviceName,
		"metric_name":  metricName,
		"timestamp":    metricValue.Timestamp,
		"data":         metricData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
