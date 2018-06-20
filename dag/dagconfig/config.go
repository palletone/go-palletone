package dagconfig

var (
	SConfig Sconfig
)

var DefaultConfig = Config{
	DbPath:     "dbpath",
	DbName:     "dbname",
	LoggerPath: "./log/out.log",
	LoggerLvl:  "DEBUG",
	IsDebug:    true,
	ErrPath:    "./log/err.log",
}

// key := strings.ToLower(typ.Name()) 大写统一转小写
type Config struct {
	DbPath     string
	DbName     string
	DbUser     string
	DbPassword string
	DbCache    int
	DbHandles  int

	// cache
	CacheSource string

	//redis
	RedisAddr   string
	RedisPwd    string
	RedisPrefix string
	RedisDb     int

	// logger
	LoggerPath string
	ErrPath    string
	LoggerLvl  string
	IsDebug    bool
}

type Sconfig struct {
	Blight bool
}
