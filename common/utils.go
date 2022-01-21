package common

import (
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/valyala/fasthttp"
)

type ExtendedLogger struct {
	*log.Logger
}

func NewLogger(prefix string) *ExtendedLogger {
	return &ExtendedLogger{Logger: log.New(os.Stdout, prefix+" ", log.LstdFlags)}
}

func (logger *ExtendedLogger) LogIfErr(err error) {
	if err != nil {
		logger.Println(err)
	}
}

func (logger *ExtendedLogger) PanicIfErr(err error) {
	if err != nil {
		logger.Panic(err)
	}
}

// please go add generics faster

func HasRole(roles []discord.RoleID, role discord.RoleID) bool {
	for _, id := range roles {
		if id == role {
			return true
		}
	}
	return false
}

func HasUser(users []discord.UserID, user discord.UserID) bool {
	for _, id := range users {
		if id == user {
			return true
		}
	}
	return false
}

func FailRequest(ctx *fasthttp.RequestCtx, code int) {
	ctx.SetStatusCode(code)
	_, _ = ctx.WriteString(fasthttp.StatusMessage(code))
}