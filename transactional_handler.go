//go:generate mockgen -destination internal/mock/slog_handler.go log/slog Handler
package altnrslog

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/newrelic/go-agent/v3/integrations/logcontext"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// TransactionalHandler is a [slog.Handler] that adds New Relic distributed tracing metadata to log records.
type TransactionalHandler struct {
	handler slog.Handler
	tx      *newrelic.Transaction
	level   slog.Level
}

// Enabled See: [slog.Handler.Enabled]
func (h *TransactionalHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level && h.handler.Enabled(ctx, level)
}

// Handle adds New Relic distributed tracing metadata to log records before passing them to the wrapped handler.
func (h *TransactionalHandler) Handle(ctx context.Context, r slog.Record) error {
	md := h.tx.GetLinkingMetadata()
	r.AddAttrs(attrsFromMetadata(md)...)
	return h.handler.Handle(ctx, r)
}

// WithAttrs See: [slog.Handler.WithAttrs]
func (h *TransactionalHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TransactionalHandler{
		h.handler.WithAttrs(attrs),
		h.tx,
		h.level,
	}
}

// WithGroup See: [slog.Handler.WithGroup]
func (h *TransactionalHandler) WithGroup(name string) slog.Handler {
	return &TransactionalHandler{
		h.handler.WithGroup(name),
		h.tx,
		h.level,
	}
}

type InnerHandlerProvider func(io.Writer) slog.Handler

// Properties is an options for creating a new [TransactionalHandler].
type Properties struct {
	innerWriter          io.Writer
	json                 bool
	slogHandlerOptions   *slog.HandlerOptions
	innerHandlerProvider InnerHandlerProvider
	logLevel             slog.Level
}

// HandlerOption is a functional option for creating a new [TransactionalHandler].
type HandlerOption func(*Properties)

// WithInnerWriter specifies the [io.Writer] that wraps [logWriter.logWriter]
//
// [logWriter.logWriter]: https://pkg.go.dev/github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter#LogWriter
func WithInnerWriter(w io.Writer) HandlerOption {
	return func(p *Properties) {
		p.innerWriter = w
	}
}

// WithSlogHandlerSpecify specifies whether to use JSON format and [slog.HandlerOptions].
func WithSlogHandlerSpecify(json bool, o *slog.HandlerOptions) HandlerOption {
	return func(p *Properties) {
		p.json = json
		p.slogHandlerOptions = o
	}
}

// WithInnerHandlerProvider specifies the function that provides the [slog.Handler] to be wrapped.
func WithInnerHandlerProvider(innerHandlerProvider InnerHandlerProvider) HandlerOption {
	return func(p *Properties) {
		p.slogHandlerOptions = nil
		p.innerHandlerProvider = innerHandlerProvider
	}
}

// WithLogLevel specifies the log level.
// if not specified, the default is [slog.LevelInfo].
// if lower than the inner handler's level, the inner handler's level will be used.
func WithLogLevel(level slog.Level) HandlerOption {
	return func(p *Properties) {
		p.logLevel = level
	}
}

// buildProperties creates a new Properties with the given options.
func buildProperties(options []HandlerOption) (props *Properties) {
	props = &Properties{}
	props.logLevel = slog.LevelInfo
	for _, o := range options {
		o(props)
	}
	return
}

// NewTransactionalHandler is constructor for [TransactionalHandler].
func NewTransactionalHandler(app *newrelic.Application, tx *newrelic.Transaction, options ...HandlerOption) *TransactionalHandler {
	p := buildProperties(options)
	iw := p.innerWriter
	if iw == nil {
		iw = os.Stdout
	}
	ww := logWriter.New(iw, app)
	ww = ww.WithTransaction(tx)

	if p.innerHandlerProvider != nil {
		return &TransactionalHandler{
			handler: p.innerHandlerProvider(ww),
			tx:      tx,
			level:   p.logLevel,
		}
	}
	if p.json {
		return &TransactionalHandler{
			handler: slog.NewJSONHandler(ww, p.slogHandlerOptions),
			tx:      tx,
			level:   p.logLevel,
		}
	}
	return &TransactionalHandler{
		handler: slog.NewTextHandler(ww, p.slogHandlerOptions),
		tx:      tx,
		level:   p.logLevel,
	}
}

// attrsFromMetadata converts New Relic linking metadata to [slog.Attr].
func attrsFromMetadata(md newrelic.LinkingMetadata) []slog.Attr {
	return []slog.Attr{
		slog.String(logcontext.KeyTraceID, md.TraceID),
		slog.String(logcontext.KeySpanID, md.SpanID),
		slog.String(logcontext.KeyEntityName, md.EntityName),
		slog.String(logcontext.KeyEntityType, md.EntityType),
		slog.String(logcontext.KeyEntityGUID, md.EntityGUID),
		slog.String(logcontext.KeyHostname, md.Hostname),
	}
}
