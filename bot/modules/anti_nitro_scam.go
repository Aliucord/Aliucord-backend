package modules

import (
	"context"
	"strings"
	"time"

	"github.com/Aliucord/Aliucord-backend/database"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	modules = append(modules, initAntiNitroScam)
}

var scamPhrases []database.ScamPhrase

func UpdateScamTitles() {
	logger.PanicIfErr(database.DB.NewSelect().Model(&scamPhrases).Scan(context.Background()))
}

func initAntiNitroScam() {
	if !config.AntiNitroScam {
		return
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if msg.Author.Bot || msg.GuildID == 0 {
			return
		}

		for _, e := range msg.Embeds {
			if e.Title == "" || e.URL == "" || strings.HasPrefix(e.URL, "https://discord.com") || strings.HasPrefix(e.URL, "https://discord.gift") {
				continue
			}

			normalizedTitle := Normalize(e.Title)
			for _, phrase := range scamPhrases {
				if normalizedTitle == phrase.Phrase {
					goto scam
				}
			}
			continue

		scam:
			s.SendTextReply(msg.ChannelID, "HARAM", msg.ID)
			// Ignore errors here since error indicates user has dms closed
			dm, err := s.CreatePrivateChannel(msg.Author.ID)
			if err == nil {
				_, _ = s.SendMessage(dm.ID, "Your account posted a Nitro Scam in the Aliucord Server. Thus, you have either been timed out or banned. If you secured your account (change password and fully uninstall and reinstall Discord), you may appeal this punishment over at https://github.com/Aliucord/ban-appeals/issues/new/choose.")
			}

			if e.Title != normalizedTitle {
				logger.LogWithCtxIfErr(
					"banning nitro scammer",
					s.Ban(msg.GuildID, msg.Author.ID, api.BanData{DeleteDays: option.NewUint(1), AuditLogReason: "Nitro Scam"}),
				)
				break
			}

			timestamp := discord.NewTimestamp(time.Now().Add(24 * time.Hour))
			logger.LogWithCtxIfErr(
				"moderating potential nitro scam",
				s.ModifyMember(msg.GuildID, msg.Author.ID, api.ModifyMemberData{CommunicationDisabledUntil: &timestamp, AuditLogReason: "Nitro Scam"}),
				s.DeleteMessage(msg.ChannelID, msg.ID, "Nitro Scam"),
			)
			break
		}
	})
}
