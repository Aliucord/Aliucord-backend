package commands

import (
	"strconv"

	"github.com/Aliucord/Aliucord-backend/bot/modules"
	"github.com/diamondburned/arikawa/v3/discord"
)

func init() {
	addCommand(&Command{
		Name:             "normalize",
		Aliases:          []string{"normalize", "asciify", "uncancer", "uncancerify"},
		Description:      "Normalize usernames (Replace special characters)",
		Usage:            "<everyone | Member1 Member2...>",
		RequiredArgCount: 1,
		OwnerOnly:        false,
		ModOnly:          true,
		Callback:         normalizeCommand,
	})
}

func normalizeCommand(ctx *CommandContext) (*discord.Message, error) {
	if ctx.Args[0] == "everyone" {
		counter := 0
		members, err := s.Members(ctx.Message.GuildID)
		if err != nil {
			return nil, err
		}
		_, _ = ctx.Reply("Working on it...")
		for _, member := range members {
			if modules.NormalizeNickname(ctx.Message.GuildID, member.User.ID, member.Nick) {
				counter++
			}
		}
		return ctx.Reply("Done! Normalized " + strconv.Itoa(counter) + " members.")
	} else if len(ctx.Message.Mentions) > 0 {
		counter := 0
		for _, mention := range ctx.Message.Mentions {
			if mention.ID != botUser.ID {
				if modules.NormalizeNickname(ctx.Message.GuildID, mention.User.ID, mention.Member.Nick) {
					counter++
				}
			}
		}
		return s.SendTextReply(ctx.Message.ChannelID, "Done! Normalized "+strconv.Itoa(counter)+" members.", ctx.Message.ID)
	} else {
		return ctx.Reply("Tag someone!")
	}
}
