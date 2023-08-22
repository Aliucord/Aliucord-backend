package modules

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Aliucord/Aliucord-backend/common"
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
		if e.MessageID.Time().Unix() > time.Now().Unix()-30*24*60*60 {
			starboardMutex.Lock()
			processReaction(e.ChannelID, e.MessageID, e.Emoji, e.UserID)
			starboardMutex.Unlock()
		}
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
	if emoji.Name != "⭐" || isChannelIgnored(channelID) {
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
	idStr := msg.ID.String()
	for _, m := range messages {
		embedsLen := len(m.Embeds)
		if m.Author.ID == s.Ready().User.ID &&
			embedsLen > 0 && m.Embeds[embedsLen-1].Footer != nil &&
			strings.HasSuffix(m.Embeds[embedsLen-1].Footer.Text, idStr) {
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
			data := api.EditMessageData{Content: option.NewNullableString(content)}
			e := generateMessageEmbed(msg, false)
			ref := msg.ReferencedMessage
			if ref == nil {
				data.Embeds = &[]discord.Embed{e}
			} else {
				data.Embeds = &[]discord.Embed{generateMessageEmbed(msg.ReferencedMessage, true), e}
				components := *starboardMsg.Components[0].(*discord.ActionRowComponent)
				if len(components) == 1 {
					data.Components = &discord.ContainerComponents{&discord.ActionRowComponent{
						components[0], generateJumpToRef(ref.URL()),
					}}
				}
			}
			s.EditMessageComplex(config.Starboard.Channel, starboardMsg.ID, data)
		}
	} else {
		components := discord.ActionRowComponent{&discord.ButtonComponent{
			Label: "Jump",
			Style: discord.LinkButtonStyle(msg.URL()),
		}}
		data := api.SendMessageData{Content: content, Components: discord.ContainerComponents{&components}}
		e := generateMessageEmbed(msg, false)
		ref := msg.ReferencedMessage
		if ref == nil {
			data.Embeds = []discord.Embed{e}
		} else {
			data.Embeds = []discord.Embed{generateMessageEmbed(ref, true), e}
			components = append(components, generateJumpToRef(ref.URL()))
		}
		s.SendMessageComplex(config.Starboard.Channel, data)
	}
}

func generateJumpToRef(url string) *discord.ButtonComponent {
	return &discord.ButtonComponent{
		Label: "Jump to referenced message",
		Style: discord.LinkButtonStyle(url),
	}
}

func generateMessageEmbed(msg *discord.Message, reply bool) discord.Embed {
	tag := msg.Author.DisplayOrTag()
	e := discord.Embed{
		Author: &discord.EmbedAuthor{
			Name: common.Ternary(reply, "Replying to "+tag, tag),
			Icon: msg.Author.AvatarURL(),
		},
		Description: msg.Content,
		Footer: &discord.EmbedFooter{
			Text: "ID: " + msg.ID.String(),
		},
		Timestamp: msg.Timestamp,
		Color:     common.Ternary(reply, discord.Color(0), 16777130),
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

func isChannelIgnored(id discord.ChannelID) bool {
	for _, cid := range config.Starboard.Ignore {
		if cid == id {
			return true
		}
	}
	return false
}
