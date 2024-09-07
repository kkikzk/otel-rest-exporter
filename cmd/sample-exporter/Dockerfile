FROM golang:1.22 as metrics-sender-builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o metrics-sender

FROM ubuntu:22.04
COPY --from=metrics-sender-builder /app/metrics-sender /usr/local/bin/metrics-sender
CMD ["/usr/local/bin/metrics-sender"]