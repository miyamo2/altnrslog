package altnrslog_test

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/miyamo2/altnrslog"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type ExampleRequest struct {
	Name string `json:"name"`
}

func Example() {
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

		var req ExampleRequest
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
