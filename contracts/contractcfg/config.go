package contractcfg

import (
	"time"
)

var DebugTest bool = false

var DefaultConfig = Config{
	//LogLevel:               logging.DEBUG,
	ContractFileSystemPath: "./chaincodes",
	ContractAddress:        "127.0.0.1:12345",
	ContractExecutetimeout: time.Duration(20) * time.Second,
	ContractDeploytimeout:  time.Duration(180) * time.Second,
	VmEndpoint:             "unix:///var/run/docker.sock",
	ContractBuilder:        "palletone/goimg",
	SysContract:            map[string]string{"deposit_syscc": "true", "sample_syscc": "true", "createToken_sycc": "true"},
}

type Config struct {
	//LogLevel               logging.Level
	ContractFileSystemPath string
	ContractAddress        string
	ContractExecutetimeout time.Duration
	ContractDeploytimeout  time.Duration
	VmEndpoint             string
	ContractBuilder        string
	SysContract            map[string]string
	//vm.docker.attachStdout
}

var contractCfg Config

func SetConfig(cfg *Config) {
	if cfg != nil {
		contractCfg = *cfg
	} else {
		contractCfg = DefaultConfig
	}
}

func GetConfig() *Config {
	if contractCfg.ContractFileSystemPath == "" || contractCfg.VmEndpoint == "" {
		contractCfg = DefaultConfig
	}
	return &contractCfg
}
