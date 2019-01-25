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
	"github.com/palletone/go-palletone/dag/modules"
)

type Validator interface {
	ValidateTransactions(txs *modules.Transactions, isGenesis bool) (map[common.Hash]modules.TxValidationCode, bool, error)
	ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) byte
	ValidateTx(tx *modules.Transaction, isCoinbase bool, worldTmpState *map[string]map[string]interface{}) modules.TxValidationCode
	ValidateUnitSignature(h *modules.Header, isGenesis bool) byte
	//ValidateUnitGroupSign(h *modules.Header, isGenesis bool) byte
}
type IUtxoQuery interface {
	ComputeTxFee(tx *modules.Transaction) (*modules.AmountAsset, error)
	GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error)
}
type IStateQuery interface {
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string, tplVersion string)
}
type IDagQuery interface {
	GetTxHashByReqId(reqid common.Hash) (common.Hash, error)
}
