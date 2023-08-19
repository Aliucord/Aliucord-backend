package modules

import (
	"slices"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	modules = append(modules, initTrollSupportRole)
}

func initTrollSupportRole() {
	cfg := config.TrollSupportRole
	if !cfg.Enabled {
		return
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if slices.Contains(msg.MentionRoleIDs, cfg.ID) && !slices.Contains(msg.Member.RoleIDs, cfg.ID) {
			err := s.AddRole(msg.GuildID, msg.Author.ID, cfg.ID, api.AddRoleData{AuditLogReason: "mentioned troll support role"})
			if err != nil {
				logger.Println("Failed to assign support role")
				logger.Println(err)
			}
		}
	})
}
