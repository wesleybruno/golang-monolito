package main

import (
	"fmt"
	"log"

	"github.com/wesleybruno/golang-monolito/internal/db"
	"github.com/wesleybruno/golang-monolito/internal/env"
	"github.com/wesleybruno/golang-monolito/internal/store"
)

func main() {

	env.LoadConfig()

	addr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", env.Config.DbUser, env.Config.DbPassword, env.Config.DbAddress, env.Config.DbName)

	conn, err := db.New(addr, env.Config.MaxOpenConns, env.Config.MaxIdleConns, env.Config.MaxIdleTime)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store)

}
