FROM golang:1.22 as otel-builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o custom-otel-collector

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=otel-builder /app/custom-otel-collector /usr/local/bin/custom-otel-collector
EXPOSE 4317 8889 8890
CMD ["/usr/local/bin/custom-otel-collector", "--config=/etc/otel-collector-config.yaml"]