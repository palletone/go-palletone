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
	"errors"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

type UtxoDatabase struct {
	db ptndb.Database
}

func NewUtxoDatabase(db ptndb.Database) *UtxoDatabase {
	return &UtxoDatabase{db: db}
}

type UtxoDb interface {
	GetPrefix(prefix []byte) map[string][]byte
	GetUtxoByIndex(indexKey []byte) ([]byte, error)
	GetUtxoEntry(key []byte) (*modules.Utxo, error)
	GetAddrOutput(addr string) ([]modules.Output, error)
	GetAddrOutpoints(addr string) ([]modules.OutPoint, error)
	GetAddrUtxos(addr string) ([]modules.Utxo, error)
	GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error)
	SaveUtxoSnapshot(index modules.ChainIndex) error

	SaveUtxoEntity(key []byte, utxo *modules.Utxo) error
	SaveUtxoEntities(key []byte, utxos *[]modules.Utxo) error
	SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error
	DeleteUtxo(key []byte) error
	GetUtxoEntities(index modules.ChainIndex) (*[]modules.Utxo, error)
}

// ###################### SAVE IMPL START ######################

func (utxodb *UtxoDatabase) SaveUtxoEntity(key []byte, utxo *modules.Utxo) error {
	return StoreBytes(utxodb.db, key, utxo)
}

//@Yiran
func (utxodb *UtxoDatabase) SaveUtxoEntities(key []byte, utxos *[]modules.Utxo) error {

	return StoreBytes(utxodb.db, key, utxos)
}

// key: outpoint_prefix + addr + outpoint's hash
func (utxodb *UtxoDatabase) SaveUtxoOutpoint(key []byte, outpoint *modules.OutPoint) error {
	return StoreBytes(utxodb.db, key, outpoint)
}

// SaveUtxoView to update the utxo set in the database based on the provided utxo view.
func (utxodb *UtxoDatabase) SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error {
	for outpoint, utxo := range view {
		// No need to update the database if the utxo was not modified.
		if utxo == nil || utxo.IsModified() {
			continue
		} else {
			key := outpoint.ToKey()
			address, _ := tokenengine.GetAddressFromScript(utxo.PkScript[:])
			// Remove the utxo if it is spent
			if utxo.IsSpent() {
				err := utxodb.db.Delete(key)
				if err != nil {
					return err
				}
				// delete index , key  outpoint .
				outpoint_key := append(AddrOutPoint_Prefix, address.Bytes()...)
				utxodb.db.Delete(append(outpoint_key, outpoint.Hash().Bytes()...))

				continue
			} else {
				val, err := rlp.EncodeToBytes(utxo)
				if err != nil {
					return err
				}
				if err := utxodb.db.Put(key, val); err != nil {
					return err
				} else { // save utxoindex and  addr and key
					outpoint_key := append(AddrOutPoint_Prefix, address.Bytes()...)
					utxodb.SaveUtxoOutpoint(append(outpoint_key, outpoint.Hash().Bytes()...), &outpoint)
				}
			}
		}
	}
	return nil
}

func (utxodb *UtxoDatabase) DeleteUtxo(key []byte) error {
	return utxodb.db.Delete(key)
}

const UTXOSNAPSHOT_PREFIX = "us"

//@Yiran
func (utxodb *UtxoDatabase) SaveUtxoSnapshot(index modules.ChainIndex) error {
	//0. examine wrong calling
	if index.Index%modules.TERMINTERVAL != 0 {
		return errors.New("SaveUtxoSnapshot must wait until last term period end")
	}
	//1. get all utxo
	utxos, err := utxodb.GetAllUtxos()
	if err != nil {
		return ErrorLogHandler(err, "utxodb.GetAllUtxos")
	}
	PTNutxos := make([]modules.Utxo, 0)
	for _, utxo := range utxos {
		if utxo.Asset.AssetId == modules.PTNCOIN {
			PTNutxos = append(PTNutxos, *utxo)
		}
	}
	//2. store utxo
	key := KeyConnector([]byte(UTXOSNAPSHOT_PREFIX), ConvertBytes(index))
	return utxodb.SaveUtxoEntities(key, &PTNutxos)
}

//func (utxodb *UtxoDatabase) GetUtxoSnapshot(index []byte) error {
//
//}

// ###################### SAVE IMPL END ######################

// ###################### GET IMPL START ######################
//  dbFetchUtxoEntry
func (utxodb *UtxoDatabase) GetUtxoEntry(key []byte) (*modules.Utxo, error) {
	utxo := new(modules.Utxo)
	data, err := utxodb.db.Get(key)
	if err != nil {
		log.Error("get utxo entry failed,================================== ", "error", err)
		return nil, err
	}

	if err := rlp.DecodeBytes(data, &utxo); err != nil {
		return nil, err
	}
	return utxo, nil
}

//@Yiran get utxo snapshot from db
func (utxodb *UtxoDatabase) GetUtxoEntities(index modules.ChainIndex) (*[]modules.Utxo, error) {
	utxos := make([]modules.Utxo, 0)
	key := KeyConnector([]byte(UTXOSNAPSHOT_PREFIX), ConvertBytes(index))
	data, err := utxodb.db.Get(key)
	if err != nil {
		return nil, err
	}
	if err := rlp.DecodeBytes(data, utxos); err != nil {
		return nil, err
	}
	return &utxos, nil
}

func (utxodb *UtxoDatabase) GetUtxoByIndex(indexKey []byte) ([]byte, error) {
	return utxodb.db.Get(indexKey)
}

func (db *UtxoDatabase) GetAddrOutput(addr string) ([]modules.Output, error) {

	data := db.GetPrefix(append(AddrOutput_Prefix, []byte(addr)...))
	outputs := make([]modules.Output, 0)
	for _, b := range data {
		out := new(modules.Output)
		if err := rlp.DecodeBytes(b, out); err == nil {
			outputs = append(outputs, *out)
		}
	}
	return outputs, nil
}
func (db *UtxoDatabase) GetAddrOutpoints(addr string) ([]modules.OutPoint, error) {
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	data := db.GetPrefix(append(AddrOutPoint_Prefix, address.Bytes()...))
	outpoints := make([]modules.OutPoint, 0)
	for _, b := range data {
		out := new(modules.OutPoint)
		if err := rlp.DecodeBytes(b, out); err == nil {
			outpoints = append(outpoints, *out)
		}
	}
	return outpoints, nil
}

func (db *UtxoDatabase) GetAddrUtxos(addr string) ([]modules.Utxo, error) {
	allutxos := make([]modules.Utxo, 0)
	outpoints, err := db.GetAddrOutpoints(addr)
	if err != nil {
		return nil, err
	}
	for _, out := range outpoints {
		if utxo, err := db.GetUtxoEntry(out.ToKey()); err == nil {
			allutxos = append(allutxos, *utxo)
		}
	}
	return allutxos, nil
}
func (db *UtxoDatabase) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	view := make(map[modules.OutPoint]*modules.Utxo, 0)

	items := db.GetPrefix(modules.UTXO_PREFIX)
	var err error
	for key, itme := range items {
		utxo := new(modules.Utxo)
		// outpint := new(modules.OutPoint)
		if err = rlp.DecodeBytes(itme, utxo); err == nil {
			outpoint := modules.KeyToOutpoint([]byte(key))
			view[*outpoint] = utxo
		}
	}

	return view, err
}

// get prefix: return maps
func (db *UtxoDatabase) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}

// ###################### GET IMPL END ######################
