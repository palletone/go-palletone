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

package txspool

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

type utxoBaseGetOp interface {
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
}
type utxoBaseOp interface {
	utxoBaseGetOp
	SaveUtxoEntity(outpoint *modules.OutPoint, utxo *modules.Utxo) error
}

//  UtxoViewpoint
type UtxoViewpoint struct {
	entries  map[modules.OutPoint]*modules.Utxo
	bestHash common.Hash
}

func (view *UtxoViewpoint) BestHash() *common.Hash {
	return &view.bestHash
}
func (view *UtxoViewpoint) SetBestHash(hash *common.Hash) {
	view.bestHash = *hash
}
func (view *UtxoViewpoint) SetEntries(key modules.OutPoint, utxo *modules.Utxo) {
	if view.entries == nil {
		view.entries = make(map[modules.OutPoint]*modules.Utxo)
	}

	view.entries[key] = utxo
}
func (view *UtxoViewpoint) AddUtxo(key modules.OutPoint, utxo *modules.Utxo) {
	if view.entries == nil {
		view.entries = make(map[modules.OutPoint]*modules.Utxo)
	}
	view.entries[key] = utxo
}
func (view *UtxoViewpoint) LookupUtxo(outpoint modules.OutPoint) *modules.Utxo {
	if view == nil {
		return nil
	}
	return view.entries[outpoint]
}
func (view *UtxoViewpoint) SpentUtxo(db utxoBaseOp, outpoints map[modules.OutPoint]struct{}) error {
	if len(outpoints) == 0 {
		return nil
	}
	for outpoint := range outpoints {
		item := new(modules.OutPoint)
		item.TxHash = outpoint.TxHash
		item.MessageIndex = outpoint.MessageIndex
		item.OutIndex = outpoint.OutIndex
		if utxo, has := view.entries[outpoint]; has {
			utxo.Spend()
			db.SaveUtxoEntity(item, utxo)
		} else {
			utxo, err := db.GetUtxoEntry(item)
			if err == nil {
				utxo.Spend()
				db.SaveUtxoEntity(item, utxo)
			}
		}
		delete(view.entries, outpoint)
	}
	return nil
}
func (view *UtxoViewpoint) FetchUnitUtxos(db utxoBaseGetOp, unit *modules.Unit) error {
	transactions := unit.Transactions()
	if len(transactions) <= 1 {
		return nil
	}

	txInFlight := map[common.Hash]int{}
	for i, tx := range transactions {
		txInFlight[tx.Hash()] = i
	}
	neededSet := make(map[modules.OutPoint]struct{})
	for i, tx := range transactions[1:] {
		// It is acceptable for a transaction input to reference
		// the output of another transaction in this block only
		// if the referenced transaction comes before the
		// current one in this block.  Add the outputs of the
		// referenced transaction as available utxos when this
		// is the case.  Otherwise, the utxo details are still
		// needed.
		//
		// NOTE: The >= is correct here because i is one less
		// than the actual position of the transaction within
		// the block due to skipping the coinbase.
		for j, msgcopy := range tx.TxMessages {
			if msgcopy.App == modules.APP_PAYMENT {
				if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
					for _, txIn := range msg.Inputs {
						//TODO for download sync
						if txIn == nil {
							continue
						}
						originHash := &txIn.PreviousOutPoint.TxHash
						if inFlightIndex, ok := txInFlight[*originHash]; ok &&
							i >= inFlightIndex {

							originTx := transactions[inFlightIndex]
							view.AddTxOut(originTx, uint32(i), uint32(j))
							continue
						}

						// Don't request entries that are already in the view
						// from the database.
						if _, ok := view.entries[*txIn.PreviousOutPoint]; ok {
							continue
						}
						neededSet[*txIn.PreviousOutPoint] = struct{}{}
					}
				}
			}
		}
	}

	return view.fetchUtxosMain(db, neededSet)
}

func (view *UtxoViewpoint) FetchUtxos(db utxoBaseGetOp, outpoints map[modules.OutPoint]struct{}) error {
	if len(outpoints) == 0 {
		return nil
	}

	return view.fetchUtxosMain(db, outpoints)
}
func (view *UtxoViewpoint) fetchUtxosMain(db utxoBaseGetOp, outpoints map[modules.OutPoint]struct{}) error {
	if len(outpoints) == 0 {
		return nil
	}
	for outpoint := range outpoints {
		item := new(modules.OutPoint)
		item.TxHash = outpoint.TxHash
		item.MessageIndex = outpoint.MessageIndex
		item.OutIndex = outpoint.OutIndex
		utxo, err := db.GetUtxoEntry(item)
		if err != nil {
			return err
		}
		view.entries[outpoint] = utxo
	}
	return nil
}

func (view *UtxoViewpoint) addTxOut(outpoint modules.OutPoint, txOut *modules.TxOut, isCoinbase bool) {
	// Don't add provably unspendable outputs.
	if tokenengine.IsUnspendable(txOut.PkScript) {
		return
	}
	utxo := view.LookupUtxo(outpoint)
	if utxo == nil {
		utxo = new(modules.Utxo)
		view.entries[outpoint] = utxo
	}
	utxo.Amount = uint64(txOut.Value)
	utxo.PkScript = txOut.PkScript
	utxo.Asset = txOut.Asset

	if isCoinbase {
		utxo.SetCoinBase() // utxo.Flags = modules.tfCoinBase
	}
}

func (view *UtxoViewpoint) AddTxOut(tx *modules.Transaction, msgIdx, txoutIdx uint32) {
	if msgIdx >= uint32(len(tx.TxMessages)) {
		return
	}

	for i, msgcopy := range tx.TxMessages {

		if (uint32(i) == msgIdx) && (msgcopy.App == modules.APP_PAYMENT) {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				if txoutIdx >= uint32(len(msg.Outputs)) {
					return
				}
				preout := modules.OutPoint{TxHash: tx.Hash(), MessageIndex: msgIdx, OutIndex: txoutIdx}
				output := msg.Outputs[txoutIdx]
				txout := &modules.TxOut{Value: int64(output.Value), PkScript: output.PkScript, Asset: output.Asset}
				view.addTxOut(preout, txout, false)
			}
		}

	}
}

func (view *UtxoViewpoint) AddTxOuts(tx *modules.Transaction) {
	preout := modules.OutPoint{TxHash: tx.Hash()}
	for i, msgcopy := range tx.TxMessages {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				msgIdx := uint32(i)
				preout.MessageIndex = msgIdx
				for j, output := range msg.Outputs {
					txoutIdx := uint32(j)
					preout.OutIndex = txoutIdx
					txout := &modules.TxOut{Value: int64(output.Value), PkScript: output.PkScript, Asset: output.Asset}
					view.addTxOut(preout, txout, false)
				}
			}
		}

	}
}

func (view *UtxoViewpoint) RemoveUtxo(outpoint modules.OutPoint) {
	delete(view.entries, outpoint)
}

func (view *UtxoViewpoint) Entries() map[modules.OutPoint]*modules.Utxo {
	return view.entries
}

func NewUtxoViewpoint() *UtxoViewpoint {
	return &UtxoViewpoint{
		entries: make(map[modules.OutPoint]*modules.Utxo),
	}
}

// ErrorCode identifies a kind of error.
type ErrorCode uint8

// RuleError identifies a rule violation.  It is used to indicate that
// processing of a block or transaction failed due to one of the many validation
// rules.  The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and access the ErrorCode field to
// ascertain the specific reason for the rule violation.
//type RuleError struct {
//	ErrorCode   ErrorCode // Describes the kind of error
//	Description string    // Human readable description of the issue
//}
type RuleError struct {
	ErrorCode   RejectCode // Describes the kind of error
	Description string     // Human readable description of the issue
}

// TxRuleError identifies a rule violation.  It is used to indicate that
// processing of a transaction failed due to one of the many validation
// rules.  The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and access the ErrorCode field to
// ascertain the specific reason for the rule violation.
type TxRuleError struct {
	RejectCode  RejectCode // The code to send with reject messages
	Description string     // Human readable description of the issue
}

type RejectCode uint8

// These constants define the various supported reject codes.
const (
	RejectMalformed       RejectCode = 0x01
	RejectInvalid         RejectCode = 0x10
	RejectObsolete        RejectCode = 0x11
	RejectDuplicate       RejectCode = 0x12
	RejectNonstandard     RejectCode = 0x40
	RejectDust            RejectCode = 0x41
	RejectInsufficientFee RejectCode = 0x42
	RejectCheckpoint      RejectCode = 0x43
)

func CheckTransactionSanity(tx *modules.Transaction) error {
	// A transaction must have at least one input.
	if len(tx.TxMessages) == 0 {
		return errors.New(fmt.Sprintf("transaction has no inputs,hash: %s", tx.Hash().String()))
	}
	// A transaction must not exceed the maximum allowed block payload when
	// serialized.
	serializedTxSize := tx.SerializeSizeStripped()
	if serializedTxSize > modules.TX_BASESIZE {
		str := fmt.Sprintf("serialized transaction is too big - got "+
			"%d, max %d", serializedTxSize, modules.TX_BASESIZE)
		return errors.New(str)
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value known
	// as a satoshi.  One bitcoin is a quantity of satoshi as defined by the
	// SatoshiPerBitcoin constant.
	// var totalSatoshi uint64
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}

		// Check for duplicate transaction inputs.
		existingTxOut := make(map[modules.OutPoint]struct{})
		for _, txIn := range payload.Inputs {
			if txIn.PreviousOutPoint != nil {
				if _, exists := existingTxOut[*txIn.PreviousOutPoint]; exists {
					return errors.New("transaction " + "contains duplicate inputs")
				}
				existingTxOut[*txIn.PreviousOutPoint] = struct{}{}
			}
		}
	}
	return nil
}
