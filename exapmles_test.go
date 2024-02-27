package altnrslog_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/miyamo2/altnrslog"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func Example() {
	type Request struct {
		Name string `json:"name"`
	}

	nr, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		panic(err)
	}

	http.HandleFunc(newrelic.WrapHandleFunc(nr, "/introduce", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tx := newrelic.FromContext(ctx)
		logHandler := altnrslog.NewTransactionalHandler(nr, tx)
		logger := slog.New(logHandler)

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.InfoContext(ctx, "START", slog.Group("request", slog.String("name", req.Name)))

		response := fmt.Sprintf("Hello, %s!", req.Name)
		defer logger.InfoContext(ctx, "END", slog.String("response", response))

		w.Write([]byte(response))
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExampleNewTransactionalHandler_WithInnerHandlerProvider() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	tx := app.StartTransaction("ExampleNewTransactionalHandler_WithInnerHandlerProvider")
	if err != nil {
		panic(err)
	}

	handlerOpt := slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return slog.Attr{
				Key:   fmt.Sprintf("replaced.%s.%s", groups[0], a.Key),
				Value: a.Value,
			}
		},
	}

	jsonHandlerProvider := func(w io.Writer) slog.Handler {
		return slog.NewJSONHandler(w, &handlerOpt)
	}

	txHandler := altnrslog.NewTransactionalHandler(app, tx,
		altnrslog.WithInnerHandlerProvider(jsonHandlerProvider))

	slog.New(txHandler)
}
