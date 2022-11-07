package commands

import (
	"regexp"
	"strconv"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	s             *state.State
	config        *common.BotConfig
	logger        *common.ExtendedLogger
	commandsMap   = make(map[string]*Command)
	commandsCount = 0

	idRegex = regexp.MustCompile("\\d{17,19}")
)

func InitCommands(botLogger *common.ExtendedLogger, botConfig *common.BotConfig, state *state.State) {
	s = state
	logger = botLogger
	config = botConfig

	initModCommands() // Requires config to be initialised, init() is called too early

	logger.Printf("Loaded %d commands\n", commandsCount)

	ready := s.Ready()
	commands := common.MapTransform(commandsMap, func(_ string, command *Command) api.CreateCommandData {
		return command.CreateCommandData
	})
	for _, guild := range ready.Guilds {
		guildID := guild.ID

		_, err := s.BulkOverwriteGuildCommands(ready.Application.ID, guildID, commands)
		if err == nil {
			logger.Printf("Registered commands in %d guild\n", guildID)
		} else {
			logger.Printf("Failed to register commands in %d guild (%v)\n", guildID, err)
		}
	}

	s.AddHandler(func(e *gateway.InteractionCreateEvent) {
		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			command, ok := commandsMap[d.Name]
			if !ok || command == nil {
				return
			}

			// extra checks if discord does something dumb
			if !slices.Contains(config.OwnerIDs, e.Member.User.ID) &&
				(command.OwnerOnly || command.ModOnly && !slices.Contains(e.Member.RoleIDs, config.RoleIDs.ModRole)) {
				return
			}

			if err := command.Execute(e, d); err != nil {
				content := option.NewNullableString("Something went wrong, sorry :(")
				if _, err2 := s.InteractionResponse(e.AppID, e.Token); err2 == nil {
					_, err2 = s.EditInteractionResponse(e.AppID, e.Token, api.EditInteractionResponseData{
						Content: content,
					})
					logger.LogIfErr(err2)
				} else {
					logger.LogIfErr(s.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Content: content,
							Flags:   discord.EphemeralMessage,
						},
					}))
				}

				logger.Printf("Error while running command %s\n%v\n", command.Name, err)
			}
		}
	})
}

type Command struct {
	api.CreateCommandData

	ModOnly   bool
	OwnerOnly bool
	Execute   func(e *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error
}

func getMultipleUsersOption(d *discord.CommandInteraction) []discord.UserID {
	usersOption := findOption(d, "users")
	if usersOption == nil {
		return []discord.UserID{}
	}

	ids := idRegex.FindAllString(usersOption.String(), -1)
	return common.SliceTransform(ids, func(idStr string) discord.UserID {
		id, _ := strconv.ParseUint(idStr, 10, 64)
		return discord.UserID(id)
	})
}

func getUserOrUsersOption(d *discord.CommandInteraction) []discord.UserID {
	userIDs := getMultipleUsersOption(d)
	if users := d.Resolved.Users; len(users) > 0 {
		userIDs = append(userIDs, maps.Keys(users)...)
	}
	return userIDs
}

func reply(e *gateway.InteractionCreateEvent, content string) error {
	return replyWithFlags(e, 0, content, nil)
}

func ephemeralReply(e *gateway.InteractionCreateEvent, content string) error {
	return replyWithFlags(e, discord.EphemeralMessage, content, nil)
}

func replyErr(e *gateway.InteractionCreateEvent, context string, err error) error {
	logger.Println("Err while " + context)
	logger.Println(err)
	return ephemeralReply(e, "Something went wrong: ```\n"+err.Error()+"```")
}

func replyWithFlags(e *gateway.InteractionCreateEvent, flags discord.MessageFlags, content string, embeds *[]discord.Embed) error {
	return s.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content:         option.NewNullableString(content),
			Flags:           flags,
			Embeds:          embeds,
			AllowedMentions: &api.AllowedMentions{},
		},
	})
}

func editReply(e *gateway.InteractionCreateEvent, content string) error {
	_, err := s.EditInteractionResponse(e.AppID, e.Token, api.EditInteractionResponseData{
		Content: option.NewNullableString(content),
	})
	return err
}

func findOption(d *discord.CommandInteraction, name string) *discord.CommandInteractionOption {
	return common.Find(d.Options, func(option *discord.CommandInteractionOption) bool {
		return option.Name == name
	})
}

func boolOrDefault(d *discord.CommandInteraction, name string, def bool) bool {
	boolOption := findOption(d, name)
	if boolOption == nil {
		return def
	}
	ret, err := boolOption.BoolValue()
	return common.Ternary(err == nil, ret, def)
}

func addCommand(cmd *Command) {
	commandsCount++
	commandsMap[cmd.Name] = cmd
}
