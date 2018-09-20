package contractcfg

import (
	"time"
)

var DefaultConfig = Config{
	//LogLevel:               logging.DEBUG,
	ContractFileSystemPath: "./chaincodes",
	Address:                "127.0.0.1:12345",
	ContractExecutetimeout: time.Duration(20) * time.Second,
	VmEndpoint:             "unix:///var/run/docker.sock",
	ContractBuilder:        "palletimg",
	SysContract:            map[string]string{"sample_syscc": "true"},
}

type Config struct {
	//LogLevel               logging.Level
	ContractFileSystemPath string
	Address                string
	ContractExecutetimeout time.Duration
	VmEndpoint             string
	ContractBuilder        string
	SysContract            map[string]string
}

var contractCfg Config
func SetConfig(cfg *Config)  {
	if cfg != nil {
		contractCfg = *cfg
	}else {
		contractCfg = DefaultConfig
	}
}

func GetConfig() *Config{
	return &contractCfg
}

