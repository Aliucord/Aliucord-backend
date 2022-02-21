package common

import "github.com/diamondburned/arikawa/v3/discord"

type (
	ToggleableModule struct {
		Enabled bool
	}

	Config struct {
		Bot            *BotConfig
		Database       *DatabaseConfig
		UpdateTracker  *UpdateTrackerConfig
		MaxDownloadVer int
		MinDownloadVer int
		Mirrors        map[int]string
		DisableAptoide []int
		Port           string
		Origin         string
	}

	BotConfig struct {
		ToggleableModule

		Token               string
		OwnerIDs            []discord.UserID
		RoleIDs             *RoleIDsConfig
		CommandsPrefix      string
		OwnerCommandsPrefix string
		Starboard           *StarboardConfig
		AutoPublish         bool
		TrollSupportRole    *TrollSupportRoleConfig
		VoiceTextChatLocker *VoiceTextChatLockerConfig
		AntiNitroScam       bool
		NormalizeNicknames  bool
		AutoReplyConfig     *AutoReplyConfig
	}

	RoleIDsConfig struct {
		ModRole         discord.RoleID
		SupportMuted    discord.RoleID
		PrdMuted        discord.RoleID
		DevMuted        discord.RoleID
		ReactionMuted   discord.RoleID
		AttachmentMuted discord.RoleID
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

	DatabaseConfig struct {
		Addr     string
		User     string
		Password string
		DB       string
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

	AutoReplyConfig struct {
		ToggleableModule

		IgnoredRoles []discord.RoleID

		PRD         discord.ChannelID
		PluginsList discord.ChannelID
		NewPlugins  discord.ChannelID

		SupportCategory discord.ChannelID
	}
)
