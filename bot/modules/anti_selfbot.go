package modules

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
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
				if e.Type == discord.NormalEmbed && !strings.Contains(e.URL, "https://twitter.com/") {
					logger.LogIfErr(s.Ban(msg.GuildID, msg.Author.ID, api.BanData{
						AuditLogReason: "sent selfbot embed (" + api.AuditLogReason(msg.URL()) + ")",
					}))
					break
				}
			}
		}
	})
}
