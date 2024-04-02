package main

import (
	"database/sql"
	"log"

	"github.com/go_backend_misc/api"
	db "github.com/go_backend_misc/db/sqlc"
	"github.com/go_backend_misc/util"

	// required to connect to DB
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect with db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)
	if err := server.Start(config.ServerAddress); err != nil {
		log.Fatal("cannot start server:", err)
	}
}
