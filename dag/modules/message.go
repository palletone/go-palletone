/* This file is part of go-palletone.
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

   @author PalletOne core developers <dev@pallet.one>
   @date 2018
*/

package modules

import (
	"fmt"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
)

type MessageType byte

const (
	APP_PAYMENT MessageType = iota
	APP_CONTRACT_TPL
	APP_CONTRACT_DEPLOY
	APP_CONTRACT_INVOKE
	APP_CONFIG
	APP_TEXT
	APP_VOTE
	OP_MEDIATOR_CREATE
)

// key: message.UnitHash(message+timestamp)
type Message struct {
	App     MessageType `json:"app"`     // message type
	Payload interface{} `json:"payload"` // the true transaction data
}

// return message struct
func NewMessage(app MessageType, payload interface{}) *Message {
	m := new(Message)
	m.App = app
	m.Payload = payload
	return m
}

func (msg *Message) CopyMessages(cpyMsg *Message) *Message {
	msg.App = cpyMsg.App
	msg.Payload = cpyMsg.Payload
	switch cpyMsg.App {
	case APP_PAYMENT, APP_CONTRACT_TPL, APP_TEXT, APP_VOTE:
		msg.Payload = cpyMsg.Payload
	case APP_CONFIG:
		payload, _ := cpyMsg.Payload.(*ConfigPayload)
		newPayload := ConfigPayload{
			ConfigSet: []PayloadMapStruct{},
		}
		for _, p := range payload.ConfigSet {
			newPayload.ConfigSet = append(newPayload.ConfigSet, PayloadMapStruct{Key: p.Key, Value: p.Value})
		}
		msg.Payload = newPayload
	case APP_CONTRACT_DEPLOY:
		payload, _ := cpyMsg.Payload.(*ContractDeployPayload)
		newPayload := ContractDeployPayload{
			TemplateId:   payload.TemplateId,
			ContractId:   payload.ContractId,
			Args:         payload.Args,
			Excutiontime: payload.Excutiontime,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Version: &StateVersion{Height: rs.Version.Height, TxIndex: rs.Version.TxIndex}})
		}
		writeSet := []PayloadMapStruct{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	case APP_CONTRACT_INVOKE:
		payload, _ := cpyMsg.Payload.(*ContractInvokePayload)
		newPayload := ContractInvokePayload{
			ContractId:   payload.ContractId,
			Args:         payload.Args,
			Excutiontime: payload.Excutiontime,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Version: &StateVersion{Height: rs.Version.Height, TxIndex: rs.Version.TxIndex}})
		}
		writeSet := []PayloadMapStruct{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	}
	return msg
}

type PayloadMapStruct struct {
	IsDelete bool
	Key      string
	Value    []byte
}

func ToPayloadMapValueBytes(data interface{}) []byte {
	b, err := rlp.EncodeToBytes(data)
	if err != nil {
		return nil
	}
	return b
}

// Token exchange message and verify message
// App: payment
type PaymentPayload struct {
	Input    []*Input  `json:"inputs"`
	Output   []*Output `json:"outputs"`
	LockTime uint32    `json:"lock_time"`
}

// NewTxOut returns a new bitcoin transaction output with the provided
// transaction value and public key script.
func NewTxOut(value uint64, pkScript []byte, asset *Asset) *Output {
	return &Output{
		Value:    value,
		PkScript: pkScript,
		Asset:    asset,
	}
}

type StateVersion struct {
	Height  ChainIndex
	TxIndex uint32
}

func (version *StateVersion) String() string {

	return fmt.Sprintf(
		"StateVersion[AssetId:{%#x}, Height:{%d},IsMain:%t,TxIdx:{%d}]",
		version.Height.AssetID,
		version.Height.Index,
		version.Height.IsMain,
		version.TxIndex)
}

func (version *StateVersion) ParseStringKey(key string) bool {
	ss := strings.Split(key, FIELD_SPLIT_STR)
	if len(ss) != 3 {
		return false
	}
	var v StateVersion
	if err := rlp.DecodeBytes([]byte(ss[2]), &v); err != nil {
		log.Error("State version parse string key", "error", err.Error())
		return false
	}
	if version == nil {
		version = &StateVersion{}
	}
	version.Height = v.Height
	version.TxIndex = v.TxIndex
	return true
}

//16+8+1+4=29
func (version *StateVersion) Bytes() []byte {
	idx := make([]byte, 8)
	littleEndian.PutUint64(idx, version.Height.Index)
	b := append(version.Height.AssetID.Bytes(), idx...)
	if version.Height.IsMain {
		b = append(b, byte(1))
	} else {
		b = append(b, byte(0))
	}
	txIdx := make([]byte, 4)
	littleEndian.PutUint32(txIdx, version.TxIndex)
	b = append(b, txIdx...)
	return b[:]
}
func (version *StateVersion) SetBytes(b []byte) {
	asset := IDType16{}
	asset.SetBytes(b[:16])
	heightIdx := littleEndian.Uint64(b[16:24])
	isMain := b[24]
	txIdx := littleEndian.Uint32(b[25:])
	cidx := ChainIndex{AssetID: asset, Index: heightIdx, IsMain: isMain == byte(1)}
	version.Height = cidx
	version.TxIndex = txIdx
}

// Contract template deploy message
// App: contract_template
type ContractTplPayload struct {
	TemplateId []byte `json:"template_id"` // contract template id
	Name       string `json:"name"`        // contract template name
	Path       string `json:"path"`        // contract template execute path
	Version    string `json:"version"`     // contract template version
	Memory     uint16 `json:"memory"`      // contract template bytecode memory size(Byte), use to compute transaction fee
	Bytecode   []byte `json:"bytecode"`    // contract bytecode
}

const (
	FIELD_TPL_BYTECODE  = "TplBytecode"
	FIELD_TPL_NAME      = "TplName"
	FIELD_TPL_PATH      = "TplPath"
	FIELD_TPL_Memory    = "TplMemory"
	FIELD_SPLIT_STR     = "^*^"
	FIELD_GENESIS_ASSET = "GenesisAsset"
)

type DelContractState struct {
	IsDelete bool
}

func (delState DelContractState) Bytes() []byte {
	data, err := rlp.EncodeToBytes(delState)
	if err != nil {
		return nil
	}
	return data
}

func (delState DelContractState) SetBytes(b []byte) error {
	if err := rlp.DecodeBytes(b, &delState); err != nil {
		return err
	}
	return nil
}

type ContractReadSet struct {
	Key     string
	Version *StateVersion
	Value   []byte
}

// 0.default vote result is the index of the option from list
// 1.If the option is specified by the voter, set Option null
// 2.Expected vote result:[]byte
//type VoteInitiatePayload struct {
//	Title       string        //vote title
//	Option      []string      //vote option list.
//	BallotChain uint64        //vote chain id
//	BallotType  IDType16      //vote asset id
//	BallotCost  big.Int       //token cost
//	ExpiredTime time.Duration //duration of voting
//}

//VotePayload YiRan@
// Mode == 0 [ Replace ] :replace all
// Mode == 1 [ Edit    ] :Replace the addresses in the first half of the account's votes addresses to refer to the addresses in the second half.
// Mode == 2 [ Delete  ] :Delete the addresses from account's votes addresses
type VotePayload struct {
	Contents [][]byte
	VoteType uint8
	Mode     uint8
}

// Contract instance message
// App: contract_deploy

type ContractDeployPayload struct {
	TemplateId   []byte             `json:"template_id"`           // contract template id
	ContractId   []byte             `json:"contract_id"`           // contract id
	Name         string             `json:"name"`                  // the name for contract
	Args         [][]byte           `json:"args"`                  // contract arguments list
	Excutiontime time.Duration      `json:"excution_time" rlp:"-"` // contract execution time, millisecond
	Jury         []common.Address   `json:"jury"`                  // contract jurors list
	ReadSet      []ContractReadSet  `json:"read_set"`              // the set data of read, and value could be any type
	WriteSet     []PayloadMapStruct `json:"write_set"`             // the set data of write, and value could be any type
}

// Contract invoke message
// App: contract_invoke
//如果是用户想修改自己的State信息，那么ContractId可以为空或者0字节
type ContractInvokePayload struct {
	ContractId   []byte             `json:"contract_id"` // contract id
	FunctionName string             `json:"function_name"`
	Args         [][]byte           `json:"args"`          // contract arguments list
	Excutiontime time.Duration      `json:"excution_time"` // contract execution time, millisecond
	ReadSet      []ContractReadSet  `json:"read_set"`      // the set data of read, and value could be any type
	WriteSet     []PayloadMapStruct `json:"write_set"`     // the set data of write, and value could be any type
	Payload      []byte             `json:"payload"`       // the contract execution result
}

//用户钱包发起的合约调用申请
type ContractInvokeRequestPayload struct {
	ContractId   []byte   `json:"contract_id"` // contract id
	FunctionName string   `json:"function_name"`
	Args         [][]byte `json:"args"` // contract arguments list
}

// Token exchange message and verify message
// App: config	// update global config
type ConfigPayload struct {
	ConfigSet []PayloadMapStruct `json:"config_set"` // the array of global config
}

// Token exchange message and verify message
// App: text
type TextPayload struct {
	Text []byte `json:"text"` // Textdata
}
