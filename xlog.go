package xlog

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/fsutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type LogOpts struct {
	cfg *XLogCfg
	wr  []io.Writer

	level zerolog.Level

	timestampFn func() time.Time
}

type LogOptFunc func(o *LogOpts)

func bindLogOpts(opt *LogOpts, opts ...LogOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithCfg(c *XLogCfg) LogOptFunc {
	return func(o *LogOpts) {
		o.cfg = c
	}
}

func WithWr(wr io.Writer) LogOptFunc {
	return func(o *LogOpts) {
		o.wr = append(o.wr, wr)
	}
}

func WithTimestampFunc(fn func() time.Time) LogOptFunc {
	return func(o *LogOpts) {
		o.timestampFn = fn
	}
}

func WithLevel(level zerolog.Level) LogOptFunc {
	return func(o *LogOpts) {
		o.level = level
	}
}

func LocalFn() time.Time {
	d := time.Now()
	return d
}

func UtcFn() time.Time {
	d := time.Now().UTC()
	return d
}

func InitLog(opts ...LogOptFunc) {
	cf := NewXLogCfg()

	opt := LogOpts{cfg: cf, timestampFn: UtcFn, level: zerolog.DebugLevel}
	bindLogOpts(&opt, opts...)

	lc := opt.cfg
	wr := opt.wr

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z"
	zerolog.TimestampFunc = opt.timestampFn

	setLog(lc.Level)

	lcf := rotateConfig{
		ConsoleLoggingEnabled: lc.LogToConsole,
		EncodeLogsAsJson:      lc.AsJson,
		FileLoggingEnabled:    lc.LogToFile,
		Directory:             fsutil.ExpandPath(lc.SaveToDir),
		Filename:              lc.FileName,
		MaxSize:               lc.MaxSize,
		MaxBackups:            lc.MaxBackups,
		MaxAge:                lc.MaxAge,
	}

	refineFileCaller()

	writers := newWriters(lcf, lc.Caller, lc.DefaultCaller)
	if len(wr) != 0 {
		writers = append(writers, wr...)
	}

	mw := io.MultiWriter(writers...)
	lg := configure(lcf, lc.Caller, mw)
	log.Logger = *lg.Logger
}

func setLog(level int) {
	l := zerolog.Level(level)
	if l > zerolog.Disabled {
		l = zerolog.TraceLevel
	}
	zerolog.SetGlobalLevel(l)
}

func refineFileCaller() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		lst := strings.Split(file, "/")
		file = lst[len(lst)-1]
		return file + ":" + strconv.Itoa(line)
	}
}

func AppendHooks(hks ...Hooks) {
	for _, hk := range hks {
		log.Logger = hk(log.Logger)
	}
}
