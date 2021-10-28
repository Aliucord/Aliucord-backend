package modules

import (
	"fmt"
	"strings"
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	modules = append(modules, initStarboard)
}

var starboardEmotes = map[int]string{
	0:  "⭐",
	4:  "🌟",
	6:  "💫",
	10: "✨",
}

func getStarboardEmote(count int) (ret string) {
	for i, e := range starboardEmotes {
		if count >= i {
			ret = e
		}
	}
	return
}

var starboardMutex = &sync.Mutex{}

func initStarboard() {
	if !config.Starboard.Enabled {
		return
	}

	s.AddHandler(func(e *gateway.MessageReactionAddEvent) {
		starboardMutex.Lock()
		processReaction(e.ChannelID, e.MessageID, e.Emoji, e.UserID)
		starboardMutex.Unlock()
	})
	s.AddHandler(func(e *gateway.MessageReactionRemoveEvent) {
		starboardMutex.Lock()
		processReaction(e.ChannelID, e.MessageID, e.Emoji, e.UserID)
		starboardMutex.Unlock()
	})
	s.AddHandler(func(e *gateway.MessageReactionRemoveAllEvent) {
		starboardMutex.Lock()
		processStarCount(&discord.Message{ID: e.MessageID}, 0)
		starboardMutex.Unlock()
	})
	s.AddHandler(func(e *gateway.MessageDeleteEvent) {
		starboardMutex.Lock()
		processStarCount(&discord.Message{ID: e.ID}, 0)
		starboardMutex.Unlock()
	})
}

func processReaction(channelID discord.ChannelID, msgID discord.MessageID, emoji discord.Emoji, userID discord.UserID) {
	if emoji.Name != "⭐" || IsChannelIgnored(channelID) {
		return
	}

	msg, err := s.Message(channelID, msgID)
	if err != nil {
		return
	}
	if msg.Author.ID == userID {
		s.DeleteUserReaction(channelID, msgID, userID, discord.APIEmoji(emoji.Name))
		return
	}

	// message doesn't have any valid content
	if msg.Content == "" && len(msg.Attachments) == 0 {
		return
	}

	count := 0
	for _, r := range msg.Reactions {
		if r.Emoji.Name == emoji.Name {
			count = r.Count
			break
		}
	}

	processStarCount(msg, count)
}

func processStarCount(msg *discord.Message, count int) {
	messages, err := s.Messages(config.Starboard.Channel, uint(s.MaxMessages()))
	if err != nil {
		logger.Println(err)
		return
	}

	var starboardMsg *discord.Message
	for _, m := range messages {
		if m.Author.ID == s.Ready().User.ID &&
			len(m.Embeds) == 1 &&
			strings.HasSuffix(m.Embeds[0].Footer.Text, msg.ID.String()) {
			starboardMsg = &m
			break
		}
	}

	if count < config.Starboard.Min {
		if starboardMsg != nil {
			s.DeleteMessage(config.Starboard.Channel, starboardMsg.ID, "no enough stars")
		}
		return
	}

	content := fmt.Sprintf("%s %d | <#%s>", getStarboardEmote(count), count, msg.ChannelID)
	if starboardMsg != nil {
		if starboardMsg.Content != content {
			embed := generateMessageEmbed(msg)
			if len(starboardMsg.Components) == 0 { // old message without jump button
				components := generateJumpButton(msg.URL())
				s.EditMessageComplex(config.Starboard.Channel, starboardMsg.ID, api.EditMessageData{
					Content:    option.NewNullableString(content),
					Embeds:     &[]discord.Embed{embed},
					Components: &components,
				})
			} else {
				s.EditMessage(config.Starboard.Channel, starboardMsg.ID, content, embed)
			}
		}
	} else {
		s.SendMessageComplex(config.Starboard.Channel, api.SendMessageData{
			Content:    content,
			Embeds:     []discord.Embed{generateMessageEmbed(msg)},
			Components: generateJumpButton(msg.URL()),
		})
	}
}

func generateJumpButton(url string) []discord.Component {
	return []discord.Component{&discord.ActionRowComponent{
		Components: []discord.Component{&discord.ButtonComponent{
			Label: "Jump",
			URL:   url,
			Style: discord.LinkButton,
		}},
	}}
}

func generateMessageEmbed(msg *discord.Message) discord.Embed {
	e := discord.Embed{
		Author: &discord.EmbedAuthor{
			Name: msg.Author.Tag(),
			Icon: msg.Author.AvatarURL(),
		},
		Description: msg.Content,
		Footer: &discord.EmbedFooter{
			Text: "ID: " + msg.ID.String(),
		},
		Timestamp: msg.Timestamp,
		Color:     16777130,
	}

	attachments := ""
	for _, a := range msg.Attachments {
		if a.Width != 0 && strings.HasPrefix(a.ContentType, "image") && e.Image == nil {
			e.Image = &discord.EmbedImage{URL: a.URL}
		} else {
			attachments += fmt.Sprintf("[%s](%s)\n", a.Filename, a.URL)
		}
	}
	if e.Image == nil {
		for _, embed := range msg.Embeds {
			if embed.Type == discord.ImageEmbed {
				e.Image = &discord.EmbedImage{URL: embed.URL}
				break
			}
		}
	}

	if attachments != "" {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Attachments",
			Value: attachments,
		})
	}

	return e
}

func IsChannelIgnored(id discord.ChannelID) bool {
	for _, cid := range config.Starboard.Ignore {
		if cid == id {
			return true
		}
	}
	return false
}
