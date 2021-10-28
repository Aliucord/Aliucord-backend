package modules

import (
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/state"
)

var (
	config *common.BotConfig
	s      *state.State

	logger *common.ExtendedLogger

	modules []func()
)

func InitAllModules(botLogger *common.ExtendedLogger, cfg *common.BotConfig, state *state.State) {
	logger = botLogger
	config = cfg
	s = state

	for _, initModule := range modules {
		initModule()
	}
}
