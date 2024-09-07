#!/bin/bash
go build -o otel-rest-exporter
cd cmd/sample-exporter
docker-compose build