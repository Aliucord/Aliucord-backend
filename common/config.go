package common

import "github.com/diamondburned/arikawa/v3/discord"

type (
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
		Enabled             bool
		Token               string
		OwnerCommandsPrefix string
		StarboardChannel    discord.ChannelID
		StarboardIgnore     []discord.ChannelID
		StarboardMin        int
		AutoPublish         bool
		TrollSupportID      discord.RoleID
		VoiceTextChatLocker *VoiceTextChatLockerConfig
	}

	VoiceTextChatLockerConfig struct {
		Enabled bool
		Voice   discord.ChannelID
		Text    discord.ChannelID
	}

	UpdateTrackerConfig struct {
		Enabled           bool
		Cache             string
		IgnoreFirstUpdate bool
		DiscordJADX       *DiscordJADXConfig
		Webhook           *UpdateWebhookConfig

		GooglePlay map[string]GooglePlayChannelConfig
	}

	UpdateWebhookConfig struct {
		Enabled bool
		ID      discord.WebhookID
		Token   string
	}

	GooglePlayChannelConfig struct {
		Email    string
		AASToken string
		JADX     bool
		Webhook  bool
	}

	DiscordJADXConfig struct {
		Enabled  bool
		AutoPush bool
		WorkDir  string
		RepoBase string
	}
)
