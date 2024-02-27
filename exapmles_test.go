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

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
	http.Handle(newrelic.WrapHandle(nr, "/introduce", httpHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_withCustomMiddleware() {
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

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tx := newrelic.FromContext(ctx)
			logHandler := altnrslog.NewTransactionalHandler(nr, tx)
			logger := slog.New(logHandler)
			ctx, err := altnrslog.StoreToContext(ctx, logger)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	http.Handle(newrelic.WrapHandle(nr, "/introduce", middleware(httpHandler)))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExampleNewTransactionalHandler() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	tx := app.StartTransaction("ExampleNewTransactionalHandler")
	if err != nil {
		panic(err)
	}

	txHandler := altnrslog.NewTransactionalHandler(app, tx)
	slog.New(txHandler)
}

func ExampleNewTransactionalHandler_withInnerHandlerProvider() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	tx := app.StartTransaction("ExampleNewTransactionalHandler_withInnerHandlerProvider")
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

func ExampleNewTransactionalHandler_withInnerWriter() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	tx := app.StartTransaction("ExampleNewTransactionalHandler_withInnerWriter")
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	txHandler := altnrslog.NewTransactionalHandler(app, tx, altnrslog.WithInnerWriter(mw))

	slog.New(txHandler)
}

func ExampleNewTransactionalHandler_withSlogHandlerSpecify() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	tx := app.StartTransaction("ExampleNewTransactionalHandler_withSlogHandlerSpecify")
	if err != nil {
		panic(err)
	}

	txHandler := altnrslog.NewTransactionalHandler(app, tx,
		altnrslog.WithSlogHandlerSpecify(true, &slog.HandlerOptions{
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				return slog.Attr{
					Key:   fmt.Sprintf("replaced.%s.%s", groups[0], a.Key),
					Value: a.Value,
				}
			},
		}))

	slog.New(txHandler)
}
