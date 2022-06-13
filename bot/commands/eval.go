package commands

import (
	"fmt"
	"strings"

	_ "github.com/Aliucord/Aliucord-backend/bot/anko-packages"
	_ "github.com/mattn/anko/packages"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
)

func init() {
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "eval",
			Description: "Evaluate go code",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "code",
					Description: "Code to evaluate",
					Required:    true,
				},
				&discord.BooleanOption{
					OptionName:  "send",
					Description: "Send output (default: true)",
				},
			},
		},
		OwnerOnly: true,
		Execute: func(ev *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
			return eval(ev, d, findOption(d, "code").String(), boolOrDefault(d, "send", true))
		},
	})
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name: "Eval",
			Type: discord.MessageCommand,
		},
		OwnerOnly: true,
		Execute: func(ev *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
			var msg *discord.Message
			for _, m := range d.Resolved.Messages {
				msg = &m
				break
			}

			if msg == nil {
				return ephemeralReply(ev, "No message provided")
			}
			return eval(ev, d, msg.Content, true)
		},
	})
}

func eval(ev *gateway.InteractionCreateEvent, d *discord.CommandInteraction, code string, send bool) error {
	if strings.HasPrefix(code, "```") {
		code = code[3 : len(code)-3]
		if strings.HasPrefix(code, "go\n") {
			code = code[3:]
		}
	}

	e := core.Import(env.NewEnv())
	_ = e.Define("s", s)
	_ = e.Define("e", ev)
	_ = e.Define("d", d)
	ret, err := vm.Execute(e, nil, code)
	if err != nil {
		return replyWithFlags(
			ev,
			common.Ternary(send, 0, api.EphemeralResponse),
			"ERROR:```go\n"+err.Error()+"```",
		)
	}
	retStr := strings.ReplaceAll(fmt.Sprint(ret), "```", "`​`​`")
	if len(retStr) > 1991 { // 2000 - (code block start + end + new line)
		retStr = retStr[:1990] + "…"
	}

	return replyWithFlags(ev, common.Ternary(send, 0, api.EphemeralResponse), "```go\n"+retStr+"```")
}
