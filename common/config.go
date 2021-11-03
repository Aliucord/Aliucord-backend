package common

import "github.com/diamondburned/arikawa/v3/discord"

type (
	ToggleableModule struct {
		Enabled bool
	}

	Config struct {
		Bot            *BotConfig
		UpdateTracker  *UpdateTrackerConfig
		MaxDownloadVer int
		MinDownloadVer int
		Mirrors        map[int]string
		Port           string
		Origin         string
		OwnerID        discord.UserID
	}

	BotConfig struct {
		ToggleableModule

		Token               string
		OwnerCommandsPrefix string
		Starboard           *StarboardConfig
		AutoPublish         bool
		TrollSupportRole    *TrollSupportRoleConfig
		VoiceTextChatLocker *VoiceTextChatLockerConfig
		AntiSelfbot         bool
		NormalizeNicknames  bool
	}

	StarboardConfig struct {
		ToggleableModule

		Channel discord.ChannelID
		Ignore  []discord.ChannelID
		Min     int
	}

	TrollSupportRoleConfig struct {
		ToggleableModule

		ID discord.RoleID
	}

	VoiceTextChatLockerConfig struct {
		ToggleableModule

		Voice discord.ChannelID
		Text  discord.ChannelID
	}

	UpdateTrackerConfig struct {
		ToggleableModule

		Cache             string
		IgnoreFirstUpdate bool
		DiscordJADX       *DiscordJADXConfig
		Webhook           *UpdateWebhookConfig

		GooglePlay map[string]GooglePlayChannelConfig
	}

	UpdateWebhookConfig struct {
		ToggleableModule

		ID    discord.WebhookID
		Token string
	}

	GooglePlayChannelConfig struct {
		Email    string
		AASToken string
		JADX     bool
		Webhook  bool
	}

	DiscordJADXConfig struct {
		ToggleableModule

		AutoPush bool
		WorkDir  string
		RepoBase string
	}
)
