package configure

import (
	"os"
	"os/user"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/consensus/consensusconfig"
	"github.com/palletone/go-palletone/dag/coredata"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/pan/downloader"
)

// DefaultConfig contains default settings for use on the Ethereum main net.
var DefaultConfig = &Config{
	SyncMode: downloader.FullSync,

	ChainId:   1,
	Consensus: consensusconfig.DefaultConfig,
	Dag:       dagconfig.DefaultConfig,
	Log:       &log.DefaultConfig,
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
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Ethereum main net block is used.
	Genesis *coredata.Genesis `toml:",omitempty"`

	// Protocol options
	ChainId  uint64 // Network ID to use for selecting peers to connect to
	SyncMode downloader.SyncMode

	// DAG options
	Dag dagconfig.Config
	//Log config
	Log *log.Config
	// Consensus options
	Consensus consensusconfig.Config
}
