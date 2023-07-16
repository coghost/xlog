package xlog

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/gookit/goutil/fsutil"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Configuration for logging
type rotateConfig struct {
	// Enable console logging
	ConsoleLoggingEnabled bool

	// EncodeLogsAsJson makes the log framework log JSON
	EncodeLogsAsJson bool
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int

	NoColor bool
}

type Logger struct {
	*zerolog.Logger
}

// configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/service-xyz/service-xyz.log and
// will be rolled according to configuration set.
func newWriters(config rotateConfig, withCaller bool, defaultCaller bool) []io.Writer {
	var writers []io.Writer
	var Green = color.New(color.FgGreen, color.Bold).SprintFunc()
	color.NoColor = config.NoColor

	var filename = func(i interface{}) string {
		arr := strings.Split(i.(string), "/")
		return Green(arr[len(arr)-1])
	}
	var cw = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    config.NoColor,
	}
	if withCaller {
		cw.FormatCaller = filename
	}

	if config.ConsoleLoggingEnabled {
		if defaultCaller {
			writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05", NoColor: config.NoColor})
		} else {
			writers = append(writers, cw)
		}
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	return writers
}

func configure(config rotateConfig, mode bool, mw io.Writer) *Logger {
	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	var logger zerolog.Logger
	if mode {
		logger = zerolog.New(mw).With().Caller().Timestamp().Logger()
	} else {
		logger = zerolog.New(mw).With().Timestamp().Logger()
	}

	logger.Trace().
		Bool("fileLogging", config.FileLoggingEnabled).
		Bool("jsonLogOutput", config.EncodeLogsAsJson).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile(config rotateConfig) io.Writer {
	name := path.Join(config.Directory, config.Filename)
	fsutil.MustCreateFile(name, os.ModePerm, os.ModePerm)

	return &lumberjack.Logger{
		Filename:   name,
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}
