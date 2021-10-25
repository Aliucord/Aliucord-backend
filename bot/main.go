package bot

import (
	"context"
	"errors"
	"strings"

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

	initStarboard()
	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if config.AutoPublish {
			channel, _ := s.Channel(msg.ChannelID)
			if channel.Type == discord.GuildNews {
				_, err := s.CrosspostMessage(msg.ChannelID, msg.ID)
				if err != nil {
					logger.Printf(
						"Failed to crosspost message:\nChannel: %s | Message: %v | Error:\n%v",
						channel.Name, msg.ID, err,
					)
				}
			}
		}

		if msg.Author.ID != cfg.OwnerID || !strings.HasPrefix(msg.Content, config.OwnerCommandsPrefix) {
			return
		}

		args := strings.Split(strings.TrimPrefix(msg.Content, config.OwnerCommandsPrefix), " ")
		if args[0] == "eval" && len(args) > 1 {
			evalCommand(msg, args[1:])
		}
	})

	s.Gateway.AddIntents(gateway.IntentDirectMessages)
	s.Gateway.AddIntents(gateway.IntentGuildMessages)
	s.Gateway.AddIntents(gateway.IntentGuildMessageReactions)

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

func CrosspostMessage(cID discord.ChannelID, id discord.MessageID) (err error) {
	if s == nil {
		err = errors.New("session is not initialized")
	} else {
		_, err = s.CrosspostMessage(cID, id)
	}
	return
}
