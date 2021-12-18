package modules

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/handler"
)

func init() {
	modules = append(modules, initVoiceTextChatLocker)
}

func initVoiceTextChatLocker() {
	cfg := config.VoiceTextChatLocker
	if !cfg.Enabled {
		return
	}

	s.PreHandler = handler.New()
	s.PreHandler.AddSyncHandler(func(state *gateway.VoiceStateUpdateEvent) {
		if state.ChannelID == cfg.Voice {
			logger.LogIfErr(s.EditChannelPermission(cfg.Text, discord.Snowflake(state.UserID), api.EditChannelPermissionData{
				Type:           discord.OverwriteMember,
				Allow:          discord.PermissionSendMessages,
				AuditLogReason: "user joined to voice channel",
			}))
		} else {
			oldState, err := s.VoiceState(state.GuildID, state.UserID)
			if err == nil && oldState.ChannelID == cfg.Voice {
				logger.LogIfErr(s.DeleteChannelPermission(cfg.Text, discord.Snowflake(state.UserID), "user left from voice channel"))
			}
		}
	})
}
