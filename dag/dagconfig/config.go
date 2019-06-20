/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package dagconfig

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

var (
	SConfig      Sconfig
	DefaultToken = "PTN"
	// DefaultToken = "ABC+10A4QIWCI46V8MZJ2UO"
	DagConfig = DefaultConfig
)

var DefaultConfig = Config{
	DbPath: "./leveldb",
	// txpool
	UnitTxSize: 1024 * 1000 * 5, //5mb

	// utxo
	UtxoIndex: true,

	// memory unit, unit number
	MemoryUnitSize: 1280,
	// Irreversible Height
	IrreversibleHeight:           1, // 单节点memdag正常缓存区块，需要将该值设置为1
	WhetherValidateUnitSignature: false,
	GenesisHash:                  "0xeb5f66d0289ea0af68860fd5a4d1a0b38389f598ae01008433a5ca9949fcf55c",
	PartitionForkUnitHeight:      0,
	//PtnAssetHex:                  modules.CoreAsset.AssetId.String(),
	//PtnAssetId:                   modules.NewPTNAsset().AssetId[:],
	AddrTxsIndex:      false,
	Token721TxIndex:   true,
	TextFileHashIndex: false,
	GasToken:          DefaultToken,
}

//func init() {
//	if DagConfig.PtnAssetHex != "" {
//		id, _ := modules.SetIdTypeByHex(DagConfig.PtnAssetHex)
//		DagConfig.PtnAssetId = id[:]
//		// modules.PTNCOIN.SetBytes(DagConfig.PtnAssetId)
//	}
//}

// global configuration of dag modules
type Config struct {
	DbPath    string
	DbCache   int
	DbHandles int

	// cache
	CacheSource string

	//redis
	RedisAddr   string
	RedisPwd    string
	RedisPrefix string
	RedisDb     int

	// txpool
	UnitTxSize float64

	// utxo
	UtxoIndex bool

	// memory unit size, unit number
	MemoryUnitSize uint

	// Irreversible height
	IrreversibleHeight int

	// Validate unit signature, just for debug version
	WhetherValidateUnitSignature bool
	// genesis hash‘s hex
	GenesisHash             string
	PartitionForkUnitHeight int

	AddrTxsIndex    bool
	Token721TxIndex bool

	TextFileHashIndex bool

	//当前节点选择的平台币，燃料币,必须为Asset全名
	GasToken string
	gasToken modules.AssetId `toml:"-"`

	SyncPartitionTokens []string
	syncPartitionTokens []modules.AssetId `toml:"-"`
}

type Sconfig struct {
	Blight bool
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "PalletOne")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "PalletOne")
		} else {
			return filepath.Join(home, ".palletone")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func (c *Config) GetGasToken() modules.AssetId {
	if c.gasToken == modules.ZeroIdType16() {
		token, _, err := modules.String2AssetId(c.GasToken)
		if err != nil {
			log.Warn("Cannot parse node.GasToken to a correct asset, token str:" + c.GasToken + ",error: " + err.Error())
			return modules.PTNCOIN
		}
		c.gasToken = token
	}
	return c.gasToken
}

//func (c *Config) GetMainToken() modules.AssetId {
//	if c.mainToken == modules.ZeroIdType16() {
//		token, _, err := modules.String2AssetId(c.MainToken)
//		{
//			if err != nil {
//				return modules.PTNCOIN
//			}
//		}
//		c.mainToken = token
//	}
//	return c.mainToken
//}
func (c *Config) GeSyncPartitionTokens() []modules.AssetId {
	if c.syncPartitionTokens == nil {
		c.syncPartitionTokens = []modules.AssetId{}
		for _, tokenString := range c.SyncPartitionTokens {
			token, _, err := modules.String2AssetId(tokenString)
			if err != nil {
				log.Warn("Cannot parse node.SyncPartitionTokens to a correct asset, token str:" + c.GasToken)
				c.syncPartitionTokens = append(c.syncPartitionTokens, token)
			}
		}
	}
	return c.syncPartitionTokens
}
