package common

import (
	"log"
	"os"
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

// because go is missing ternary operator:

func InlineIfBool(cond, a, b bool) bool {
	if cond {
		return a
	}
	return b
}

func InlineIfStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func InlineIfInt(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}
