package config

var (
	DConfig DagConfig
	SConfig Sconfig
)

type DagConfig struct {
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
}

type Sconfig struct {
	Blight bool
}
