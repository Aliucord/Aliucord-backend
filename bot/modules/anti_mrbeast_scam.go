package modules

import (
	"github.com/diamondburned/arikawa/v3/gateway"
	"strconv"
	"strings"
)

func init() {
	modules = append(modules, initAntiMrBeastScam)
}

// Deletes all messages that have exactly 4 image attachments, with each attachment's
// name being an increasing number.
func initAntiMrBeastScam() {
	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if msg.Author.Bot || msg.GuildID == 0 {
			return
		}

		if len(msg.Attachments) != 4 {
			return
		}

		for i, attachment := range msg.Attachments {
			if !strings.HasPrefix(attachment.ContentType, "image/") {
				return
			}

			name := strings.Split(attachment.Filename, ".")[0]
			if strconv.Itoa(i) != name {
				return
			}
		}

		logger.LogWithCtxIfErr(
			"moderating potential MrBeast scam",
			s.DeleteMessage(msg.ChannelID, msg.ID, ""),
		)
	})
}
