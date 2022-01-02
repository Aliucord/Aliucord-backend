package modules

import (
	"strings"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	modules = append(modules, initAntiZip)
}

func initAntiZip() {
	if !config.AntiZip {
		return
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if len(msg.Attachments) == 0 || msg.Member == nil {
			return
		}

		for _, role := range msg.Member.RoleIDs {
			if common.HasRole(config.RoleIDs.IgnoredRoles, role) {
				return
			}
		}

		for _, attachment := range msg.Attachments {
			if strings.HasSuffix(attachment.Filename, ".zip") {
				err := s.DeleteMessage(msg.ChannelID, msg.ID, "Sent zip")
				if err != nil {
					logger.Println(err)
				}
				return
			}
		}
	})
}
