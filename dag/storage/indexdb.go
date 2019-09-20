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
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

type IndexDb struct {
	db ptndb.Database
}

func NewIndexDb(db ptndb.Database) *IndexDb {
	return &IndexDb{db: db}
}

type IIndexDb interface {
	SaveAddressTxId(address common.Address, txid common.Hash) error
	GetAddressTxIds(address common.Address) ([]common.Hash, error)
	//清空AddressTxIds
	TruncateAddressTxIds() error
	SaveTokenTxId(asset *modules.Asset, txid common.Hash) error
	GetTokenTxIds(asset *modules.Asset) ([]common.Hash, error)

	SaveMainDataTxId(maindata []byte, txid common.Hash) error
	GetMainDataTxIds(maindata []byte) ([]common.Hash, error)
	SaveProofOfExistence(poe *modules.ProofOfExistence) error
	QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error)
}

func (db *IndexDb) SaveAddressTxId(address common.Address, txid common.Hash) error {
	key := append(constants.ADDR_TXID_PREFIX, address.Bytes()...)
	key = append(key, txid[:]...)
	return db.db.Put(key, txid[:])
}
func (db *IndexDb) GetAddressTxIds(address common.Address) ([]common.Hash, error) {
	prefix := append(constants.ADDR_TXID_PREFIX, address.Bytes()...)
	data := getprefix(db.db, prefix)
	result := make([]common.Hash, 0)
	for _, v := range data {
		hash := common.Hash{}
		hash.SetBytes(v)
		result = append(result, hash)
	}
	return result, nil
}

func (db *IndexDb) TruncateAddressTxIds() error {
	iter := db.db.NewIteratorWithPrefix(constants.ADDR_TXID_PREFIX)
	for iter.Next() {
		key := iter.Key()
		err := db.db.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}
func (db *IndexDb) SaveTokenTxId(asset *modules.Asset, txid common.Hash) error {
	key := append(constants.TOKEN_TXID_PREFIX, asset.Bytes()...)
	key = append(key, txid[:]...)
	return db.db.Put(key, txid[:])
}

func (db *IndexDb) GetTokenTxIds(asset *modules.Asset) ([]common.Hash, error) {
	prefix := append(constants.TOKEN_TXID_PREFIX, asset.Bytes()...)
	data := getprefix(db.db, prefix)
	result := make([]common.Hash, 0)
	for _, v := range data {
		hash := common.Hash{}
		hash.SetBytes(v)
		result = append(result, hash)
	}
	return result, nil
}

//save filehash key:IDX_MAIN_DATA_TXID   value:Txid
func (db *IndexDb) SaveMainDataTxId(filehash []byte, txid common.Hash) error {
	key := append(constants.IDX_MAIN_DATA_TXID, filehash...)
	key = append(key, []byte(txid.String())...)

	return db.db.Put(key, txid[:])
}

func (db *IndexDb) GetMainDataTxIds(filehash []byte) ([]common.Hash, error) {
	key := append(constants.IDX_MAIN_DATA_TXID, filehash...)
	data := getprefix(db.db, key)
	result := make([]common.Hash, 0)
	for _, v := range data {
		hash := common.Hash{}
		hash.SetBytes(v)
		result = append(result, hash)
	}
	return result, nil
}

func (db *IndexDb) SaveProofOfExistence(poe *modules.ProofOfExistence) error {
	if len(poe.Reference) == 0 {
		return nil
	}
	key := append(constants.IDX_REF_DATA_PREFIX, poe.Reference...)
	key = append(key, poe.TxId.Bytes()...)

	return StoreToRlpBytes(db.db, key, poe)
}

func (db *IndexDb) QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error) {
	prefix := append(constants.IDX_REF_DATA_PREFIX, ref...)
	iter := db.db.NewIteratorWithPrefix(prefix)
	result := []*modules.ProofOfExistence{}
	for iter.Next() {
		value := iter.Value()
		poe := &modules.ProofOfExistence{}
		err := rlp.DecodeBytes(value, poe)
		if err != nil {
			return nil, err
		}
		result = append(result, poe)
	}
	return result, nil
}
