package database

import (
	"database/sql"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	DB     *bun.DB
	logger = common.NewLogger("[database]")
)

func InitDB(config *common.DatabaseConfig) {
	DB = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(config.Addr),
		pgdriver.WithUser(config.User),
		pgdriver.WithPassword(config.Password),
		pgdriver.WithDatabase(config.DB),
		pgdriver.WithTLSConfig(nil),
	)), pgdialect.New())
	DB.SetMaxOpenConns(1)

	if err := createSchema(); err != nil {
		logger.Println("Failed to create schema")
		logger.Panic(err)
	}
}
