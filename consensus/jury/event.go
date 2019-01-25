/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package jury

import (
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common"
)

type ContractEventType int
type EventType int

const (
	CONTRACT_EVENT_EXEC   ContractEventType = 1 //合约执行，系统合约由Mediator完成，用户合约由Jury完成
	CONTRACT_EVENT_SIG                      = 2 //多Jury执行合约并签名转发确认，由Jury接收并处理
	CONTRACT_EVENT_COMMIT                   = 4 //提交给Mediator进行验证确认并写到交易池
)

const (
	RequestEvent EventType = 1
	ResultEvent            = 2
)

type ContractEvent struct {
	Addrs []common.Address //user contract jury addr

	CType ContractEventType
	Tx    *modules.Transaction
}

type ElectionRequestEvent struct {
	Num  int
	Data []byte
}
type ElectionResultEvent struct {
}

type ElectionEvent struct {
	Num  int    //del
	Data []byte //del

	EType EventType
	Event interface{}
}
