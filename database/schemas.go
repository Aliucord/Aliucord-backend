package database

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/go-pg/pg/v10/orm"
)

type MuteSchema struct {
	UserID  discord.UserID `pg:",notnull"`
	RoleID  discord.RoleID `pg:",notnull"`
	EndDate int64
	Reason  string
}

func createSchema() error {
	models := []interface{}{
		(*MuteSchema)(nil),
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

func createMute() (orm.Result, error) {
	mute := MuteSchema{
		UserID:  0,
		RoleID:  0,
		EndDate: 0,
		Reason:  "",
	}

	return DB.Model(mute).Insert()
}