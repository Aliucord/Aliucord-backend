package commands

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"strconv"
	"strings"
)

func init() {
	addCommand(&Command{
		Name:             "help",
		Aliases:          []string{"h", "commands"},
		Description:      "You are here :3",
		Usage:            "[command name]",
		RequiredArgCount: 0,
		ModOnly:          false,
		OwnerOnly:        false,
		Callback:         helpCommand,
	})
}

func helpCommand(ctx *CommandContext) (*discord.Message, error) {
	// Change prefix from <@ID> to @Username if it is a mention
	if strings.Replace(strings.TrimRight(ctx.Prefix, " "), "!", "", -1) == "<@"+botUser.ID.String()+">" {
		ctx.Prefix = "@" + botUser.Username + " "
	}

	if len(ctx.Args) > 0 {
		cmd := commandsMap[ctx.Args[0]]
		if cmd == nil {
			return ctx.Reply("No such command: " + ctx.Args[0])
		}

		embed := discord.Embed{
			Title:       ctx.Prefix + cmd.Name,
			Description: "OwnerOnly: " + getEmoji(cmd.OwnerOnly) + "\nModOnly: " + getEmoji(cmd.ModOnly) + "\nRequired Args: " + strconv.Itoa(cmd.RequiredArgCount) + "\n\n" + cmd.Description,
			Fields: []discord.EmbedField{
				{
					Name:   "Aliases",
					Value:  strings.Join(cmd.Aliases, ", "),
					Inline: false,
				},
				{
					Name:   "Usage",
					Value:  "```\n" + ctx.Prefix + cmd.Name + " " + cmd.Usage + "```",
					Inline: false,
				},
			},
		}
		return ctx.ReplyEmbed("", embed)
	}

	sb := strings.Builder{}

	for key, cmd := range commandsMap {
		// Skip aliases
		if key == cmd.Name {
			_, err := sb.WriteString("`" + ctx.Prefix + cmd.Name + "`: " + cmd.Description + "\n")
			if err != nil {
				return nil, err
			}
		}
	}

	embed := discord.Embed{
		Title:       "Help",
		Description: "For more info on a command, run `" + ctx.Prefix + "help command`!\n\n" + sb.String(),
	}

	return ctx.ReplyEmbed("", embed)
}

func getEmoji(b bool) string {
	if b {
		return "✅"
	} else {
		return "❌"
	}
}
