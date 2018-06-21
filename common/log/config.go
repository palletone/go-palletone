package log

var DefaultConfig = Config{
	LoggerPath: "./log/out.log",
	LoggerLvl:  "INFO",
	IsDebug:    true,
	ErrPath:    "./log/err.log",
}

// key := strings.ToLower(typ.Name()) 大写统一转小写
type Config struct {
	// logger
	LoggerPath string
	ErrPath    string
	LoggerLvl  string
	IsDebug    bool
}
