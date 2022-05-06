package commands

import (
	"strconv"

	"github.com/Aliucord/Aliucord-backend/bot/modules"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "normalize",
			Description: "Normalize usernames (Replace special characters)",
			Options: []discord.CommandOption{
				&discord.UserOption{
					OptionName:  "user",
					Description: "user to normalize",
				},
				&discord.StringOption{
					OptionName:  "users",
					Description: "users to normalize",
				},
			},
		},
		ModOnly: true,
		Execute: normalizeCommand,
	})
}

func normalizeCommand(e *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
	userIDs := getUserOrUsersOption(d)
	if len(userIDs) == 0 {
		return ephemeralReply(e, "Provide either `user` or `users` option.")
	}

	counter := 0
	for _, id := range userIDs {
		member, err := s.Member(e.GuildID, id)
		if err == nil && modules.NormalizeNickname(e.GuildID, id, member.Nick) {
			counter++
		}
	}

	return ephemeralReply(e, "Done! Normalized "+strconv.Itoa(counter)+" members.")
}
