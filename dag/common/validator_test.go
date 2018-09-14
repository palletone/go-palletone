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

package common

import (
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
)

func TestValidator(t *testing.T) {
	outpoint := modules.OutPoint{
		MessageIndex: 2,
		OutIndex:     3,
	}
	createT := big.Int{}
	totalIncome := uint64(100000000)
	addr := new(common.Address)
	addr.SetString("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	script := tokenengine.GenerateP2PKHLockScript(addr.Bytes())
	input := &modules.Input{
		PreviousOutPoint: &outpoint,
		SignatureScript:  []byte("xxxxxxxxxx"),
		Extra:            createT.SetInt64(time.Now().Unix()).Bytes(),
	}
	output := &modules.Output{
		Value: totalIncome,
		Asset: &modules.Asset{
			AssetId: modules.BTCCOIN,
			ChainId: 1},
		PkScript: script,
	}

	inputs := make([]*modules.Input, 0)
	outputs := make([]*modules.Output, 0)
	inputs = append(inputs, input)
	outputs = append(outputs, output)
	tx := new(modules.Transaction)
	tx.TxMessages = append(tx.TxMessages, &modules.Message{App: modules.APP_PAYMENT, Payload: &modules.PaymentPayload{Input: inputs, Output: outputs, LockTime: uint32(999)}},
		&modules.Message{App: modules.APP_TEXT, Payload: &modules.TextPayload{Text: []byte("test text.")}}, &modules.Message{App: modules.APP_CONTRACT_TPL, Payload: &modules.ContractTplPayload{Name: "contract name"}})
	tx.Hash()
	log.Println("tx hash :", tx.TxHash.String(), tx.TxMessages[2])
	//dbconn := storage.ReNewDbConn("D:\\Workspace\\Code\\Go\\src\\github.com\\palletone\\go-palletone\\bin\\gptn\\leveldb")
	dbconn := storage.ReNewDbConn(dagconfig.DbPath)
	worldTmpState := map[string]map[string]interface{}{}
	dagDb := storage.NewDagDatabase(dbconn)
	utxoDb := storage.NewUtxoDatabase(dbconn)
	stateDb := storage.NewStateDatabase(dbconn)
	validate := NewValidate(dagDb, utxoDb, stateDb)
	code := validate.ValidateTx(tx, false, &worldTmpState)
	log.Println("validator code:", code, worldTmpState)
}
