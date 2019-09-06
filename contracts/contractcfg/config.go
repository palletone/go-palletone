package contractcfg

import (
	"fmt"
	"github.com/palletone/go-palletone/configure"
	"time"
)

var DebugTest bool = false
var GptnVersion = fmt.Sprintf("%d.%d.%d", configure.VersionMajor, configure.VersionMinor, configure.VersionPatch)

const (
	Goimg     = "palletone/goimg"
	Javaimg   = "palletone/javaimg"
	Nodejsimg = "palletone/nodejsimg"
)

var DefaultConfig = Config{
	//LogLevel:               logging.DEBUG,
	//ContractFileSystemPath: "./chaincodes",
	//IsJury:                 false,
	ContractAddress:        "127.0.0.1:12345",
	ContractExecutetimeout: time.Duration(20) * time.Second,
	ContractDeploytimeout:  time.Duration(180) * time.Second,
	//CommonBuilder:          "palletone/goimg",
	//GolangBuilder:          "palletone/goimg",
	//JavaBuilder:            "palletone/javaimg",
	//NodejsBuilder:          "palletone/nodejsimg",
	VmEndpoint:  "unix:///var/run/docker.sock",
	SysContract: map[string]string{"deposit_syscc": "true", "sample_syscc": "true", "createToken_sycc": "true"},
}

type Config struct {
	//IsJury                 bool          //是否是jury节点
	ContractAddress        string        //节点ip地址
	ContractExecutetimeout time.Duration //合约调用执行时间
	ContractDeploytimeout  time.Duration //合约部署执行时间
	//CommonBuilder          string        //公共基础经镜像
	//GolangBuilder          string        //Golang基础镜像
	//JavaBuilder            string        //Java基础镜像
	//NodejsBuilder          string        //Nodejs基础镜像
	VmEndpoint  string //与docker服务连接协议
	SysContract map[string]string
	//LogLevel               logging.Level
	//ContractFileSystemPath string
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
	//if contractCfg.ContractFileSystemPath == "" || contractCfg.VmEndpoint == "" {
	//	contractCfg = DefaultConfig
	//}
	return &contractCfg
}
