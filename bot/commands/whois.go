package commands

import (
	"math"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name: "Whois",
			Type: discord.UserCommand,
		},
		Execute: whoisCommand,
	})

	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "whois",
			Description: "Get info on a user",
			Options: []discord.CommandOption{
				&discord.UserOption{
					OptionName:  "user",
					Description: "user",
				},
				&discord.StringOption{
					OptionName:  "id",
					Description: "id of user to lookup",
				},
			},
		},
		Execute: whoisCommand,
	})
}

func whoisCommand(e *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
	id := d.TargetID
	if id == 0 {
		for _, opt := range d.Options {
			if opt.Name == "user" || opt.Name == "id" {
				var err error
				id, err = opt.SnowflakeValue()
				if err != nil {
					return ephemeralReply(e, "That was not a valid id!!")
				}
				break
			}
		}
	}

	user, err := s.User(discord.UserID(id))
	if err != nil {
		return ephemeralReply(e, "No such user")
	}

	fields := []discord.EmbedField{
		{
			Name:   "ID",
			Value:  user.ID.String(),
			Inline: false,
		},
		{
			Name:   "Created At",
			Value:  formatTime(user.CreatedAt()),
			Inline: false,
		},
	}

	if e.GuildID != 0 {
		member, err := s.Member(e.GuildID, user.ID)
		if err == nil && member.Joined.IsValid() {
			fields = append(fields, discord.EmbedField{
				Name:   "Joined At",
				Value:  formatTime(member.Joined.Time()),
				Inline: false,
			})
		}
	}

	name := user.Tag()
	if user.Bot {
		name = "🤖 " + name
	}

	return replyWithFlags(e, discord.EphemeralMessage, user.Mention(), &[]discord.Embed{
		{
			Author: &discord.EmbedAuthor{
				Name: name,
				Icon: user.AvatarURL(),
			},
			Fields:    fields,
			Timestamp: discord.Timestamp(user.CreatedAt()),
		},
	})
}

func formatTime(t time.Time) string {
	res := t.Format("Jan 02 2006 15:04:05") + " ("

	days := int(math.Floor(float64(time.Now().Sub(t)) / float64(time.Hour*24)))
	switch days {
	case 0:
		res += "Today"
	case 1:
		res += "Yesterday"
	default:
		res += strconv.Itoa(days) + " days ago"
	}

	return res + ")"
}
