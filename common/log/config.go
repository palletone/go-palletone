package log

var DefaultConfig = Config{
	OutputPaths:      []string{"stdout", "./log/out.log"},
	ErrorOutputPaths: []string{"stderr", "./log/err.log"},
	LoggerLvl:        "INFO",
	Encoding:         "console", // json
	Development:      true,
}

type Config struct {
	// logger
	OutputPaths      []string
	ErrorOutputPaths []string
	LoggerLvl        string // log level
	Encoding         string // encoding
	Development      bool
}
