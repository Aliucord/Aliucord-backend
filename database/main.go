package database

import (
	"context"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/go-pg/pg/v10"
)

var (
	DB     *pg.DB
	logger = common.NewLogger("[database]")
)

func InitDB(config *common.DatabaseConfig) {
	DB = pg.Connect(&pg.Options{
		Addr:     config.Addr,
		User:     config.User,
		Password: config.Password,
		Database: config.DB,
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
