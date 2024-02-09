package amqp091otel

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

func ContextFromDelivery(msg amqp091.Delivery) context.Context {
	if ack, ok := msg.Acknowledger.(*acknowledger); ok {
		return ack.ctx
	}
	return context.Background()
}
