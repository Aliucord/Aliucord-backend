package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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

	_, err = s.EditInteractionResponse(e.AppID, e.Token, api.EditInteractionResponseData{
		Content: option.NewNullableString("Pong! " + msg.Timestamp.Time().Sub(e.ID.Time()).String()),
	})
	return err
}
