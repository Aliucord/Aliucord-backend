package common

import (
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/discord"
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

// LogWithCtxIfErr logs one or more errors with context.
// Context should be simple description with -ing verb like
// "adding role" (will be logged as "Exception while adding role")
func (logger *ExtendedLogger) LogWithCtxIfErr(context string, errs ...error) {
	hasLogged := false
	for _, err := range errs {
		if err != nil {
			if !hasLogged {
				logger.Println("Exception while " + context)
				hasLogged = true
			}
			logger.Println("\t", err)
		}
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
