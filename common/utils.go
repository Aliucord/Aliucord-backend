package common

import (
	"log"
	"os"
	"unicode"
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

func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func ToTitle(s string) string {
	return string(unicode.ToTitle(rune(s[0]))) + s[1:]
}
