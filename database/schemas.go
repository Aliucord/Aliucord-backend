package database

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/go-pg/pg/v10/orm"
)

type Mute struct {
	UserID  discord.UserID `pg:",notnull"`
	RoleID  discord.RoleID `pg:",notnull"`
	GuildID discord.GuildID `pg:",notnull"`
	Reason  string `pg:",notnull"`
	EndDate int64
}

func createSchema() error {
	models := []interface{}{
		(*Mute)(nil),
	}

	for _, model := range models {
		if err := DB.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			return err
		}
	}
	return nil
}
