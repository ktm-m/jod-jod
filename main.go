package main

import (
	"github.com/Montheankul-K/jod-jod/config"
	"github.com/Montheankul-K/jod-jod/db"
	"github.com/Montheankul-K/jod-jod/db/migration"
	"github.com/Montheankul-K/jod-jod/server"
)

func main() {
	cfg := config.GetConfig()
	database := db.InitDatabase(cfg.Database)
	if err := migration.Migrate(database); err != nil {
		panic(err)
	}

	srv := server.InitServer(cfg, database)
	if err := srv.Start(); err != nil {
		panic(err)
	}
}
