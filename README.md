# OpenTelemetry REST Exporter

This project provides a custom OpenTelemetry exporter that exposes metrics via a REST API. It allows you to retrieve the latest metrics data through HTTP endpoints.

## Features

- Exposes metrics data via REST API
- Supports retrieving all metrics or specific metrics by service host, service name, and metric name
- Implements OpenTelemetry Collector exporter interface
- Easy integration with existing OpenTelemetry setups

## Installation

To use this exporter, you need to include it in your OpenTelemetry Collector build. Add the following import to your collector's `main.go`:

```go
import (
    "github.com/kkikzk/otel-rest-exporter/exporter"
)
```

Then, register the exporter in your collector's factory:

```go
factories, err := componenttest.NopFactories()
if err != nil {
    log.Fatalf("failed to create nop factories: %v", err)
}
factories.Exporters[restexporter.NewFactory().Type()] = restexporter.NewFactory()
```

## Configuration

Add the following to your collector's configuration file:

```yaml
exporters:
  rest:
    endpoint: ":8890"  # The port on which the REST API will be exposed
```

## Usage

Once the collector is running with this exporter, you can access metrics via the following endpoints:

- `/metrics`: Returns all available metrics
- `/metrics/{service.host}/{service.name}/{metric.name}`: Returns a specific metric

### Example

To retrieve a specific metric:

```
GET http://localhost:8890/metrics/host1/service1/my_metric
```

## Building

To build the project:

```
./build.sh
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

kkikzk

## Acknowledgments

- OpenTelemetry community
- All contributors to this project
