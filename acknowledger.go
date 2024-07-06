package amqp091otel

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/rabbitmq/amqp091-go"
)

type acknowledger struct {
	ch    *Channel
	acker amqp091.Acknowledger // The real acknowledger is amqp091.Channel
	ctx   context.Context      //nolint:containedctx // consumer needs to retrieve the context via ContextFromDelivery.
	span  trace.Span
}

func (a *acknowledger) Ack(tag uint64, multiple bool) error {
	err := a.acker.Ack(tag, multiple)
	if multiple {
		a.endMultiple(tag, codes.Ok, "ack", err)
	} else {
		a.endOne(tag, codes.Ok, "ack", err)
	}
	return err
}

func (a *acknowledger) Nack(tag uint64, multiple, requeue bool) error {
	err := a.acker.Nack(tag, multiple, requeue)
	if multiple {
		a.endMultiple(tag, codes.Error, "nack", err)
	} else {
		a.endOne(tag, codes.Error, "nack", err)
	}
	return err
}

func (a *acknowledger) Reject(tag uint64, requeue bool) error {
	err := a.acker.Reject(tag, requeue)
	a.endOne(tag, codes.Error, "reject", err)
	return err
}

func (a *acknowledger) endMultiple(lastTag uint64, code codes.Code, desc string, err error) {
	a.ch.m.Lock()
	defer a.ch.m.Unlock()

	for tag, span := range a.ch.spanMap {
		span.SetAttributes(semconv.MessagingOperationName(desc))
		if tag <= lastTag {
			if err != nil {
				span.RecordError(err)
			}
			span.SetStatus(code, desc)
			span.End()
			delete(a.ch.spanMap, tag)
		}
	}
}

func (a *acknowledger) endOne(tag uint64, code codes.Code, desc string, err error) {
	a.ch.m.Lock()
	defer a.ch.m.Unlock()

	a.span.SetAttributes(semconv.MessagingOperationName(desc))
	if err != nil {
		a.span.RecordError(err)
	}
	a.span.SetStatus(code, desc)
	a.span.End()
	delete(a.ch.spanMap, tag)
}
