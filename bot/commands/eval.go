package commands

import (
	"fmt"
	"strings"

	_ "github.com/Aliucord/Aliucord-backend/bot/anko-packages"
	_ "github.com/mattn/anko/packages"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
)

func init() {
	addCommand(&Command{
		Name:             "eval",
		Aliases:          []string{},
		Description:      "Evaluate go code",
		Usage:            "<code>",
		RequiredArgCount: 1,
		OwnerOnly:        true,
		ModOnly:          false,
		Callback:         evalCommand,
	})
}

func evalCommand(ctx *CommandContext) (*discord.Message, error) {
	code := strings.Join(ctx.Args, " ")
	if strings.HasPrefix(code, "```") {
		code = code[3:len(code) - 3]
		if strings.HasPrefix(code, "go\n") {
			code = code[3:]
		}
	}

	e := core.Import(env.NewEnv())
	_ = e.Define("s", s)
	_ = e.Define("msg", ctx.Message)
	_ = e.Define("ctx", ctx)
	ret, err := vm.Execute(e, nil, code)
	if err != nil {
		return ctx.ReplyNoMentions("ERROR:```go\n" + err.Error() + "```")
	}
	retStr := strings.ReplaceAll(fmt.Sprint(ret), "```", "`​`​`")
	if len(retStr) > 1991 { // 2000 - (code block start + end + new line)
		retStr = retStr[:1990] + "…"
	}

	return ctx.ReplyNoMentions("```go\n" + retStr + "```")
}
