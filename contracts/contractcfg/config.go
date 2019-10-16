package contractcfg

import (
	"fmt"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/contracts/comm"
	"time"
)

var DebugTest bool = false
var GptnVersion = fmt.Sprintf("%d.%d.%d", configure.VersionMajor, configure.VersionMinor, configure.VersionPatch)

const (
	Goimg     = "palletone/goimg"
	Javaimg   = "palletone/javaimg"
	Nodejsimg = "palletone/nodejsimg"
	defaultListenerPort = ":12345"
	defaultListenerIp = "0.0.0.0"
	defaultVmEndpoint = "unix:///var/run/docker.sock")

type Config struct {
	ContractAddress        string        //Jury节点ip地址
	ContractExecutetimeout time.Duration //合约调用执行时间
	ContractDeploytimeout  time.Duration //合约部署执行时间
	VmEndpoint  string //与docker服务连接协议
	SysContract map[string]string
}

func NewContractConfig() *Config {
	config :=  &Config{
		ContractAddress:        defaultListenerIp + defaultListenerPort,
		ContractExecutetimeout: time.Duration(20) * time.Second,
		ContractDeploytimeout:  time.Duration(180) * time.Second,
		VmEndpoint:             defaultVmEndpoint,
		SysContract:            map[string]string{"deposit_syscc": "true", "sample_syscc": "true", "createToken_sycc": "true"},
	}
	//  获取本地ip
	ip := comm.GetInternalIp()
	if ip != "" {
		config.ContractAddress = ip +defaultListenerPort
	}
	return config
}

var config =  &Config{}

func SetConfig(cfg *Config) {
	if cfg != nil {
		config = cfg
	} else {
		config = NewContractConfig()
	}
}

func GetConfig() *Config {
	return config
}
