package xlog

type XLogCfg struct {
	Level  int  `ini:"level"`
	Caller bool `ini:"caller"`
	// with full caller name or not
	DefaultCaller bool `ini:"default_caller"`

	LogToConsole bool `ini:"log_to_console"`
	LogToFile    bool `ini:"log_to_file"`
	AsJson       bool `ini:"as_json"`
	MaxSize      int  `ini:"max_size"`
	MaxBackups   int  `ini:"max_backups"`
	MaxAge       int  `ini:"max_age"`

	SaveToDir string `ini:"save_to_dir"`
	FileName  string `ini:"file_name"`
}

func NewXLogCfg() *XLogCfg {
	return &XLogCfg{
		Level:         0,
		Caller:        true,
		DefaultCaller: false,
		LogToConsole:  true,
		LogToFile:     true,
		AsJson:        true,

		MaxSize:    10,
		MaxAge:     0,
		MaxBackups: 0,

		SaveToDir: "~/tmp/xkitlog",
		FileName:  "xkit_log.log",
	}
}
