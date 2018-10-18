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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
)

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
func (view *UtxoViewpoint) SpentUtxo(db storage.IUtxoDb, outpoints map[modules.OutPoint]struct{}) error {
	if len(outpoints) == 0 {
		return nil
	}
	for outpoint := range outpoints {
		if utxo, has := view.entries[outpoint]; has {
			utxo.Spend()
			db.SaveUtxoEntity(&outpoint, utxo)
		} else {
			utxo, err := db.GetUtxoEntry(&outpoint)
			if err == nil {
				utxo.Spend()
				db.SaveUtxoEntity(&outpoint, utxo)
			}
		}
		delete(view.entries, outpoint)
	}
	return nil
}
func (view *UtxoViewpoint) FetchUnitUtxos(db storage.IUtxoDb, unit *modules.Unit) error {
	txInFlight := map[common.Hash]int{}
	transactions := unit.Transactions()
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
					for _, txIn := range msg.Input {
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

// fetch OutUtxos ,return  utxos  all outpoint.
// func (view *UtxoViewpoint) FetchOutputUtxos(db storage.IUtxoDb, unit *modules.Unit) (map[modules.OutPoint]struct{}, error) {
// 	transactions := unit.Transactions()
// 	needSet := make(map[modules.OutPoint]struct{})
// 	for i, tx := range transactions {
// 		// TODO
// 		//

// 	}
// 	return needSet, nil
// }
func (view *UtxoViewpoint) FetchUtxos(db storage.IUtxoDb, outpoints map[modules.OutPoint]struct{}) error {
	if len(outpoints) == 0 {
		return nil
	}
	neededSet := make(map[modules.OutPoint]struct{})
	for outpoint := range outpoints {
		if _, ok := view.entries[outpoint]; ok {
			continue
		}
		neededSet[outpoint] = struct{}{}
	}
	return view.fetchUtxosMain(db, neededSet)

}
func (view *UtxoViewpoint) fetchUtxosMain(db storage.IUtxoDb, outpoints map[modules.OutPoint]struct{}) error {
	if len(outpoints) == 0 {
		return nil
	}
	for outpoint := range outpoints {
		utxo, err := db.GetUtxoEntry(&outpoint)
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

	// isCoinbase ?
	// flags --->  标记utxo状态
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
				if txoutIdx >= uint32(len(msg.Output)) {
					return
				}
				preout := modules.OutPoint{TxHash: tx.Hash(), MessageIndex: msgIdx, OutIndex: txoutIdx}
				output := msg.Output[txoutIdx]
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
				for j, output := range msg.Output {
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
