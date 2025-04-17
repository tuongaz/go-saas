package main

import (
	"context"
	"net/http"

	"github.com/tuongaz/go-saas/core"
)

func main() {
	app, err := core.New()
	if err != nil {
		panic(err)
	}

	if err := app.Start(); err != nil {
		panic(err)
	}

	app.OnBeforeServe().Add(func(ctx context.Context, e *core.OnBeforeServeEvent) error {
		app.PublicRoute("/", func(r core.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello World"))
			})
		})

		return nil
	})

}
