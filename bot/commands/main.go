package commands

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

var (
	s             *state.State
	config        *common.BotConfig
	logger        *common.ExtendedLogger
	commandsMap   = make(map[string]*Command)
	commandsCount = 0
	prefixRegex   *regexp.Regexp
	botUser       *discord.User
)

func InitCommands(botLogger *common.ExtendedLogger, botConfig *common.BotConfig, state *state.State, me *discord.User) {
	s = state
	logger = botLogger
	config = botConfig
	botUser = me

	prefixRegex = regexp.MustCompile("^(<@!?" + me.ID.String() + ">|" + regexp.QuoteMeta(config.CommandsPrefix) + ")\\s*")

	initModCommands() // Requires config to be initialised, init() is called too early

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		prefix := prefixRegex.FindString(msg.Content)
		if prefix == "" {
			return
		}
		args := strings.Fields(msg.Content[len(prefix):])
		if len(args) == 0 {
			return
		}

		command := commandsMap[strings.ToLower(args[0])]
		if command == nil {
			return
		}
		if command.OwnerOnly && !common.HasUser(config.OwnerIDs, msg.Author.ID) {
			return
		}
		if command.ModOnly && !common.HasRole(msg.Member.RoleIDs, config.RoleIDs.ModRole) {
			return
		}

		ctx := CommandContext{
			Message: msg,
			Args:    args[1:],
			Prefix:  prefix,
		}
		if command.RequiredArgCount > len(ctx.Args) {
			_, _ = ctx.Reply("Too few arguments. Expected " + strconv.Itoa(command.RequiredArgCount) + ", got " + strconv.Itoa(len(ctx.Args)))
			return
		}

		_, err := command.Callback(&ctx)
		if err != nil {
			_, _ = ctx.Reply("Something went wrong, sorry :(")
			logger.Printf("Error while running command %s with args %s", command.Name, strings.Join(args, ","))
			logger.Println(err)
		}
	})

	logger.Printf("Loaded %d commands\n", commandsCount)
}

type Command struct {
	Name             string
	Aliases          []string
	RequiredArgCount int
	Description      string
	Usage            string
	ModOnly          bool
	OwnerOnly        bool
	Callback         func(*CommandContext) (*discord.Message, error)
}

type CommandContext struct {
	Message *gateway.MessageCreateEvent
	Args    []string
	Prefix  string
}

func (ctx *CommandContext) Reply(content string) (*discord.Message, error) {
	return s.SendTextReply(ctx.Message.ChannelID, content, ctx.Message.ID)
}

func (ctx *CommandContext) ReplyEmbed(content string, embeds ...discord.Embed) (*discord.Message, error) {
	return s.SendMessageReply(ctx.Message.ChannelID, content, ctx.Message.ID, embeds...)
}

func (ctx *CommandContext) ReplyNoMentions(content string) (*discord.Message, error) {
	return s.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Content:         content,
		Reference:       &discord.MessageReference{MessageID: ctx.Message.ID},
		AllowedMentions: &api.AllowedMentions{RepliedUser: option.False},
	})
}

func addCommand(cmd *Command) {
	commandsCount++
	commandsMap[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		commandsMap[alias] = cmd
	}
}
