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

package common

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto/sha3"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/asset"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"time"
	"reflect"
)

func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func RHashStr(x interface{}) string {
	x_byte, err := json.Marshal(x)
	if err != nil {
		return ""
	}
	s256 := sha256.New()
	s256.Write(x_byte)
	return fmt.Sprintf("%x", s256.Sum(nil))

}

//  last unit
func CurrentUnit() *modules.Unit {
	return &modules.Unit{}
}

// get unit
func GetUnit(hash *common.Hash, index modules.ChainIndex) *modules.Unit {
	unit_bytes, err := storage.Get(append(storage.UNIT_PREFIX, hash.Bytes()...))
	if err != nil {
		return nil
	}
	unit := new(modules.Unit)
	if err := json.Unmarshal(unit_bytes, &unit); err == nil {
		return unit
	}
	return nil
}

/**
生成创世单元，需要传入创世单元的配置信息以及coinbase交易
generate genesis unit, need genesis unit configure fields and transactions list
*/
func NewGenesisUnit(txs modules.Transactions) (*modules.Unit, error) {
	gUnit := modules.Unit{Gasprice:0, Gasused:0, Creationdate:time.Now()}

	// genesis unit asset id
	gAssetID := asset.NewAsset()

	// genesis unit height
	chainIndex := modules.ChainIndex{AssetID: gAssetID, IsMain: true, Index: 0}

	// transactions merkle root
	root := core.DeriveSha(txs)

	// generate genesis unit header
	header := modules.Header{
		AssetIDs: []modules.IDType36{gAssetID},
		GasLimit: 0,
		GasUsed:  0,
		Number:   chainIndex,
		Root:     root,
	}

	gUnit.UnitHeader = &header
	// copy txs
	if len(txs) > 0 {
		gUnit.Txs = make([]*modules.Transaction, len(txs))
		for i, pTx := range txs {
			tx := modules.Transaction{
				AccountNonce: pTx.AccountNonce,
				TxHash:       pTx.TxHash,
				From:         pTx.From,
				Excutiontime: pTx.Excutiontime,
				Memery:       pTx.Memery,
				CreationDate: pTx.CreationDate,
				TxFee:        pTx.TxFee,
			}
			if len(pTx.TxMessages) > 0 {
				tx.TxMessages = make([]modules.Message, len(pTx.TxMessages))
				for j := 0; j < len(pTx.TxMessages); j++ {
					tx.TxMessages[j] = pTx.TxMessages[j]
				}
			}
			gUnit.Txs[i] = &tx
		}
	}
	return &gUnit, nil
}

/**
为创世单元生成ConfigPayload
To generate config payload for genesis unit
 */
func GenGenesisConfigPayload(genesisConf *core.Genesis) (modules.ConfigPayload, error) {
	var confPay modules.ConfigPayload

	confPay.ConfigSet = make(map[string]interface{})
	confPay.ConfigSet["version"] = genesisConf.Version
	confPay.ConfigSet["initialActiveMediators"] = genesisConf.InitialActiveMediators
	confPay.ConfigSet["InitialMediatorCandidates"] = genesisConf.InitialMediatorCandidates

	t := reflect.TypeOf(genesisConf.SystemConfig)
	v := reflect.ValueOf(genesisConf.SystemConfig)
	for k := 0; k < t.NumField(); k++ {
		confPay.ConfigSet[t.Field(k).Name] = v.Field(k).Interface()
	}

	return confPay, nil
}
