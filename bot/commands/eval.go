package commands

import (
	"fmt"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
	"strings"
)

func init() {
	addCommand(&Command{
		Name:        "eval",
		Aliases:     []string{},
		Description: "Evaluate go code",
		Usage:       "<code>",
		RequiredArgCount: 1,
		OwnerOnly:   true,
		ModOnly:     false,
		Callback:    evalCommand,
	})
}

func evalCommand(ctx *CommandContext) (*discord.Message, error) {
	e := core.Import(env.NewEnv())
	e.Define("s", s)
	e.Define("msg", ctx.Message)
	e.Define("ctx", ctx)
	ret, err := vm.Execute(e, nil, strings.Join(ctx.Args, " "))
	if err != nil {
		return ctx.ReplyNoMentions("ERROR:```go\n" + err.Error() + "```")
	}
	retStr := strings.ReplaceAll(fmt.Sprint(ret), "```", "`​`​`")
	if len(retStr) > 1991 { // 2000 - (code block start + end + new line)
		retStr = retStr[:1990] + "…"
	}

	return ctx.ReplyNoMentions("```go\n" + retStr + "```")
}
