package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "ping",
			Description: "Ping!",
		},
		Execute: pingCommand,
	})
}

func pingCommand(e *gateway.InteractionCreateEvent, _ *discord.CommandInteraction) error {
	err := reply(e, "Pong!")
	if err != nil {
		return err
	}

	msg, err := s.InteractionResponse(e.AppID, e.Token)
	if err != nil {
		return err
	}

	return editReply(e, "Pong! "+msg.Timestamp.Time().Sub(e.ID.Time()).String())
}
