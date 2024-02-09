package amqp091otel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rabbitmq/amqp091-go"
)

func Test_publishingMessageCarrier_Get(t *testing.T) {
	t.Parallel()
	type fields struct {
		msg *amqp091.Publishing
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "exists",
			fields: fields{msg: &amqp091.Publishing{Headers: amqp091.Table{"foo": "bar"}}},
			args:   args{key: "foo"},
			want:   "bar",
		}, {
			name:   "not exists",
			fields: fields{msg: &amqp091.Publishing{Headers: amqp091.Table{"foo1": "bar"}}},
			args:   args{key: "foo"},
			want:   "",
		}, {
			name:   "ignore not string",
			fields: fields{msg: &amqp091.Publishing{Headers: amqp091.Table{"foo": 1}}},
			args:   args{key: "foo"},
			want:   "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := publishingMessageCarrier{
				msg: tt.fields.msg,
			}
			if got := c.Get(tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_publishingMessageCarrier_Set(t *testing.T) {
	t.Parallel()
	msg := &amqp091.Publishing{}
	carrier := newPublishingMessageCarrier(msg)

	carrier.Set("foo", "bar")
	carrier.Set("foo1", "bar1")
	carrier.Set("foo1", "bar2")
	carrier.Set("foo2", "bar3")

	assert.Equal(t, carrier.msg.Headers, amqp091.Table{"foo": "bar", "foo1": "bar2", "foo2": "bar3"})
}

func Test_publishingMessageCarrier_Keys(t *testing.T) {
	t.Parallel()
	type fields struct {
		msg *amqp091.Publishing
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "empty",
			fields: fields{msg: &amqp091.Publishing{Headers: amqp091.Table{}}},
			want:   []string{},
		}, {
			name:   "one",
			fields: fields{msg: &amqp091.Publishing{Headers: amqp091.Table{"foo": "bar"}}},
			want:   []string{"foo"},
		}, {
			name:   "many",
			fields: fields{msg: &amqp091.Publishing{Headers: amqp091.Table{"foo": "bar", "foo1": "bar1"}}},
			want:   []string{"foo", "foo1"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := publishingMessageCarrier{
				msg: tt.fields.msg,
			}
			got := c.Keys()
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_deliveryMessageCarrier_Get(t *testing.T) {
	t.Parallel()
	type fields struct {
		msg *amqp091.Delivery
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "exists",
			fields: fields{msg: &amqp091.Delivery{Headers: amqp091.Table{"foo": "bar"}}},
			args:   args{key: "foo"},
			want:   "bar",
		}, {
			name:   "not exists",
			fields: fields{msg: &amqp091.Delivery{Headers: amqp091.Table{"foo1": "bar"}}},
			args:   args{key: "foo"},
			want:   "",
		}, {
			name:   "ignore not string",
			fields: fields{msg: &amqp091.Delivery{Headers: amqp091.Table{"foo": 1}}},
			args:   args{key: "foo"},
			want:   "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := deliveryMessageCarrier{
				msg: tt.fields.msg,
			}
			if got := c.Get(tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deliveryMessageCarrier_Set(t *testing.T) {
	t.Parallel()
	msg := &amqp091.Delivery{}
	carrier := newDeliveryMessageCarrier(msg)

	carrier.Set("foo", "bar")
	carrier.Set("foo1", "bar1")
	carrier.Set("foo1", "bar2")
	carrier.Set("foo2", "bar3")

	assert.Equal(t, carrier.msg.Headers, amqp091.Table{"foo": "bar", "foo1": "bar2", "foo2": "bar3"})
}

func Test_deliveryMessageCarrier_Keys(t *testing.T) {
	t.Parallel()
	type fields struct {
		msg *amqp091.Delivery
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "empty",
			fields: fields{msg: &amqp091.Delivery{Headers: amqp091.Table{}}},
			want:   []string{},
		}, {
			name:   "one",
			fields: fields{msg: &amqp091.Delivery{Headers: amqp091.Table{"foo": "bar"}}},
			want:   []string{"foo"},
		}, {
			name:   "many",
			fields: fields{msg: &amqp091.Delivery{Headers: amqp091.Table{"foo": "bar", "foo1": "bar1"}}},
			want:   []string{"foo", "foo1"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := deliveryMessageCarrier{
				msg: tt.fields.msg,
			}
			got := c.Keys()
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
