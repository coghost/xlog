package zlog

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogOpts struct {
	devEnv bool
}

type LogOptFunc func(o *LogOpts)

func bindLogOpts(opt *LogOpts, opts ...LogOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithDevEnv(b bool) LogOptFunc {
	return func(o *LogOpts) {
		o.devEnv = b
	}
}

// MustNewZapLogger create zap logger, if prod is empty or prod[0] is false, use debug logger.
// else return prod logger.
func MustNewZapLogger(opts ...LogOptFunc) *zap.Logger {
	opt := &LogOpts{devEnv: false}
	bindLogOpts(opt, opts...)

	// writeSyncer shared in dev/prod
	writeSyncer := getLumerjackLogWriter()

	level := zapcore.InfoLevel
	fsEncoder := getProdEncoder()
	csEncoder := fsEncoder

	if opt.devEnv {
		fsEncoder = getDevEncoder(false)
		csEncoder = getDevEncoder(true)
		level = zapcore.DebugLevel
	}

	coreConsole := zapcore.NewCore(csEncoder, zapcore.AddSync(os.Stdout), level)
	core := zapcore.NewCore(fsEncoder, writeSyncer, level)
	coreTee := zapcore.NewTee(core, coreConsole)
	logger := zap.New(coreTee, zap.AddCaller())

	if opt.devEnv {
		ReplaceGlobalToShowLogZapL(logger)
	}

	return logger
}

func getProdEncoder() zapcore.Encoder { //nolint
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getDevEncoder(isConsole bool) zapcore.Encoder { //nolint
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if isConsole {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.ConsoleSeparator = " "
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func ReplaceGlobalToShowLogZapL(logger *zap.Logger) {
	// zap.L().Debug("global zap logger is replaced.")
	zap.ReplaceGlobals(logger)
}

func getLumerjackLogWriter() zapcore.WriteSyncer { //nolint
	const (
		backupFiles = 5
		days        = 30
		size        = 10
	)

	lumberJackLogger := &lumberjack.Logger{
		Filename:   "/tmp/zlog.log",
		MaxSize:    size,
		MaxBackups: backupFiles,
		MaxAge:     days,
		Compress:   false,
	}

	return zapcore.AddSync(lumberJackLogger)
}
