package amqp091otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const libName = "github.com/wzy9607/amqp091otel"
const amqpLibName = "github.com/rabbitmq/amqp091-go"

type config struct {
	TracerProvider trace.TracerProvider
	Propagators    propagation.TextMapPropagator

	Tracer trace.Tracer
}

func newConfig(opts []Option) *config {
	cfg := &config{
		Propagators:    otel.GetTextMapPropagator(),
		TracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	cfg.Tracer = cfg.TracerProvider.Tracer(
		libName,
		trace.WithInstrumentationVersion(version),
	)

	return cfg
}

// Option sets optional config properties.
type Option func(*config)

// WithTracerProvider sets the tracer provider.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(cfg *config) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	}
}

// WithPropagators sets the propagators.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return func(cfg *config) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	}
}
