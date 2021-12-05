package database

import (
	"context"
	"os"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/go-pg/pg/v10"
)

var (
	DB     *pg.DB
	logger = common.NewLogger("[database]")
)

func init() {
	DB = pg.Connect(&pg.Options{
		Addr:     os.Getenv("POSTGRES_HOST") + ":" + os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
		OnConnect: func(ctx context.Context, cn *pg.Conn) error {
			logger.Println("Successfully connected")
			return nil
		},
	})
	if err := createSchema(); err != nil {
		logger.Println("Failed to create schema")
		logger.Panic(err)
	}
}
