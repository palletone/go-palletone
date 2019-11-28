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

import "github.com/palletone/go-palletone/dag/modules"
//ValidatorAllPass used for UT
type ValidatorAllPass struct{

}
func(v *ValidatorAllPass) ValidateTx(tx *modules.Transaction, isFullTx bool) ([]*modules.Addition, ValidationCode, error){
	return nil, TxValidationCode_VALID, nil
}
//验证一个Unit中的所有交易是否是合法交易
//ValidateTransactions(txs modules.Transactions) error
//除了群签名外，验证Unit是否是合法Unit,包括其中的所有交易都会逐一验证
func(v *ValidatorAllPass)ValidateUnitExceptGroupSig(unit *modules.Unit) ValidationCode{
	return TxValidationCode_VALID
}
//ValidateUnitExceptPayment(unit *modules.Unit) error
//验证一个Header是否合法（Mediator签名有效）
func(v *ValidatorAllPass)ValidateHeader(h *modules.Header) ValidationCode{
	return TxValidationCode_VALID
}
func(v *ValidatorAllPass)ValidateUnitGroupSign(h *modules.Header) error{
	return nil
}
func(v *ValidatorAllPass)CheckTxIsExist(tx *modules.Transaction) bool{
	return false
}