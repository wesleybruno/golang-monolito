package main

import (
	"fmt"
	"log"

	"github.com/wesleybruno/golang-monolito/internal/db"
	"github.com/wesleybruno/golang-monolito/internal/env"
	"github.com/wesleybruno/golang-monolito/internal/store"
)

const version = "0.0.2"

//	@title			GoSocial API
//	@description	This is a sample of api.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {

	env.LoadConfig()

	cfg := config{
		addr:   env.Config.ApiPort,
		env:    env.Config.Env,
		apiUrl: env.Config.ApiUrl,
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
