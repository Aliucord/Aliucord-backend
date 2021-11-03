package bot

import (
	"context"
	"strings"

	"github.com/Aliucord/Aliucord-backend/bot/modules"
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

var (
	config *common.BotConfig
	s      *state.State

	logger = common.NewLogger("[bot]")
)

func StartBot(cfg *common.Config) {
	config = cfg.Bot

	var err error
	s, err = state.New("Bot " + config.Token)
	if err != nil {
		logger.Fatal("Session failed", err)
	}
	s.Gateway.ErrorLog = func(err error) {
		logger.Println("Error:", err)
	}

	modules.InitAllModules(logger, config, s)

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if msg.Author.ID != cfg.OwnerID || !strings.HasPrefix(msg.Content, config.OwnerCommandsPrefix) {
			return
		}

		args := strings.Split(strings.TrimPrefix(msg.Content, config.OwnerCommandsPrefix), " ")
		if args[0] == "eval" && len(args) > 1 {
			evalCommand(msg, args[1:])
		} else if args[0] == "normalize" {
			if len(msg.Mentions) == 0 {
				sendReply(msg.ChannelID, "Mention someone!", msg.ID)
			} else {
				for _, mention := range msg.Mentions {
					modules.NormalizeNickname(msg.GuildID, mention.ID, modules.NickOrUsername(mention.Member.Nick, mention.Username))
				}
				sendReply(msg.ChannelID, "Done!", msg.ID)
			}
		}
	})

	s.Gateway.AddIntents(gateway.IntentDirectMessages |
		gateway.IntentGuilds |
		gateway.IntentGuildMembers |
		gateway.IntentGuildVoiceStates |
		gateway.IntentGuildMessages |
		gateway.IntentGuildMessageReactions,
	)

	if err = s.Open(context.Background()); err != nil {
		logger.Fatal(err)
	}

	me, err := s.Me()
	if err != nil {
		logger.Println(err)
	} else {
		logger.Println("Started as", me.Tag())
	}
}

func StopBot() {
	s.Close()
}

func sendReply(cID discord.ChannelID, content string, id discord.MessageID) (*discord.Message, error) {
	return s.SendMessageComplex(cID, api.SendMessageData{
		Content:         content,
		Reference:       &discord.MessageReference{MessageID: id},
		AllowedMentions: &api.AllowedMentions{RepliedUser: option.False},
	})
}
