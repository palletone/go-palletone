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
 *  * @date 2018-2019
 *
 */

package validator

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidate_ValidateTx_EmptyTx_NoPayment(t *testing.T) {
	tx := &modules.Transaction{} //Empty Tx
	validat := NewValidate(nil, nil, nil)
	err := validat.ValidateTx(tx, false)
	assert.NotNil(t, err)
	t.Log(err)
	tx.AddMessage(modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte("m")}))
	err = validat.ValidateTx(tx, false)
	assert.NotNil(t, err)
	t.Log(err)
}
func TestValidate_ValidateTx_MsgCodeIncorrect(t *testing.T) {
	tx := &modules.Transaction{}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, &modules.DataPayload{MainData: []byte("m")}))

	validat := NewValidate(nil, nil, nil)
	err := validat.ValidateTx(tx, false)
	assert.NotNil(t, err)
	t.Log(err)

}

var hash1 = common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac")

func TestValidate_ValidateTx_IncorrectFee(t *testing.T) {
	tx := &modules.Transaction{}
	outPoint := &modules.OutPoint{hash1, 0, 1}
	pay1 := newTestPayment(outPoint, 2000)
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay1))
	utxoq := &testutxoQuery{}
	validat := NewValidate(nil, utxoq, nil)
	err := validat.ValidateTx(tx, false)
	assert.NotNil(t, err)
	t.Log(err)

	pay1.Outputs[0].Value = 999
	err = validat.ValidateTx(tx, false)
	assert.NotNil(t, err)
	t.Log(err)

}

type testutxoQuery struct {
}

func (u *testutxoQuery) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	lockScript := []byte{}
	if outpoint.TxHash == hash1 {
		return &modules.Utxo{Asset: modules.NewPTNAsset(), Amount: 1000, PkScript: lockScript}, nil
	}
	return nil, errors.New("Incorrect Hash")
}
func newTestPayment(point *modules.OutPoint, outAmt uint64) *modules.PaymentPayload {
	pay1s := &modules.PaymentPayload{
		LockTime: 12345,
	}
	a := &modules.Asset{AssetId: modules.PTNCOIN}

	output := modules.NewTxOut(outAmt, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay1s.AddTxOut(output)
	i := modules.NewTxIn(point, []byte("a"))
	pay1s.AddTxIn(i)
	return pay1s
}
