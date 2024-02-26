//go:generate mockgen -destination mock/slog_handler.go log/slog Handler
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
}

// Enabled See: [slog.Handler.Enabled]
func (h *TransactionalHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
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
	}
}

// WithGroup See: [slog.Handler.WithGroup]
func (h *TransactionalHandler) WithGroup(name string) slog.Handler {
	return &TransactionalHandler{
		h.handler.WithGroup(name),
		h.tx,
	}
}

// Properties is an options for creating a new [TransactionalHandler].
type Properties struct {
	innerWriter        io.Writer
	json               bool
	slogHandlerOptions *slog.HandlerOptions
}

// HandlerOption is a functional option for creating a new [TransactionalHandler].
type HandlerOption func(*Properties)

// WithInnerWriter specifies the [io.Writer] that wraps [logWriter.logWriter]
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

// buildProperties creates a new Properties with the given options.
func buildProperties(options []HandlerOption) *Properties {
	p := &Properties{}
	for _, o := range options {
		o(p)
	}
	return p
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

	if p.json {
		return &TransactionalHandler{
			handler: slog.NewJSONHandler(ww, p.slogHandlerOptions),
			tx:      tx,
		}
	}
	return &TransactionalHandler{
		handler: slog.NewTextHandler(ww, p.slogHandlerOptions),
		tx:      tx,
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
