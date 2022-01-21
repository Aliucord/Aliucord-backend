package database

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
)

type Mute struct {
	UserID  discord.UserID  `bun:",notnull"`
	RoleID  discord.RoleID  `bun:",notnull"`
	GuildID discord.GuildID `bun:",notnull"`
	Reason  string          `bun:",notnull"`
	EndDate int64
}

type Donor struct {
	UserID  discord.UserID `bun:",notnull"`
	SubTier int            `bun:",notnull"`
	CustomRoleID discord.RoleID
	CustomRole2ID discord.RoleID
}

func createSchema() error {
	ctx := context.Background()
	models := []interface{}{
		(*Mute)(nil),
	}

	for _, model := range models {
		if _, err := DB.NewCreateTable().IfNotExists().Model(model).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
