package modules

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	modules = append(modules, initAntiSelfbot)
}

func initAntiSelfbot() {
	if !config.AntiSelfbot {
		return
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if !msg.Author.Bot {
			for _, e := range msg.Embeds {
				if e.Type == discord.NormalEmbed {
					s.Ban(msg.GuildID, msg.Author.ID, api.BanData{
						DeleteDays:     option.NewUint(1),
						AuditLogReason: "sent selfbot embed",
					})
					break
				}
			}
		}
	})
}
