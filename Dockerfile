FROM golang:1.22 as builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o custom-otel-collector

FROM debian:buster-slim
COPY --from=builder /app/custom-otel-collector /usr/local/bin/custom-otel-collector

CMD ["/usr/local/bin/custom-otel-collector", "--config", "/etc/otel-collector-config.yaml"]


FROM ubuntu:22.04

# 必要なパッケージをインストール
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/custom-otel-collector /usr/local/bin/custom-otel-collector
EXPOSE 4317 8889
CMD ["/usr/local/bin/custom-otel-collector", "--config=/etc/otel-collector-config.yaml"]