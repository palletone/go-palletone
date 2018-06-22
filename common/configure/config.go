package configure

import (
	"os"
	"os/user"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/consensus/consensusconfig"
	"github.com/palletone/go-palletone/core/gen/genconfig"
	"github.com/palletone/go-palletone/core/node/nodeconfig"
	"github.com/palletone/go-palletone/dag/dagconfig"
)

// DefaultConfig contains default settings for use on the Ethereum main net.
var DefaultConfig = &Config{

	Consensus: &consensusconfig.DefaultConfig,
	Dag:       &dagconfig.DefaultConfig,
	Log:       &log.DefaultConfig,
	Genesis:   &genconfig.DefaultConfig,
	Node:      &nodeconfig.DefaultConfig,
}

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		if user, err := user.Current(); err == nil {
			home = user.HomeDir
		}
	}
}

//go get github.com/fjl/gencodec
//gencodec -type Config -formats toml -out gen_config.go

type Config struct {
	//Abount genesis, include token,chain and state
	Genesis *genconfig.Config

	// DAG options
	Dag *dagconfig.Config
	//Log config
	Log *log.Config
	// Consensus options
	Consensus *consensusconfig.Config
	//Node config
	Node *nodeconfig.Config
}
