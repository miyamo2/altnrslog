# altnrslog

[![Go Reference](https://pkg.go.dev/badge/github.com/miyamo2/altnrslog.svg)](https://pkg.go.dev/github.com/miyamo2/altnrslog)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/miyamo2/altnrslog?logo=go&style=flat-square)](https://img.shields.io/github/go-mod/go-version/miyamo2/altnrslog?logo=go&style=flat-square)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/miyamo2/altnrslog?style=flat-square)](https://img.shields.io/github/v/release/miyamo2/altnrslog?style=flat-square)
[![codecov](https://codecov.io/gh/miyamo2/altnrslog/graph/badge.svg?token=GLLLYODW45)](https://codecov.io/gh/miyamo2/altnrslog)
[![GitHub License](https://img.shields.io/github/license/miyamo2/altnrslog?style=flat-square&color=blue)](https://img.shields.io/github/license/miyamo2/altnrslog?style=flat-square&color=blue)

altnrslog is an alternative library for [_Logs in Context_](https://docs.newrelic.com/docs/logs/logs-context/logs-in-context/) with `log/slog`.

altnrslog can also transfer `slog.Attr`.

## Getting started

### Installation

```sh
go get github.com/miyamo2/altnrslog
```

### Simple Usage

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/miyamo2/altnrslog"
	"github.com/newrelic/go-agent/v3/newrelic"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type IntroduceRequest struct {
	Name string `json:"name"`
}

func main() {
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

		var req IntroduceRequest
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

```

## Contributing

Feel free to open PR or an Issue.
