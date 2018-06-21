package log

var DefaultConfig = Config{
	LoggerPath: "./log/out.log",
	LoggerLvl:  "INFO",
	IsDebug:    true,
	ErrPath:    "./log/err.log",
	Encoding:   "console", // json
}

type Config struct {
	// logger
	LoggerPath string // out path
	ErrPath    string // err path
	LoggerLvl  string // log levle
	Encoding   string // encoding
	IsDebug    bool   // is dubug
}
