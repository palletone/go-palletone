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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
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
	//GetPrefix(prefix []byte) map[string][]byte
	//SaveIndexValue(key []byte, value interface{}) error
	//GetUtxoByIndex(idx *modules.UtxoIndex) (*modules.Utxo, error)
	//DeleteUtxoByIndex(idx *modules.UtxoIndex) error
	SaveAddressTxId(address common.Address, txid common.Hash) error
	GetAddressTxIds(address common.Address) ([]common.Hash, error)

	SaveTokenTxId(asset *modules.Asset, txid common.Hash) error
	GetTokenTxIds(asset *modules.Asset) ([]common.Hash, error)

	//GetFromAddressTxIds(addr string) ([]common.Hash, error)
	//GetTxFromAddresses(tx *modules.Transaction) ([]string, error)

	SaveMainDataTxId(maindata []byte, txid common.Hash) error
	GetMainDataTxIds(maindata []byte) ([]common.Hash, error)
}

// ###################### SAVE IMPL START ######################
//func (idxdb *IndexDb) SaveIndexValue(key []byte, value interface{}) error {
//	return StoreToRlpBytes(idxdb.db, key, value)
//}

// ###################### SAVE IMPL END ######################
// ###################### GET IMPL START ######################
//func (idxdb *IndexDb) GetPrefix(prefix []byte) map[string][]byte {
//	return getprefix(idxdb.db, prefix)
//}

// ###################### GET IMPL END ######################
//func (idxdb *IndexDb) GetUtxoByIndex(idx *modules.UtxoIndex) (*modules.Utxo, error) {
//	key := idx.ToKey()
//	utxo := new(modules.Utxo)
//	err := RetrieveFromRlpBytes(idxdb.db, key, utxo)
//	return utxo, err
//}
//func (idxdb *IndexDb) DeleteUtxoByIndex(idx *modules.UtxoIndex) error {
//	return idxdb.db.Delete(idx.ToKey())
//}

func (db *IndexDb) SaveAddressTxId(address common.Address, txid common.Hash) error {
	key := append(constants.ADDR_TXID_PREFIX, address.Bytes()...)
	key = append(key, txid[:]...)
	log.Debugf("Index address[%s] and tx[%s]", address.String(), txid.String())
	return db.db.Put(key, txid[:])
}
func (db *IndexDb) GetAddressTxIds(address common.Address) ([]common.Hash, error) {
	prefix := append(constants.ADDR_TXID_PREFIX, address.Bytes()...)
	data := getprefix(db.db, prefix)
	var result []common.Hash
	for _, v := range data {
		hash := common.Hash{}
		hash.SetBytes(v)
		result = append(result, hash)
	}
	return result, nil
}
func (db *IndexDb) SaveTokenTxId(asset *modules.Asset, txid common.Hash) error {
	key := append(constants.TOKEN_TXID_PREFIX, asset.Bytes()...)
	key = append(key, txid[:]...)
	log.Debugf("Index Token[%s] and tx[%s]", asset.String(), txid.String())
	return db.db.Put(key, txid[:])
}

func (db *IndexDb) GetTokenTxIds(asset *modules.Asset) ([]common.Hash, error) {
	prefix := append(constants.TOKEN_TXID_PREFIX, asset.Bytes()...)
	data := getprefix(db.db, prefix)
	var result []common.Hash
	for _, v := range data {
		hash := common.Hash{}
		hash.SetBytes(v)
		result = append(result, hash)
	}
	return result, nil
}

//
//func (db *IndexDb) GetFromAddressTxIds(addr string) ([]common.Hash, error) {
//	hashs := make([]common.Hash, 0)
//	data, err := db.db.Get(append(constants.AddrTx_From_Prefix, []byte(addr)...))
//	if err != nil {
//
//		return nil, err
//	}
//	if err := rlp.DecodeBytes(data, &hashs); err != nil {
//		return hashs, err
//	}
//	return hashs, nil
//}
//
//func (db *IndexDb) GetTxFromAddresses(tx *modules.Transaction) ([]string, error) {
//
//	froms := make([]string, 0)
//	if tx == nil {
//		return froms, errors.New("tx is nil, not exist address.")
//	}
//	outpoints, _ := tx.GetAddressInfo()
//	for _, op := range outpoints {
//		addr, err := db.getOutpointAddr(op)
//		if err == nil {
//			froms = append(froms, addr)
//		}
//	}
//
//	return froms, nil
//}
//func (db *IndexDb) getOutpointAddr(outpoint *modules.OutPoint) (string, error) {
//	out_key := append(constants.OUTPOINT_ADDR_PREFIX, outpoint.ToKey()...)
//	data, err := db.db.Get(out_key[:])
//	if len(data) <= 0 {
//		return "", errors.New(fmt.Sprintf("address is null. outpoint_key(%s)", outpoint.ToKey()))
//	}
//	if err != nil {
//		return "", err
//	}
//	var str string
//	err0 := rlp.DecodeBytes(data, &str)
//	return str, err0
//}

//save filehash key:IDX_MAIN_DATA_TXID   value:Txid
func (db *IndexDb) SaveMainDataTxId(filehash []byte, txid common.Hash) error {
	key := append(constants.IDX_MAIN_DATA_TXID, []byte(filehash)...)
	key = append(key, []byte(txid.String())...)

	return db.db.Put(key, txid[:])
}

func (db *IndexDb) GetMainDataTxIds(filehash []byte) ([]common.Hash, error) {
	key := append(constants.IDX_MAIN_DATA_TXID, []byte(filehash)...)
	data := getprefix(db.db, key)
	var result []common.Hash
	for _, v := range data {
		hash := common.Hash{}
		hash.SetBytes(v)
		result = append(result, hash)
	}
	return result, nil
}
