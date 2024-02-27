package altnrslog

import (
	"context"
	"errors"
	"github.com/google/go-cmp/cmp"
	mslog "github.com/miyamo2/altnrslog/internal/mock"
	"github.com/newrelic/go-agent/v3/integrations/logcontext"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/mock/gomock"
	"io"
	"log/slog"
	"reflect"
	"testing"
)

type mockWriter struct{}

func (m mockWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func Test_attrsFromMetadata(t *testing.T) {
	type test struct {
		args newrelic.LinkingMetadata
		want []slog.Attr
	}

	tests := map[string]test{
		"happy-path: valid linking metadata": {
			args: newrelic.LinkingMetadata{
				TraceID:    "trace-id",
				SpanID:     "span-id",
				EntityName: "entity-name",
				EntityType: "entity-type",
				EntityGUID: "entity-guid",
				Hostname:   "hostname",
			},
			want: []slog.Attr{
				slog.String(logcontext.KeyTraceID, "trace-id"),
				slog.String(logcontext.KeySpanID, "span-id"),
				slog.String(logcontext.KeyEntityName, "entity-name"),
				slog.String(logcontext.KeyEntityType, "entity-type"),
				slog.String(logcontext.KeyEntityGUID, "entity-guid"),
				slog.String(logcontext.KeyHostname, "hostname"),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := attrsFromMetadata(tt.args)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_NewTransactionalHandler_WithJSONHandler(t *testing.T) {
	app := newrelic.Application{}
	tx := newrelic.Transaction{}
	options := []HandlerOption{WithSlogHandlerSpecify(true, nil)}
	handler := NewTransactionalHandler(&app, &tx, options...)
	_, ok := handler.handler.(*slog.JSONHandler)
	if !ok {
		t.Errorf("NewTransactionalHandler() = %T, want %T", handler.handler, &slog.JSONHandler{})
	}
}

func Test_NewTransactionalHandler_WithTextHandler(t *testing.T) {
	app := newrelic.Application{}
	tx := newrelic.Transaction{}
	options := []HandlerOption{WithSlogHandlerSpecify(false, nil)}
	handler := NewTransactionalHandler(&app, &tx, options...)
	_, ok := handler.handler.(*slog.TextHandler)
	if !ok {
		t.Errorf("NewTransactionalHandler() = %T, want %T", handler.handler, &slog.TextHandler{})
	}
}

func Test_NewTransactionalHandler_WithInnerHandler(t *testing.T) {
	app := newrelic.Application{}
	tx := newrelic.Transaction{}
	options := []HandlerOption{WithInnerHandlerProvider(func(w io.Writer) slog.Handler {
		return &mslog.MockHandler{}
	})}
	handler := NewTransactionalHandler(&app, &tx, options...)
	_, ok := handler.handler.(*mslog.MockHandler)
	if !ok {
		t.Errorf("NewTransactionalHandler() = %T, want %T", handler.handler, &mslog.MockHandler{})
	}
}

func TestTransactionalHandler_WithAttrs(t *testing.T) {
	app := newrelic.Application{}
	tx := newrelic.Transaction{}
	sut := NewTransactionalHandler(&app, &tx, WithSlogHandlerSpecify(true, nil))
	handler := sut.WithAttrs([]slog.Attr{})
	outerHandler, ok := handler.(*TransactionalHandler)
	if !ok {
		t.Errorf("WithAttrs() = %T, want %T", handler, &TransactionalHandler{})
	}
	_, ok = outerHandler.handler.(*slog.JSONHandler)
	if !ok {
		t.Errorf("WithAttrs() inner = %T, want %T", outerHandler.handler, &slog.JSONHandler{})
	}
}

func TestTransactionalHandler_WithGroup(t *testing.T) {
	app := newrelic.Application{}
	tx := newrelic.Transaction{}
	sut := NewTransactionalHandler(&app, &tx, WithSlogHandlerSpecify(true, nil))
	handler := sut.WithGroup("foo")
	outerHandler, ok := handler.(*TransactionalHandler)
	if !ok {
		t.Errorf("WithGroup() = %T, want %T", handler, &TransactionalHandler{})
	}
	_, ok = outerHandler.handler.(*slog.JSONHandler)
	if !ok {
		t.Errorf("WithGroup() inner = %T, want %T", outerHandler.handler, &slog.JSONHandler{})
	}
}

func TestTransactionalHandler_Enabled(t *testing.T) {
	type fields struct {
		handler slog.Handler
	}
	type args struct {
		ctx   context.Context
		level slog.Level
	}
	type test struct {
		fields fields
		args   args
		want   bool
	}
	tests := map[string]test{
		"happy-path enable": {
			fields: fields{
				handler: slog.NewTextHandler(&mockWriter{}, &slog.HandlerOptions{
					Level: slog.LevelWarn,
				}),
			},
			args: args{
				ctx:   context.Background(),
				level: slog.LevelWarn,
			},
			want: true,
		},
		"happy-path disable": {
			fields: fields{
				handler: slog.NewTextHandler(&mockWriter{}, &slog.HandlerOptions{
					Level: slog.LevelWarn,
				}),
			},
			args: args{
				ctx:   context.Background(),
				level: slog.LevelInfo,
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			h := &TransactionalHandler{
				handler: tt.fields.handler,
			}
			if got := h.Enabled(tt.args.ctx, tt.args.level); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildProperties(t *testing.T) {
	type args struct {
		options []HandlerOption
	}
	type test struct {
		args args
		want *Properties
	}

	CmpInnerHandlerProvider := func() cmp.Option {
		return cmp.FilterValues(
			func(x, y InnerHandlerProvider) bool {
				return x != nil && y != nil
			},
			cmp.Transformer("ToPtr", func(in InnerHandlerProvider) (out uintptr) {
				return reflect.ValueOf(in).Pointer()
			}))
	}

	mockHandlerProvider := func(w io.Writer) slog.Handler {
		return &mslog.MockHandler{}
	}

	tests := map[string]test{
		"happy-path: WithInnerWriter": {
			args: args{
				options: []HandlerOption{WithInnerWriter(&mockWriter{})},
			},
			want: &Properties{
				innerWriter: &mockWriter{},
			},
		},
		"happy-path: WithSlogHandlerSpecify": {
			args: args{
				options: []HandlerOption{WithSlogHandlerSpecify(true, &slog.HandlerOptions{
					AddSource: true,
					Level:     slog.LevelDebug,
				})},
			},
			want: &Properties{
				json: true,
				slogHandlerOptions: &slog.HandlerOptions{
					AddSource: true,
					Level:     slog.LevelDebug,
				},
			},
		},
		"happy-path: WithInnerHandlerProvider": {
			args: args{
				options: []HandlerOption{WithInnerHandlerProvider(mockHandlerProvider)},
			},
			want: &Properties{
				innerHandlerProvider: mockHandlerProvider,
			},
		},
	}
	opt := cmp.AllowUnexported(Properties{})
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := buildProperties(tt.args.options)
			if diff := cmp.Diff(*got, *tt.want, opt, CmpInnerHandlerProvider()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestTransactionalHandler_Handle(t *testing.T) {

	type args struct {
		r slog.Record
	}

	type mockExpect struct {
		err error
	}

	type test struct {
		args       args
		mockExpect mockExpect
		want       error
	}
	tests := map[string]test{
		"happy-path": {
			args: args{
				r: slog.Record{},
			},
			mockExpect: mockExpect{
				err: nil,
			},
			want: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockHandler := mslog.NewMockHandler(mockCtrl)
			mockHandler.EXPECT().Handle(gomock.Any(), testHelper_DummySlogRecord(t)).Return(tt.mockExpect.err).Times(1)
			h := &TransactionalHandler{
				handler: mockHandler,
				tx:      &newrelic.Transaction{},
			}
			err := h.Handle(context.Background(), tt.args.r)
			if !errors.Is(err, tt.want) {
				t.Errorf("Handle() = %v, want %v", err, tt.want)
			}
		})
	}
}

func testHelper_DummySlogRecord(t *testing.T) slog.Record {
	t.Helper()
	r := slog.Record{}
	r.AddAttrs(attrsFromMetadata(newrelic.LinkingMetadata{})...)
	return r
}
