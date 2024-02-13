package amqp091otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rabbitmq/amqp091-go"
)

func TestContextFromDelivery(t *testing.T) {
	t.Parallel()
	type key struct{}
	ctxInst := context.WithValue(context.Background(), key{}, "value")
	type args struct {
		msg amqp091.Delivery
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "instrumented context",
			args: args{
				msg: amqp091.Delivery{
					Acknowledger: &acknowledger{ctx: ctxInst},
				},
			},
			want: ctxInst,
		}, {
			name: "background context",
			args: args{
				msg: amqp091.Delivery{},
			},
			want: context.Background(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, ContextFromDelivery(tt.args.msg), "ContextFromDelivery(%v)", tt.args.msg)
		})
	}
}
