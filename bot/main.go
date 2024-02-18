package bot

import (
	"context"

	"github.com/Aliucord/Aliucord-backend/bot/commands"
	"github.com/Aliucord/Aliucord-backend/bot/modules"
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

var (
	config *common.BotConfig
	s      *state.State

	logger = common.NewLogger("[bot]")
)

func StartBot(cfg *common.Config) {
	config = cfg.Bot
	config.ApkCacheDir = cfg.ApkCacheDir
	config.Origin = cfg.Origin

	s = state.New("Bot " + config.Token)
	modules.InitAllModules(logger, config, s)

	s.AddIntents(gateway.IntentDirectMessages |
		gateway.IntentGuilds |
		gateway.IntentGuildMembers |
		gateway.IntentGuildVoiceStates |
		gateway.IntentGuildMessages |
		gateway.IntentGuildMessageReactions,
	)

	logger.PanicIfErr(s.Open(context.Background()))

	me, err := s.Me()
	if err != nil {
		logger.Println(err)
	} else {
		logger.Println("Started as", me.Tag())
		commands.InitCommands(logger, config, s)
	}
}

func StopBot() {
	s.Close()
}
