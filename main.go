package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ai-marketing/ai-marketing-server/config"
	"github.com/ai-marketing/ai-marketing-server/databases"
	"github.com/ai-marketing/ai-marketing-server/server"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env variables")
	}

	conf := config.GetConfig()
	db := databases.NewPostgresDatabase(conf.Database)

	srv := server.NewGinServer(conf, db.GetConnection())

	srv.Start()
}
