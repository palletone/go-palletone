/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

type UtxoDb struct {
	db                   ptndb.Database
	tokenEngine          tokenengine.ITokenEngine
	ADDR_OUTPOINT_PREFIX []byte
	OUTPOINT_ADDR_PREFIX []byte
	SPENT_UTXO_PREFIX    []byte
	UTXO_PREFIX          []byte
	UTXO_INDEX_PREFIX    []byte
	isRequest            bool
}

func NewUtxoDb(db ptndb.Database, tokenEngine tokenengine.ITokenEngine, isRequest bool) *UtxoDb {
	utxodb := &UtxoDb{db: db, tokenEngine: tokenEngine, isRequest: isRequest}
	if isRequest {
		utxodb.ADDR_OUTPOINT_PREFIX = constants.REQUEST_ADDR_OUTPOINT_PREFIX
		utxodb.OUTPOINT_ADDR_PREFIX = constants.REQUEST_OUTPOINT_ADDR_PREFIX
		utxodb.SPENT_UTXO_PREFIX = constants.REQUEST_STXO_PREFIX
		utxodb.UTXO_PREFIX = constants.REQUEST_UTXO_PREFIX
		utxodb.UTXO_INDEX_PREFIX = constants.REQUEST_UTXO_INDEX_PREFIX
	} else {
		utxodb.ADDR_OUTPOINT_PREFIX = constants.ADDR_OUTPOINT_PREFIX
		utxodb.OUTPOINT_ADDR_PREFIX = constants.OUTPOINT_ADDR_PREFIX
		utxodb.SPENT_UTXO_PREFIX = constants.SPENT_UTXO_PREFIX
		utxodb.UTXO_PREFIX = constants.UTXO_PREFIX
		utxodb.UTXO_INDEX_PREFIX = constants.UTXO_INDEX_PREFIX
	}
	return utxodb
}

type IUtxoDb interface {
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)

	GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error)
	GetAddrUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error)
	GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error)
	SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error
	SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error
	DeleteUtxo(outpoint *modules.OutPoint, spentTxId common.Hash, spentTime uint64) error
	IsUtxoSpent(outpoint *modules.OutPoint) (bool, error)
	GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error)
	ClearAddrUtxo(addr common.Address) error
	ClearUtxo() error

	SaveRequestAndTxHash(requestHash common.Hash, txHash common.Hash) error
	GetRequestAndTxMapping(hash common.Hash) (common.Hash, error)
}

// ###################### UTXO index for Address ######################
// key: outpoint_prefix + addr + outpoint's Bytes
func (utxodb *UtxoDb) saveUtxoOutpoint(address common.Address, outpoint *modules.OutPoint) error {
	key := append(utxodb.ADDR_OUTPOINT_PREFIX, address.Bytes()...)
	key = append(key, outpoint.Bytes()...)

	// save outpoint tofind address
	out_key := append(utxodb.OUTPOINT_ADDR_PREFIX, outpoint.ToKey(utxodb.UTXO_PREFIX)...)
	StoreToRlpBytes(utxodb.db, out_key, address.String())
	return StoreToRlpBytes(utxodb.db, key, outpoint)
}
func (utxodb *UtxoDb) batchSaveUtxoOutpoint(batch ptndb.Batch, address common.Address,
	outpoint *modules.OutPoint) error {
	key := append(utxodb.ADDR_OUTPOINT_PREFIX, address.Bytes()...)
	key = append(key, outpoint.Bytes()...)
	return StoreToRlpBytes(batch, key, outpoint)
}
func (utxodb *UtxoDb) deleteUtxoOutpoint(address common.Address, outpoint *modules.OutPoint) error {
	key := append(utxodb.ADDR_OUTPOINT_PREFIX, address.Bytes()...)
	key = append(key, outpoint.Bytes()...)
	return utxodb.db.Delete(key)
}
func (db *UtxoDb) GetAddrOutpoints(address common.Address) ([]modules.OutPoint, error) {
	data := getprefix(db.db, append(db.ADDR_OUTPOINT_PREFIX, address.Bytes()...))
	outpoints := make([]modules.OutPoint, 0)
	for _, b := range data {
		out := new(modules.OutPoint)
		if err := rlp.DecodeBytes(b, out); err == nil {
			outpoints = append(outpoints, *out)
		}
	}
	return outpoints, nil
}

// ###################### UTXO Entity ######################
func (utxodb *UtxoDb) SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error {
	key := outpoint.ToKey(utxodb.UTXO_PREFIX)
	address, _ := utxodb.tokenEngine.GetAddressFromScript(utxo.PkScript[:])
	err := StoreToRlpBytes(utxodb.db, key, utxo)
	if err != nil {
		return err
	}
	log.DebugDynamic(func() string {
		return fmt.Sprintf("DB[%s,%p,req=%t] save utxo[%s]",
			reflect.TypeOf(utxodb.db).String(), utxodb.db, utxodb.isRequest, outpoint.String())
	})
	return utxodb.saveUtxoOutpoint(address, outpoint)
}

// SaveUtxoView to update the utxo set in the database based on the provided utxo view.
func (utxodb *UtxoDb) SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error {
	batch := utxodb.db.NewBatch()
	log.Debugf("Start batch save utxo, batch count:%d", len(view))
	for outpoint, utxo := range view {
		// No need to update the database if the utxo was not modified.
		if utxo == nil {
			continue
		} else {
			key := outpoint.ToKey(utxodb.UTXO_PREFIX)
			err := StoreToRlpBytes(batch, key, utxo)
			if err != nil {
				log.Errorf("store utxo to db failed, key:[%s]", outpoint.String())
				return err
			}
			address, _ := utxodb.tokenEngine.GetAddressFromScript(utxo.PkScript[:])
			// save utxoindex and  addr and key
			item := new(modules.OutPoint)
			item.TxHash = outpoint.TxHash
			item.MessageIndex = outpoint.MessageIndex
			item.OutIndex = outpoint.OutIndex
			if err := utxodb.batchSaveUtxoOutpoint(batch, address, item); err != nil {
				log.Errorf("batch_save_utxo failed,addr[%s] , error:[%s]", address.String(), err)
			}
		}
	}

	return batch.Write()
}

// Remove the utxo
func (utxodb *UtxoDb) DeleteUtxo(outpoint *modules.OutPoint, spentTxId common.Hash, spentTime uint64) error {
	//1. get utxo
	utxo, err := utxodb.GetUtxoEntry(outpoint)
	if err != nil {
		return err
	}
	log.Debugf("DB--------DeleteUtxo----------170---delete utxo-----------------")
	key := outpoint.ToKey(utxodb.UTXO_PREFIX)
	//1 delete UTXO
	err = utxodb.db.Delete(key)
	if err != nil {
		return err
	}
	//2 create stxo
	utxodb.SaveUtxoSpent(outpoint, utxo, spentTxId, spentTime)
	//3 delete index
	address, _ := utxodb.tokenEngine.GetAddressFromScript(utxo.PkScript[:])
	utxodb.deleteUtxoOutpoint(address, outpoint)
	log.Debugf("DB[%p,req=%t] delete utxo by key:%s, spend by tx[%s]", utxodb.db, utxodb.isRequest, outpoint.String(), spentTxId.String())
	return nil
}

// ###################### SAVE IMPL END ######################

// ###################### GET IMPL START ######################
//  dbFetchUtxoEntry
func (utxodb *UtxoDb) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	if outpoint == nil {
		return nil, errors.ErrNullPoint
	}
    log.Debugf("GetUtxoEntry    try ReqUtxo------193------")
	utxo := new(modules.Utxo)
	key := outpoint.ToKey(utxodb.UTXO_PREFIX)

	err := RetrieveFromRlpBytes(utxodb.db, key, utxo)
	if err != nil {
		log.Debugf("GetUtxoEntry not utxodb.db, try ReqUtxo------200------")
		log.DebugDynamic(func() string {
			return fmt.Sprintf("DB[%s,%p,req=%t] get utxo[%s] error:%s",
				reflect.TypeOf(utxodb.db).String(), utxodb.db, utxodb.isRequest, outpoint.String(), err.Error())
		})
		if errors.IsNotFoundError(err) {
			return nil, errors.ErrUtxoNotFound
		}
		return nil, err
	}
	log.Debugf("GetUtxoEntry utxodb.db, success------210---")
	return utxo, nil
}
func (utxodb *UtxoDb) SaveUtxoSpent(outpoint *modules.OutPoint, utxo *modules.Utxo,
	spentTxId common.Hash, spentTime uint64) error {
	stxo := modules.NewStxo(utxo, spentTxId, spentTime)
	return utxodb.SaveStxoEntry(outpoint, stxo)
}
func (utxodb *UtxoDb) SaveStxoEntry(outpoint *modules.OutPoint, stxo *modules.Stxo) error {
	key := append(utxodb.SPENT_UTXO_PREFIX, outpoint.ToKey(utxodb.UTXO_PREFIX)...)
	//stxo := modules.NewStxo(utxo, spentTxId, spentTime)
	log.Debug("utxodb.db--SaveStxoEntry----221-----------------------")
	return StoreToRlpBytes(utxodb.db, key, stxo)
}
func (utxodb *UtxoDb) IsUtxoSpent(outpoint *modules.OutPoint) (bool, error) {
	key := append(utxodb.SPENT_UTXO_PREFIX, outpoint.ToKey(utxodb.UTXO_PREFIX)...)
	return utxodb.db.Has(key)
}
func (utxodb *UtxoDb) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	key := append(utxodb.SPENT_UTXO_PREFIX, outpoint.ToKey(utxodb.UTXO_PREFIX)...)
	stxo := &modules.Stxo{}
	err := RetrieveFromRlpBytes(utxodb.db, key, stxo)
	if err != nil {
		return nil, err
	}
	return stxo, nil
}

//GetAddrUtxos if asset is nil, query all Asset from address
func (db *UtxoDb) GetAddrUtxos(addr common.Address, asset *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, error) {
	allutxos := make(map[modules.OutPoint]*modules.Utxo)
	outpoints, err := db.GetAddrOutpoints(addr)
	if err != nil {
		return nil, err
	}
	for _, out := range outpoints {
		item := new(modules.OutPoint)
		item.TxHash = out.TxHash
		item.MessageIndex = out.MessageIndex
		item.OutIndex = out.OutIndex
		if utxo, err := db.GetUtxoEntry(item); err == nil {
			if asset == nil || asset.IsSimilar(utxo.Asset) {
				allutxos[out] = utxo
			}
		}
	}
	return allutxos, nil
}
func (db *UtxoDb) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	view := make(map[modules.OutPoint]*modules.Utxo)

	items := getprefix(db.db, db.UTXO_PREFIX)
	var err error
	for key, itme := range items {
		utxo := new(modules.Utxo)
		if err = rlp.DecodeBytes(itme, utxo); err == nil {

			outpoint := modules.KeyToOutpoint([]byte(key))
			view[*outpoint] = utxo

		}
	}

	return view, err
}
func (db *UtxoDb) ClearUtxo() error {
	err := clearByPrefix(db.db, db.UTXO_PREFIX)
	if err != nil {
		return err
	}
	err = clearByPrefix(db.db, db.ADDR_OUTPOINT_PREFIX)
	if err != nil {
		return err
	}
	err = clearByPrefix(db.db, db.UTXO_INDEX_PREFIX)
	if err != nil {
		return err
	}
	return nil
}
func (db *UtxoDb) ClearAddrUtxo(addr common.Address) error {
	outpoints, err := db.GetAddrOutpoints(addr)
	if err != nil {
		return err
	}
	for _, outpoint := range outpoints {
		err := db.db.Delete(outpoint.ToKey(db.UTXO_PREFIX))
		if err != nil {
			return err
		}
	}

	return nil
}
func clearByPrefix(db ptndb.Database, prefix []byte) error {
	iter := db.NewIteratorWithPrefix(prefix)
	for iter.Next() {
		key := iter.Key()
		err := db.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// ###################### GET IMPL END ######################
func (utxodb *UtxoDb) SaveRequestAndTxHash(requestHash common.Hash, txHash common.Hash) error {
	key1 := append(constants.REQUEST_AND_TX_PREFIX, requestHash[:]...)
	err := utxodb.db.Put(key1, txHash[:])
	if err != nil {
		return err
	}
	key2 := append(constants.REQUEST_AND_TX_PREFIX, txHash[:]...)
	return utxodb.db.Put(key2, requestHash[:])
}
func (utxodb *UtxoDb) GetRequestAndTxMapping(hash common.Hash) (common.Hash, error) {
	key := append(constants.REQUEST_AND_TX_PREFIX, hash[:]...)
	value, err := utxodb.db.Get(key)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(value), nil
}
