package commands

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

func init() {
	addCommand(&Command{
		Name:        "ping",
		Aliases:     []string{},
		Description: "Ping!",
		Usage:       "",
		RequiredArgCount: 0,
		ModOnly:     false,
		OwnerOnly:   false,
		Callback:    pingCommand,
	})
}

func pingCommand(ctx *CommandContext) (*discord.Message, error) {
	return ctx.Reply("Pong!")
}
