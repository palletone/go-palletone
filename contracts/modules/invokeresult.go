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

package modules

import (
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

type ContractInvokeResult struct {
	ContractId   []byte                     `json:"contract_id"` // contract id
	FunctionName string                     `json:"function_name"`
	Args         [][]byte                   `json:"args"`          // contract arguments list
	Excutiontime time.Duration              `json:"excution_time"` // contract execution time, millisecond
	ReadSet      []modules.ContractReadSet  `json:"read_set"`      // the set data of read, and value could be any type
	WriteSet     []modules.ContractWriteSet `json:"write_set"`     // the set data of write, and value could be any type
	Payload      []byte                     `json:"payload"`       // the contract execution result
	TokenPayOut  []*modules.TokenPayOut     `json:"token_payout"`  //从合约地址付出Token
	TokenSupply  []*modules.TokenSupply     `json:"token_supply"`  //增发Token请求产生的结果
	TokenDefine  *modules.TokenDefine       `json:"token_define"`   //定义新Token
}

func (result *ContractInvokeResult) ToContractInvokePayload() *modules.ContractInvokePayload {
	return modules.NewContractInvokePayload(result.ContractId, result.Args, result.Excutiontime, result.ReadSet, result.WriteSet, result.Payload)
}
