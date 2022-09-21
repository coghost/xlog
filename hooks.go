package xlog

import (
	"runtime"
	"strings"

	"github.com/rs/zerolog"
)

const (
	skip    = 2
	funcKey = "caller_fn"
)

type Hooks func(l zerolog.Logger) zerolog.Logger

var funcName zerolog.HookFunc = func(e *zerolog.Event, _ zerolog.Level, _ string) {
	pc, _, _, ok := runtime.Caller(zerolog.CallerSkipFrameCount + skip)
	if !ok {
		return
	}

	fn := runtime.FuncForPC(pc).Name()
	lst := strings.Split(fn, ".")

	n := 1
	if len(lst) < n {
		fn = lst[0]
	} else {
		fn = strings.Join(lst[len(lst)-n:], ".")
	}

	e.Str(funcKey, fn)
}

func WithFuncNameHook() Hooks {
	return func(l zerolog.Logger) zerolog.Logger {
		return l.Hook(funcName)
	}
}
