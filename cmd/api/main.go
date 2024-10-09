package main

import (
	"fmt"
	"log"

	"github.com/wesleybruno/golang-monolito/internal/db"
	"github.com/wesleybruno/golang-monolito/internal/env"
	"github.com/wesleybruno/golang-monolito/internal/store"
)

const version = "0.0.1"

func main() {

	env.LoadConfig()

	cfg := config{
		addr: env.Config.ApiPort,
		env:  env.Config.Env,
		dbConfig: dbConfig{
			addr:         fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", env.Config.DbUser, env.Config.DbPassword, env.Config.DbAddress, env.Config.DbName),
			maxOpenConns: env.Config.MaxOpenConns,
			maxIdleConns: env.Config.MaxIdleConns,
			maxIdleTime:  env.Config.MaxIdleTime,
		},
	}

	db, err := db.New(cfg.dbConfig.addr, cfg.dbConfig.maxOpenConns, cfg.dbConfig.maxIdleConns, cfg.dbConfig.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("database connection pool established")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

}
