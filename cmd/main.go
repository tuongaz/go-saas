package main

import (
	"github.com/tuongaz/go-saas/core"
	"github.com/tuongaz/go-saas/pkg/log"
)

func main() {
	app, err := core.New()
	if err != nil {
		log.Panic("failed to start new app", err)
	}

	if err := app.Start(); err != nil {
		log.Panic("failed to run app", err)
	}
}
