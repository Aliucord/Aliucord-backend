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
		Execute:   evalCommand,
	})
}

// it works pretty bad with slash commands tbh, maybe add minimal command handler for normal messages?
func evalCommand(ev *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
	code := findOption(d, "code").String()
	send := boolOrDefault(d, "send", true)

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
