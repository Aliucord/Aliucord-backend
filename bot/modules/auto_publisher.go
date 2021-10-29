package modules

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	modules = append(modules, initAutoPublisher)
}

func initAutoPublisher() {
	if !config.AutoPublish {
		return
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		channel, err := s.Channel(msg.ChannelID)
		if err != nil {
			logger.Printf("Failed to get channel\n%v\n", err)
		} else if channel.Type == discord.GuildNews {
			_, err = s.CrosspostMessage(msg.ChannelID, msg.ID)
			if err != nil {
				logger.Printf(
					"Failed to crosspost message:\nChannel: %s | Message: %v | Error:\n%v",
					channel.Name, msg.ID, err,
				)
			}
		}
	})
}
