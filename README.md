# OpenTelemetry RabbitMQ Client Instrumentation for Golang

[![ci](https://github.com/wzy9607/amqp091otel/actions/workflows/pull-request.yml/badge.svg)](https://github.com/wzy9607/amqp091otel/actions/workflows/pull-request.yml)
[![codecov](https://codecov.io/gh/wzy9607/amqp091otel/graph/badge.svg?token=3994PBP60N)](https://codecov.io/gh/wzy9607/amqp091otel)

This module provides OpenTelemetry instrumentation for the Go RabbitMQ Client
Library [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go).

As for now, this module only provides tracing instrumentation.

## Compatibility

This module supports the same go versions as the
[opentelemetry-go project](https://github.com/open-telemetry/opentelemetry-go?tab=readme-ov-file#compatibility).

## Installation

```bash
go get -u github.com/wzy9607/amqp091otel
```
