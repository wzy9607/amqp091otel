package amqp091otel

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/wzy9607/amqp091otel/internal/mocks/amqp091"
)

func initMockTracerProvider() (*tracesdk.TracerProvider, *tracetest.InMemoryExporter) {
	exp := tracetest.NewInMemoryExporter()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSyncer(exp))
	return tp, exp
}

func Test_acknowledger(t *testing.T) {
	t.Parallel()
	type fields struct {
		ch    *Channel
		acker *mockamqp091.MockAcknowledger
		ctx   context.Context
		span  trace.Span
	}
	type args struct {
		tag      uint64
		multiple bool
		requeue  bool
	}
	tests := []struct {
		name                 string
		setup                func(t *testing.T, fields *fields, args args) (exp *tracetest.InMemoryExporter)
		fields               fields
		args                 args
		wantErr              bool
		wantNonEndedSpansLen int
		wantEndedSpansLen    int
		wantSpanStatus       tracesdk.Status
	}{
		{
			name: "ack one",
			setup: func(t *testing.T, fields *fields, args args) (exp *tracetest.InMemoryExporter) {
				t.Helper()
				tp, exp := initMockTracerProvider()
				_, fields.ch.spanMap[args.tag] = tp.Tracer("test").Start(fields.ctx, "should end 1")
				_, fields.ch.spanMap[args.tag+1] = tp.Tracer("test").Start(fields.ctx, "should not end")
				fields.span = fields.ch.spanMap[args.tag]

				fields.acker.EXPECT().Ack(args.tag, args.multiple).Return(nil)
				return exp
			},
			args: args{
				tag: 2,
			},
			wantNonEndedSpansLen: 1,
			wantEndedSpansLen:    1,
			wantSpanStatus:       tracesdk.Status{Code: codes.Ok, Description: ""},
		}, {
			name: "ack multiple",
			setup: func(t *testing.T, fields *fields, args args) (exp *tracetest.InMemoryExporter) {
				t.Helper()
				tp, exp := initMockTracerProvider()
				_, fields.ch.spanMap[args.tag] = tp.Tracer("test").Start(fields.ctx, "should end 1")
				_, fields.ch.spanMap[args.tag-1] = tp.Tracer("test").Start(fields.ctx, "should end 2")
				_, fields.ch.spanMap[args.tag+1] = tp.Tracer("test").Start(fields.ctx, "should not end")
				fields.span = fields.ch.spanMap[args.tag]

				fields.acker.EXPECT().Ack(args.tag, args.multiple).Return(nil)
				return exp
			},
			args: args{
				tag:      2,
				multiple: true,
			},
			wantNonEndedSpansLen: 1,
			wantEndedSpansLen:    2,
			wantSpanStatus:       tracesdk.Status{Code: codes.Ok, Description: ""},
		}, {
			name: "nack one",
			setup: func(t *testing.T, fields *fields, args args) (exp *tracetest.InMemoryExporter) {
				t.Helper()
				tp, exp := initMockTracerProvider()
				_, fields.ch.spanMap[args.tag] = tp.Tracer("test").Start(fields.ctx, "should end 1")
				_, fields.ch.spanMap[args.tag+1] = tp.Tracer("test").Start(fields.ctx, "should not end")
				fields.span = fields.ch.spanMap[args.tag]

				fields.acker.EXPECT().Nack(args.tag, args.multiple, args.requeue).Return(nil)
				return exp
			},
			args: args{
				tag: 2,
			},
			wantNonEndedSpansLen: 1,
			wantEndedSpansLen:    1,
			wantSpanStatus:       tracesdk.Status{Code: codes.Error, Description: "nack"},
		}, {
			name: "nack multiple, got error",
			setup: func(t *testing.T, fields *fields, args args) (exp *tracetest.InMemoryExporter) {
				t.Helper()
				tp, exp := initMockTracerProvider()
				_, fields.ch.spanMap[args.tag] = tp.Tracer("test").Start(fields.ctx, "should end 1")
				_, fields.ch.spanMap[args.tag-1] = tp.Tracer("test").Start(fields.ctx, "should end 2")
				_, fields.ch.spanMap[args.tag+1] = tp.Tracer("test").Start(fields.ctx, "should not end")
				fields.span = fields.ch.spanMap[args.tag]

				fields.acker.EXPECT().Nack(args.tag, args.multiple, args.requeue).Return(errors.New("some error"))
				return exp
			},
			args: args{
				tag:      2,
				multiple: true,
			},
			wantErr:              true,
			wantNonEndedSpansLen: 1,
			wantEndedSpansLen:    2,
			wantSpanStatus:       tracesdk.Status{Code: codes.Error, Description: "nack"},
		}, {
			name: "reject, got error",
			setup: func(t *testing.T, fields *fields, args args) (exp *tracetest.InMemoryExporter) {
				t.Helper()
				tp, exp := initMockTracerProvider()
				_, fields.ch.spanMap[args.tag] = tp.Tracer("test").Start(fields.ctx, "should end 1")
				_, fields.ch.spanMap[args.tag+1] = tp.Tracer("test").Start(fields.ctx, "should not end")
				fields.span = fields.ch.spanMap[args.tag]

				fields.acker.EXPECT().Reject(args.tag, args.requeue).Return(errors.New("some error"))
				return exp
			},
			args: args{
				tag: 2,
			},
			wantErr:              true,
			wantNonEndedSpansLen: 1,
			wantEndedSpansLen:    1,
			wantSpanStatus:       tracesdk.Status{Code: codes.Error, Description: "reject"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fields = fields{
				ch:    &Channel{spanMap: make(map[uint64]trace.Span)},
				acker: mockamqp091.NewMockAcknowledger(t),
				ctx:   context.Background(),
			}
			exp := tt.setup(t, &tt.fields, tt.args)
			a := &acknowledger{
				ch:    tt.fields.ch,
				acker: tt.fields.acker,
				ctx:   tt.fields.ctx,
				span:  tt.fields.span,
			}
			var err error
			switch {
			case strings.HasPrefix(tt.name, "ack"):
				err = a.Ack(tt.args.tag, tt.args.multiple)
			case strings.HasPrefix(tt.name, "nack"):
				err = a.Nack(tt.args.tag, tt.args.multiple, tt.args.requeue)
			case strings.HasPrefix(tt.name, "reject"):
				err = a.Reject(tt.args.tag, tt.args.requeue)
			default:
				t.Fatalf("unknown test case: %s", tt.name)
			}
			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			spans := exp.GetSpans()
			// only check ended spans
			spans = slices.DeleteFunc(spans, func(s tracetest.SpanStub) bool { return s.EndTime.IsZero() })
			assert.Len(t, spans, tt.wantEndedSpansLen)
			for _, span := range spans {
				assert.Truef(t, strings.HasPrefix(span.Name, "should end"), "span %s is not expected to be ended", span.Name)
				assert.Equal(t, tt.wantSpanStatus, span.Status)
				if tt.wantErr {
					assert.Len(t, span.Events, 1)
				} else {
					assert.Empty(t, span.Events)
				}
			}

			assert.Lenf(t, tt.fields.ch.spanMap, tt.wantNonEndedSpansLen, "spanMap should only have non-ended spans")
			for tag := range tt.fields.ch.spanMap {
				assert.Greater(t, tag, tt.args.tag)
			}
		})
	}
}
