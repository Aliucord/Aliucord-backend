package bot

import (
	"fmt"
	"strings"

	_ "github.com/Aliucord/Aliucord-backend/bot/anko-packages"
	_ "github.com/mattn/anko/packages"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
)

func evalCommand(msg *gateway.MessageCreateEvent, args []string) {
	e := core.Import(env.NewEnv())
	e.Define("s", s)
	e.Define("msg", msg)
	ret, err := vm.Execute(e, nil, strings.Join(args, " "))
	if err != nil {
		sendReply(msg.ChannelID, "ERROR:```go\n"+err.Error()+"```", msg.ID)
		return
	}
	retStr := strings.ReplaceAll(fmt.Sprint(ret), "```", "`​`​`")
	if len(retStr) > 1991 { // 2000 - (code block start + end + new line)
		retStr = retStr[:1990] + "…"
	}
	sendReply(msg.ChannelID, "```go\n"+retStr+"```", msg.ID)
}
