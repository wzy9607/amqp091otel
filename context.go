package amqp091otel

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

// ContextFromDelivery returns the context that contains the traces associated with the amqp091.Delivery.
// If the delivery does not contain traces, it returns a background context.
// Consumer can use this function to continue the distributed tracing.
func ContextFromDelivery(msg amqp091.Delivery) context.Context {
	if ack, ok := msg.Acknowledger.(*acknowledger); ok {
		return ack.ctx
	}
	return context.Background()
}
