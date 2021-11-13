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

	s, err := state.New("Bot " + config.Token); logger.PanicIfErr(err)
	s.Gateway.ErrorLog = func(err error) {
		logger.Println("Error:", err)
	}

	modules.InitAllModules(logger, config, s)

	s.Gateway.AddIntents(gateway.IntentDirectMessages |
		gateway.IntentGuilds |
		gateway.IntentGuildMembers |
		gateway.IntentGuildVoiceStates |
		gateway.IntentGuildMessages |
		gateway.IntentGuildMessageReactions,
	)

	err = s.Open(context.Background()); logger.PanicIfErr(err)

	me, err := s.Me()
	if err != nil {
		logger.Println(err)
	} else {
		logger.Println("Started as", me.Tag())
		commands.InitCommands(logger, config, s, me)
	}
}

func StopBot() {
	s.Close()
}
