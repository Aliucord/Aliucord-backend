package modules

import (
	"github.com/Aliucord/Aliucord-backend/common"
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
		if common.HasRole(msg.MentionRoleIDs, cfg.ID) && !common.HasRole(msg.Member.RoleIDs, cfg.ID) {
			err := s.AddRole(msg.GuildID, msg.Author.ID, cfg.ID, api.AddRoleData{AuditLogReason: "mentioned troll support role"})
			if err != nil {
				logger.Println("Failed to assign support role")
				logger.Println(err)
			}
		}
	})
}
