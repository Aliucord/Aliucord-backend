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

func Find[E any](s []E, filter func(*E) bool) *E {
	for _, e := range s {
		el := &e
		if filter(el) {
			return el
		}
	}
	return nil
}

func SliceTransform[E, R any](s []E, transform func(E) R) []R {
	l := len(s)
	ret := make([]R, l, l)
	for i, e := range s {
		ret[i] = transform(e)
	}
	return ret
}

func MapTransform[K comparable, V, R any](m map[K]V, transform func(K, V) R) (ret []R) {
	for k, v := range m {
		ret = append(ret, transform(k, v))
	}
	return
}

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
