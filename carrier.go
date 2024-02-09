package amqp091otel

import (
	"go.opentelemetry.io/otel/propagation"

	"github.com/rabbitmq/amqp091-go"
)

var (
	_ propagation.TextMapCarrier = (*publishingMessageCarrier)(nil)
	_ propagation.TextMapCarrier = (*deliveryMessageCarrier)(nil)
)

// publishingMessageCarrier injects and extracts traces from a amqp091.Publishing.
type publishingMessageCarrier struct {
	msg *amqp091.Publishing
}

// newPublishingMessageCarrier creates a new publishingMessageCarrier.
func newPublishingMessageCarrier(msg *amqp091.Publishing) publishingMessageCarrier {
	return publishingMessageCarrier{msg: msg}
}

// Get returns the value associated with the passed key.
func (c publishingMessageCarrier) Get(key string) string {
	valAny, ok := c.msg.Headers[key]
	if !ok {
		return ""
	}
	val, ok := valAny.(string)
	if !ok {
		return ""
	}
	return val
}

// Set stores the key-value pair.
func (c publishingMessageCarrier) Set(key, val string) {
	c.msg.Headers[key] = val
}

// Keys lists the keys stored in this carrier.
func (c publishingMessageCarrier) Keys() []string {
	out := make([]string, 0, len(c.msg.Headers))
	for key := range c.msg.Headers {
		out = append(out, key)
	}
	return out
}

// deliveryMessageCarrier injects and extracts traces from a amqp091.Delivery.
type deliveryMessageCarrier struct {
	msg *amqp091.Delivery
}

// newDeliveryMessageCarrier creates a new deliveryMessageCarrier.
func newDeliveryMessageCarrier(msg *amqp091.Delivery) deliveryMessageCarrier {
	return deliveryMessageCarrier{msg: msg}
}

// Get returns the value associated with the passed key.
func (c deliveryMessageCarrier) Get(key string) string {
	valAny, ok := c.msg.Headers[key]
	if !ok {
		return ""
	}
	val, ok := valAny.(string)
	if !ok {
		return ""
	}
	return val
}

// Set stores the key-value pair.
func (c deliveryMessageCarrier) Set(key, val string) {
	c.msg.Headers[key] = val
}

// Keys lists the keys stored in this carrier.
func (c deliveryMessageCarrier) Keys() []string {
	out := make([]string, 0, len(c.msg.Headers))
	for key := range c.msg.Headers {
		out = append(out, key)
	}
	return out
}
