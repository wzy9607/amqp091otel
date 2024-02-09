package amqp091otel

import (
	"context"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/rabbitmq/amqp091-go"
)

const (
	netProtocolVer = "0.9.1"
	messageSystem  = "rabbitmq"
)

func queueAnonymous(queue string) bool {
	return strings.HasPrefix(queue, "amq.gen-")
}

type Channel struct {
	*amqp091.Channel
	uri amqp091.URI
	cfg *config
	// When ack multiple, we need to end spans of every delivery before the tag,
	// so we keep a map of every span that haven't ended.
	spanMap map[uint64]trace.Span
	m       sync.Mutex
}

func NewChannel(amqpChan *amqp091.Channel, url string, opts ...Option) (*Channel, error) {
	uri, err := amqp091.ParseURI(url)
	if err != nil {
		return nil, err
	}
	uri.Password = ""
	cfg := newConfig(opts)
	return &Channel{
		Channel: amqpChan,
		uri:     uri,
		cfg:     cfg,
		spanMap: map[uint64]trace.Span{},
		m:       sync.Mutex{},
	}, nil
}

func (ch *Channel) commonAttrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.ServiceName(amqpLibName),
		semconv.ServiceVersion(amqpLibVersion),
		semconv.MessagingSystem(messageSystem),
		semconv.NetworkProtocolName(ch.uri.Scheme),
		semconv.NetworkProtocolVersion(netProtocolVer),
		semconv.NetworkTransportTCP,
		semconv.ServerAddress(ch.uri.Host),
		semconv.ServerPort(ch.uri.Port),
		// todo network.peer.address, network.peer.port, network.type, server.address, server.port
	}
}

func (*Channel) nameWhenPublish(exchange string) string {
	if exchange == "" {
		exchange = "(default)"
	}
	return exchange + " publish"
}

func (*Channel) nameWhenConsume(queue string) string {
	if queueAnonymous(queue) {
		queue = "(anonymous)"
	}
	return queue + " process"
}

func (ch *Channel) startConsumerSpan(msg *amqp091.Delivery, queue string, operation attribute.KeyValue) {
	// Extract a span context from message
	carrier := newDeliveryMessageCarrier(msg)
	parentCtx := ch.cfg.Propagators.Extract(context.Background(), carrier)

	// Create a span
	attrs := []attribute.KeyValue{
		semconv.MessagingDestinationAnonymous(queueAnonymous(queue)),
		semconv.MessagingDestinationName(queue),
		semconv.MessagingDestinationPublishAnonymous(msg.Exchange == ""),
		semconv.MessagingDestinationPublishName(msg.Exchange),
		semconv.MessagingRabbitmqDestinationRoutingKey(msg.RoutingKey),
		operation,
		// todo messaging.client_id
	}
	if msg.CorrelationId != "" {
		attrs = append(attrs, semconv.MessagingMessageConversationID(msg.CorrelationId))
	}
	if msg.MessageId != "" {
		attrs = append(attrs, semconv.MessagingMessageID(msg.MessageId))
	}
	attrs = append(attrs, ch.commonAttrs()...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	ctx, span := ch.cfg.Tracer.Start(parentCtx, //nolint:spancheck // span ends when msg is ack/nack/rejected
		ch.nameWhenConsume(queue), opts...)
	msg.Acknowledger = &acknowledger{
		ch:   ch,
		ctx:  ctx,
		span: span,
	}

	ch.m.Lock()
	defer ch.m.Unlock()
	ch.spanMap[msg.DeliveryTag] = span
} //nolint:spancheck // span ends when msg is ack/nack/rejected

func (ch *Channel) Consume(
	queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table,
) (<-chan amqp091.Delivery, error) {
	deliveries, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	if err != nil {
		return deliveries, err
	}
	newDeliveries := make(chan amqp091.Delivery)
	go func() {
		for msg := range deliveries {
			ch.startConsumerSpan(&msg, queue, semconv.MessagingOperationDeliver)
			newDeliveries <- msg
		}
		close(newDeliveries)
	}()
	return newDeliveries, nil
}

func (ch *Channel) PublishWithContext(
	ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing,
) error {
	_, err := ch.PublishWithDeferredConfirmWithContext(ctx, exchange, key, mandatory, immediate, msg)
	return err
}

func (ch *Channel) PublishWithDeferredConfirmWithContext(
	ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing,
) (*amqp091.DeferredConfirmation, error) {
	// Create a span.
	attrs := []attribute.KeyValue{
		semconv.MessagingDestinationAnonymous(exchange == ""),
		semconv.MessagingDestinationName(exchange),
		semconv.MessagingRabbitmqDestinationRoutingKey(key),
		semconv.MessagingOperationPublish,
		// todo messaging.client_id
	}
	if msg.CorrelationId != "" {
		attrs = append(attrs, semconv.MessagingMessageConversationID(msg.CorrelationId))
	}
	if msg.MessageId != "" {
		attrs = append(attrs, semconv.MessagingMessageID(msg.MessageId))
	}
	attrs = append(attrs, ch.commonAttrs()...)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}
	ctx, span := ch.cfg.Tracer.Start(ctx, ch.nameWhenPublish(exchange), opts...)

	// Inject current span context
	carrier := newPublishingMessageCarrier(&msg)
	ch.cfg.Propagators.Inject(ctx, carrier)

	dc, err := ch.Channel.PublishWithDeferredConfirmWithContext(ctx, exchange, key, mandatory, immediate, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		// todo error.type
	}
	span.End()
	return dc, err
}

func (ch *Channel) Get(queue string, autoAck bool) (msg amqp091.Delivery, ok bool, err error) {
	msg, ok, err = ch.Channel.Get(queue, autoAck)
	if err != nil || !ok {
		return
	}
	ch.startConsumerSpan(&msg, queue, semconv.MessagingOperationReceive)
	return msg, ok, err
}
