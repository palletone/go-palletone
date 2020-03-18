/*
 *  This file is part of go-palletone.
 *  go-palletone is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *  go-palletone is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *  You should have received a copy of the GNU General Public License
 *  along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 *
 *  @author PalletOne core developer <dev@pallet.one>
 *  @date 2018-2020
 */

package txpool2

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
	"github.com/stretchr/testify/assert"
)

func TestTxList_GetSortTxs(t *testing.T) {
	txlist := newTxList()
	txlist.AddTx(mockTxpoolTx(nil, Hash("A")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("A")}, Hash("B")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("B")}, Hash("C")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("C"), Hash("B")}, Hash("D")))
	txlist.AddTx(mockTxpoolTx(nil, Hash("X")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("X")}, Hash("Y1")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("X")}, Hash("Y2")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("X")}, Hash("Y3")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("Y1")}, Hash("Z1")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("Y2"), Hash("Y3")}, Hash("Z2")))

	sortedTx := ""
	txlist.GetSortedTxs(func(tx *txspool.TxPoolTransaction) (getNext bool, err error) {
		//t.Logf("tx[%s]",string(tx.TxHash[:]))
		sortedTx += string(tx.TxHash[:]) + ";"
		return true, nil
	})
	t.Log(sortedTx)
	match, _ := regexp.MatchString("A.*B.*C.*D.*", sortedTx)
	assert.True(t, match)

	match, _ = regexp.MatchString("X.*Y.*Z.*", sortedTx)
	assert.True(t, match)
}
func Hash(s string) common.Hash {
	return common.BytesToHash([]byte(s))
}
func mockTxpoolTx(parentTxHash []common.Hash, txHash common.Hash) *txspool.TxPoolTransaction {
	pHash := Hash("p")
	if len(parentTxHash) > 0 {
		pHash = parentTxHash[0]
	}
	pay := mockPaymentTx(pHash, 0, 0)
	tx := &txspool.TxPoolTransaction{DependOnTxs: map[common.Hash]bool{}, Tx: pay}
	for _, p := range parentTxHash {
		tx.DependOnTxs[p] = true
	}
	tx.TxHash = txHash
	return tx
}

func TestTxList_GetAllTxs(t *testing.T) {
	txlist := newTxList()
	txlist.AddTx(mockTxpoolTx(nil, Hash("A")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("A")}, Hash("B")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("B")}, Hash("C")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("C"), Hash("B")}, Hash("D")))
	all := txlist.GetAllTxs()
	for hash := range all {
		t.Logf("%s", hash.String())
	}
	assert.Equal(t, 4, len(all))
}

func TestTxList_GetSortTxs_Long(t *testing.T) {
	txlist := newTxList()
	tx := mockTxpoolTx(nil, Hash("A"))
	txlist.AddTx(tx)
	result := "A;"
	for i := 0; i < 100; i++ {
		tx = mockTxpoolTx([]common.Hash{tx.TxHash}, Hash("A"+strconv.Itoa(i)))
		txlist.AddTx(tx)
		result += "A" + strconv.Itoa(i) + ";"
	}
	sortedTx := ""
	txlist.GetSortedTxs(func(tx *txspool.TxPoolTransaction) (getNext bool, err error) {
		sortedTx += string(tx.TxHash[:]) + ";"
		if tx.TxHash == Hash("A10") {
			return false, nil
		}
		return true, nil
	})
	t.Log(sortedTx)
	t.Log(result)
	sortedTx = strings.Replace(sortedTx, "\x00", "", -1)
	assert.EqualValues(t, "A;A0;A1;A2;A3;A4;A5;A6;A7;A8;A9;A10;", sortedTx)
}
func TestTxList_DiscardTxs(t *testing.T) {
	txlist := initTxlist()
	txlist.DiscardTx(Hash("A"))
	txlist.DiscardTx(Hash("B"))
	result := printTxList(txlist)
	t.Log(result)
	match, _ := regexp.MatchString("A.*B.*", result)
	assert.False(t, match)

	match, _ = regexp.MatchString("C.*D.*", result)
	assert.True(t, match)

	match, _ = regexp.MatchString("X.*Y.*Z.*", result)
	assert.True(t, match)
}

func initTxlist() *txList {
	txlist := newTxList()
	txlist.AddTx(mockTxpoolTx(nil, Hash("A")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("A")}, Hash("B")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("B")}, Hash("C")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("C"), Hash("B")}, Hash("D")))
	txlist.AddTx(mockTxpoolTx(nil, Hash("X")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("X")}, Hash("Y1")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("X")}, Hash("Y2")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("X")}, Hash("Y3")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("Y1")}, Hash("Z1")))
	txlist.AddTx(mockTxpoolTx([]common.Hash{Hash("Y2"), Hash("Y3")}, Hash("Z2")))
	return txlist
}
func printTxList(txlist *txList) string {
	sortedTx := ""
	txlist.GetSortedTxs(func(tx *txspool.TxPoolTransaction) (getNext bool, err error) {
		sortedTx += string(tx.TxHash[:]) + ";"
		return true, nil
	})
	return sortedTx
}
func TestTxList_GetUnpackedTxs(t *testing.T) {
	txlist := initTxlist()
	txlist.DiscardTx(Hash("A"))
	txlist.DiscardTx(Hash("X"))
	txlist.UpdateTxStatusPacked(Hash("B"), Hash("U1"), 123)
	txlist.UpdateTxStatusPacked(Hash("C"), Hash("U1"), 123)
	result, err := txlist.GetTxsByStatus(txspool.TxPoolTxStatus_Unpacked)
	assert.Nil(t, err)
	txStr := ""
	for _, tx := range result {
		txStr += string(tx.TxHash[:]) + ";"
	}
	t.Log(txStr)
	match, _ := regexp.MatchString("[ABCX]", txStr)
	assert.False(t, match)
}
func TestTxList_GetUtxoEntry(t *testing.T) {
	txA := mockPaymentTx(Hash("0"), 0, 0)
	txB := mockPaymentTx(txA.Hash(), 0, 0)
	//txC:=mockPaymentTx(txB.Hash(),0,0)
	txlist := newTxList()
	txlist.AddTx(Tx2PoolTx(txA))
	for o := range txA.GetNewUtxos() {
		_, err := txlist.GetUtxoEntry(&o)
		assert.Nil(t, err)
	}
	txlist.AddTx(Tx2PoolTx(txB))
	for o := range txA.GetNewUtxos() {
		_, err := txlist.GetUtxoEntry(&o)
		assert.NotNil(t, err)
	}
	//txlist.AddTx(Tx2PoolTx(txC))
}
func Tx2PoolTx(tx *modules.Transaction) *txspool.TxPoolTransaction {
	return &txspool.TxPoolTransaction{Tx: tx, TxHash: tx.Hash()}
}
func mockPaymentTx(preTxHash common.Hash, msgIdx, outIdx uint32) *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	output := modules.NewTxOut(15000, lockScript, modules.NewPTNAsset())
	pay1s.AddTxOut(output)

	input := modules.Input{}
	input.PreviousOutPoint = modules.NewOutPoint(preTxHash, msgIdx, outIdx)
	input.SignatureScript = []byte{}
	pay1s.AddTxIn(&input)
	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}
	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	return tx
}
func mockContractInvokeRequest(preTxHash common.Hash, msgIdx, outIdx uint32, contractId []byte) *modules.Transaction {
	tx := mockPaymentTx(preTxHash, msgIdx, outIdx)
	reqPayload := &modules.ContractInvokeRequestPayload{
		ContractId: contractId,
		Args:       [][]byte{[]byte("put"), []byte("a"), []byte("100")},
		Timeout:    0,
	}
	invoke := modules.NewMessage(modules.APP_CONTRACT_INVOKE_REQUEST, reqPayload)
	tx.AddMessage(invoke)
	return tx
}
func mockContractInvokeFullTx(preTxHash common.Hash, msgIdx, outIdx uint32, contractId []byte) *modules.Transaction {
	tx := mockContractInvokeRequest(preTxHash, msgIdx, outIdx, contractId)
	invokePayload := &modules.ContractInvokePayload{
		ContractId: []byte("contractA"),
		Args:       [][]byte{[]byte("put"), []byte("a"), []byte("100")},
		ReadSet:    nil,
		WriteSet:   []modules.ContractWriteSet{{Key: "a", Value: []byte("100")}},
		Payload:    nil,
		ErrMsg:     modules.ContractError{},
	}
	invoke := modules.NewMessage(modules.APP_CONTRACT_INVOKE, invokePayload)
	tx.AddMessage(invoke)
	return tx
}

func TestTxList_GetTx(t *testing.T) {
	txlist := initTxlist()
	txA, err := txlist.GetTx(Hash("A"))
	assert.Nil(t, err)
	assert.Equal(t, txA.TxHash, Hash("A"))
	txM, err := txlist.GetTx(Hash("M"))
	assert.Nil(t, txM)
	assert.NotNil(t, err)
}
