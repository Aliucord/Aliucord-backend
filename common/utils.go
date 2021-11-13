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

func (logger *ExtendedLogger) PanicIfErr(err error) {
	if err != nil {
		logger.Panic(err)
	}
}
