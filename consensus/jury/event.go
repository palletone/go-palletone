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
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/common/util"
)

type ContractEventType uint8
type ElectionEventType uint8
type AdapterEventType uint32

const (
	CONTRACT_EVENT_EXEC   ContractEventType = 1 //合约执行，系统合约由Mediator完成，用户合约由Jury完成
	CONTRACT_EVENT_SIG    ContractEventType = 2 //多Jury执行合约并签名转发确认，由Jury接收并处理
	CONTRACT_EVENT_COMMIT ContractEventType = 4 //提交给Mediator进行验证确认并写到交易池
	CONTRACT_EVENT_ELE    ContractEventType = 8 //节点选举
)
const (
	ELECTION_EVENT_VRF_REQUEST ElectionEventType = 1
	ELECTION_EVENT_VRF_RESULT  ElectionEventType = 2
	ELECTION_EVENT_SIG_REQUEST ElectionEventType = 3
	ELECTION_EVENT_SIG_RESULT  ElectionEventType = 4
)

type AdapterInf struct {
	JuryMsgAll map[string]*MsgSigCollect
}
type MsgSigCollect struct {
	OneMsgAllSig map[string]JuryMsgSig
}
type JuryMsgSig struct {
	Signature []byte
	Answer    []byte
}
type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

//contract
type ContractEvent struct {
	CType ContractEventType
	Ele   *modules.ElectionNode `rlp:"nil"`
	Tx    *modules.Transaction
}

func (ce *ContractEvent) Hash() common.Hash {
	return util.RlpHash(ce)
}

//Election
type ElectionRequestEvent struct {
	ReqId     common.Hash
	JuryCount uint64
}
type ElectionResultEvent struct {
	ReqId     common.Hash
	JuryCount uint64
	Ele       modules.ElectionInf
}

//sig
type ElectionSigRequestEvent struct {
	ReqId     common.Hash
	JuryCount uint64
	Ele       []modules.ElectionInf
}
type ElectionSigResultEvent struct {
	ReqId     common.Hash
	JuryCount uint64
	Sig       modules.SignatureSet
}

type ElectionEvent struct {
	EType ElectionEventType `json:"etype"`
	Event interface{}       `json:"event"`
}
type ElectionEventBytes struct {
	EType ElectionEventType `json:"etype"`
	Event []byte            `json:"event"`
}

func (es *ElectionEventBytes) Hash() common.Hash {
	return util.RlpHash(es)
}

func (es *ElectionEventBytes) ToElectionEvent() (*ElectionEvent, error) {
	var err error
	event := ElectionEvent{}
	event.EType = es.EType
	if es.EType == ELECTION_EVENT_VRF_REQUEST {
		var evt ElectionRequestEvent
		err = json.Unmarshal(es.Event, &evt)
		if err != nil {
			return nil, err
		}
		event.Event = &evt
	} else if es.EType == ELECTION_EVENT_VRF_RESULT {
		var evt ElectionResultEvent
		err = json.Unmarshal(es.Event, &evt)
		if err != nil {
			return nil, err
		}
		event.Event = &evt
	} else if es.EType == ELECTION_EVENT_SIG_REQUEST {
		var evt ElectionSigRequestEvent
		err = json.Unmarshal(es.Event, &evt)
		if err != nil {
			return nil, err
		}
		event.Event = &evt
	} else if es.EType == ELECTION_EVENT_SIG_RESULT {
		var evt ElectionSigResultEvent
		err = json.Unmarshal(es.Event, &evt)
		if err != nil {
			return nil, err
		}
		event.Event = &evt
	}
	return &event, nil
}

func (ev *ElectionEvent) ToElectionEventBytes() (*ElectionEventBytes, error) {
	es := &ElectionEventBytes{}

	byteJson, err := json.Marshal(ev.Event)
	es.EType = ev.EType
	es.Event = byteJson
	return es, err
}

//Adapter
type AdapterEvent struct {
	AType AdapterEventType `json:"atype"`
	Event interface{}      `json:"event"`
}

func (av *AdapterEvent) ToAdapterEventBytes() (*AdapterEventBytes, error) {
	as := &AdapterEventBytes{}

	byteJson, err := json.Marshal(av.Event)
	as.AType = av.AType
	as.Event = byteJson
	return as, err
}

type AdapterRequestEvent struct {
	ReqId       common.Hash    `json:"reqId"`
	ContractId  common.Address `json:"contractId"`  //
	ConsultData []byte         `json:"consultdata"` //
	Answer      []byte         `json:"Answer"`
	Sig         []byte         `json:"sig"`
	Pubkey      []byte         `json:"Pubkey"`
}

type AdapterEventBytes struct {
	AType AdapterEventType `json:"atype"`
	Event []byte           `json:"event"`
}

func (es *AdapterEventBytes) ToAdapterEvent() (*AdapterEvent, error) {
	event := AdapterEvent{}
	event.AType = es.AType
	var req AdapterRequestEvent
	err := json.Unmarshal(es.Event, &req)
	if err != nil {
		return nil, err
	}
	event.Event = &req

	return &event, nil
}
