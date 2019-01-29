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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/stretchr/testify/assert"
	"testing"
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/dag/modules"
)


func TestGetUnit(t *testing.T) {
	//log.Println("dbconn is nil , renew db  start ...")

	db, _ := ptndb.NewMemDatabase()
	//l := log.NewTestLog()
	dagdb := NewDagDb(db)
	u, err := dagdb.GetHeader(common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"))
	assert.Nil(t, u, "empty db, must return nil Unit")
	assert.NotNil(t, err)
}

func TestPrintHashList(t *testing.T) {
	hash1 := common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
	hash2 := common.HexToHash("0xddff4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d493ee")
	txsHash := []common.Hash{hash1, hash2}
	t.Logf("%x", txsHash)

}

func TestGetHeader(t *testing.T) {
	key := new(ecdsa.PrivateKey)
	key, _ = crypto.GenerateKey()
	h := new(modules.Header)
	//h.AssetIDs = append(h.AssetIDs, PTNCOIN)
	au := modules.Authentifier{}
	address := crypto.PubkeyToAddress(&key.PublicKey)
	t.Log("address:", address)

	//author := &Author{
	//	Address:        address,
	//	Pubkey:         []byte("1234567890123456789"),
	//	TxAuthentifier: *au,
	//}

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &modules.ChainIndex{}
	h.Number.AssetID = modules.PTNCOIN
	h.Number.Index = uint64(333333)
	h.Extra = make([]byte, 20)
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)
	//tr := common.Hash{}
	//tr = tr.SetString("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	h.TxRoot = common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	sig, _ := crypto.Sign(h.TxRoot[:], key)
	au.R = sig[:32]
	au.S = sig[32:64]
	au.V = sig[64:]
	h.Authors = au
	h.Creationdate = 123

	t.Logf("%#v", h)

	db, _ := ptndb.NewMemDatabase()
	dagdb := NewDagDb(db)

	err := dagdb.SaveHeader(h)
	assert.Nil(t, err)
	dbHeader, err := dagdb.GetHeader(h.Hash())
	assert.Nil(t, err)
	t.Logf("%#v", dbHeader)
	assertRlpHashEqual(t, h, dbHeader)
}