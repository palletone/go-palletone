// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ptn

import (
	//"path/filepath"
	//"runtime"
	"time"

	//"github.com/palletone/go-palletone/consensus/consensusconfig"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/txspool"
	"github.com/palletone/go-palletone/ptn/downloader"
)

// DefaultConfig contains default settings for use on the PalletOne main net.
var DefaultConfig = Config{
	SyncMode:      downloader.FastSync,
	NetworkId:     1,
	LightServ:     10,
	LightPeers:    25,
	CorsPeers:     0,
	DatabaseCache: 768,
	TrieCache:     256,
	TrieTimeout:   5 * time.Minute,
	CryptoLib:     []byte{0, 0},

	TxPool:         txspool.DefaultTxPoolConfig,
	Dag:            dagconfig.DagConfig,
	MediatorPlugin: mediatorplugin.DefaultConfig,
	Jury:           jury.DefaultConfig,
	Contract:       contractcfg.DefaultConfig,
}

func init() {
	//home := os.Getenv("HOME")
	//if home == "" {
	//	if user, err := user.Current(); err == nil {
	//		home = user.HomeDir
	//	}
	//}
	/*would recover
	if runtime.GOOS == "windows" {
		DefaultConfig.Ethash.DatasetDir = filepath.Join(home, "AppData", "Ethash")
	} else {
		DefaultConfig.Ethash.DatasetDir = filepath.Join(home, ".ethash")
	}*/
}

//go:generate gencodec -type Config -field-override configMarshaling -formats toml -out gen_config.go

type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the PalletOne main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode
	NoPruning bool

	// Light client options
	LightServ  int `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightPeers int `toml:",omitempty"` // Maximum number of LES client peers

	// Cors client options
	//CorsServ  int `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	CorsPeers int `toml:",omitempty"` // Maximum number of LES client peers

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	TrieCache          int
	TrieTimeout        time.Duration

	// Mining-related options
	//Etherbase    common.Address `toml:",omitempty"`
	MinerThreads int    `toml:",omitempty"`
	ExtraData    []byte `toml:",omitempty"`
	CryptoLib    []byte

	// Transaction pool options
	TxPool txspool.TxPoolConfig `toml:"-"`

	// Gas Price Oracle options
	//GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool
	// DAG options
	Dag dagconfig.Config `toml:"-"`

	//jury Account
	Jury jury.Config `toml:"-"`

	//Contract config
	Contract contractcfg.Config `toml:"-"`

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// append by Albert·Gou
	MediatorPlugin mediatorplugin.Config `toml:"-"`

	//must be equal to the node.GasToken
	//TokenSubProtocol string `toml:"-"`
}
