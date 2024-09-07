package restexporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type restExporter struct {
	// 設定やクライアントなどのフィールドをここに追加
}

func newRestExporter(_ context.Context, _ exporter.Settings) (exporter.Metrics, error) {
	fmt.Println("StnewRestExporterart Called.")
	return &restExporter{}, nil
}

func (e *restExporter) Capabilities() consumer.Capabilities {
	fmt.Println("Capabilities Called.")
	return consumer.Capabilities{MutatesData: false}
}

func (e *restExporter) Start(_ context.Context, _ component.Host) error {
	fmt.Println("Start Called.")
	return nil
}

func (e *restExporter) Shutdown(_ context.Context) error {
	fmt.Println("Shutdown Called.")
	return nil
}

func (e *restExporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	// ここでメトリクスを処理し、RESTエンドポイントに送信する
	fmt.Println("Received metrics:", md.MetricCount())
	return nil
}

// Config holds the configuration for the REST exporter.
type Config struct {
	// 設定フィールドをここに記述
	Endpoint string `mapstructure:"endpoint"`
}

// NewFactory creates a factory for REST exporter.
func NewFactory() exporter.Factory {
	fmt.Println("NewFactory Called.")
	restType, err := component.NewType("rest")
	if err != nil {
		// エラーハンドリング。例えば：
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
		// デフォルト設定をここで指定
		Endpoint: "http://default-endpoint:8890",
	}
}

func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	fmt.Println("endpoint:", cfg.(*Config))
	return newRestExporter(ctx, set)
}
