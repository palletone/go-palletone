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

package adaptor

import (
	"encoding/hex"
	"encoding/json"
)

//SimpleTransferTokenTx 一个简单的Token转账交易
type SimpleTransferTokenTx struct {
	TxBasicInfo
	FromAddress string       `json:"from_address"` //转出地址
	ToAddress   string       `json:"to_address"`   //转入地址
	Amount      *AmountAsset `json:"amount"`       //转账金额
	Fee         *AmountAsset `json:"fee"`          //转账交易费
	AttachData  []byte       `json:"attach_data"`  //附加的数据（备注之类的）
}

func (tx *SimpleTransferTokenTx) String() string {
	d, _ := json.Marshal(tx)
	return string(d)
}

type simpleTransferTokenTx4Json struct {
	txBasicInfo4Json
	FromAddress string       `json:"from_address"` //转出地址
	ToAddress   string       `json:"to_address"`   //转入地址
	Amount      *AmountAsset `json:"amount"`       //转账金额
	Fee         *AmountAsset `json:"fee"`          //转账交易费
	AttachData  string       `json:"attach_data"`  //附加的数据（备注之类的）
}

func setSimpleTransferTokenTxFromJson(tx *SimpleTransferTokenTx, tx4json simpleTransferTokenTx4Json) {
	setTxBasicInfoFromJson(&tx.TxBasicInfo, tx4json.txBasicInfo4Json)
	tx.FromAddress = tx4json.FromAddress
	tx.ToAddress = tx4json.ToAddress
	tx.Amount = tx4json.Amount
	tx.Fee = tx4json.Fee
	tx.AttachData, _ = hex.DecodeString(tx4json.AttachData)
}
func convertSimpleTransferTokenTx2Json(tx SimpleTransferTokenTx) simpleTransferTokenTx4Json {
	tx4Json := simpleTransferTokenTx4Json{}
	tx4Json.FromAddress = tx.FromAddress
	tx4Json.ToAddress = tx.ToAddress
	tx4Json.Amount = tx.Amount
	tx4Json.Fee = tx.Fee
	tx4Json.AttachData = hex.EncodeToString(tx.AttachData)
	tx4Json.txBasicInfo4Json = convertTxBasicInfo2Json(tx.TxBasicInfo)
	return tx4Json
}
func (tx *SimpleTransferTokenTx) MarshalJSON() ([]byte, error) {
	tx4json := convertSimpleTransferTokenTx2Json(*tx)
	return json.Marshal(tx4json)
}
func (tx *SimpleTransferTokenTx) UnmarshalJSON(input []byte) error {
	tx4Json := simpleTransferTokenTx4Json{}
	err := json.Unmarshal(input, &tx4Json)
	if err != nil {
		return err
	}
	setSimpleTransferTokenTxFromJson(tx, tx4Json)
	return nil
}

//多地址对多地址的转账交易
// type MultiAddrTransferTokenTx struct {
// 	TxBasicInfo
// 	FromAddress map[string]*AmountAsset //转出地址
// 	ToAddress   map[string]*AmountAsset //转入地址
// 	Fee         *AmountAsset            //转账交易费
// 	AttachData  []byte                  //附加的数据（备注之类的）
// }
