package main

import (
	"log"

	"github.com/wesleybruno/golang-monolito/internal/env"
)

func main() {

	env.LoadConfig()

	cfg := config{
		addr: env.Config.ApiPort,
	}

	app := &application{
		config: cfg,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

}
