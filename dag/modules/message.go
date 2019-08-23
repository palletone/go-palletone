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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
)

type MessageType byte

const (
	APP_PAYMENT MessageType = iota

	APP_CONTRACT_TPL
	APP_CONTRACT_DEPLOY
	APP_CONTRACT_INVOKE
	APP_CONTRACT_STOP
	APP_SIGNATURE

	APP_DATA
	APP_ACCOUNT_UPDATE

	APP_UNKNOW = 99

	APP_CONTRACT_TPL_REQUEST    = 100
	APP_CONTRACT_DEPLOY_REQUEST = 101
	APP_CONTRACT_INVOKE_REQUEST = 102
	APP_CONTRACT_STOP_REQUEST   = 103
	// 为了兼容起见:
	// 添加別的request需要添加在 APP_CONTRACT_TPL_REQUEST 与 APP_CONTRACT_STOP_REQUEST 之间
	// 添加别的msg类型，需要添加到 APP_ACCOUNT_UPDATE 与 APP_UNKNOW之间
)

const (
	FoundationAddress = "FoundationAddress"
	JuryList          = "JuryList"
	DeveloperList     = "DeveloperList"
	DepositRate       = "DepositRate"
)

const (
	DesiredSysParamsWithoutVote = "desiredSysParamsWithoutVote"
	DesiredSysParamsWithVote    = "desiredSysParamsWithVote"
	DesiredActiveMediatorCount  = "ActiveMediatorCount"
)

func (mt MessageType) IsRequest() bool {
	return mt > APP_UNKNOW
}

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
	//msg.Payload = cpyMsg.Payload
	switch cpyMsg.App {
	default:
		//case APP_PAYMENT, APP_CONTRACT_TPL, APP_DATA:
		msg.Payload = cpyMsg.Payload

	case APP_CONTRACT_DEPLOY:
		payload, _ := cpyMsg.Payload.(*ContractDeployPayload)
		newPayload := ContractDeployPayload{
			ContractId: payload.ContractId,
			EleNode:    payload.EleNode,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Version: &StateVersion{
				Height: rs.Version.Height, TxIndex: rs.Version.TxIndex}})
		}
		writeSet := []ContractWriteSet{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, ContractWriteSet{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	case APP_CONTRACT_INVOKE_REQUEST:
		payload, _ := cpyMsg.Payload.(*ContractInvokeRequestPayload)
		newPayload := ContractInvokeRequestPayload{
			ContractId: payload.ContractId,
			Args:       payload.Args,
			Timeout:    payload.Timeout,
		}
		msg.Payload = newPayload
	case APP_CONTRACT_INVOKE:
		payload, _ := cpyMsg.Payload.(*ContractInvokePayload)
		newPayload := ContractInvokePayload{
			ContractId: payload.ContractId,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Version: &StateVersion{
				Height: rs.Version.Height, TxIndex: rs.Version.TxIndex}})
		}
		writeSet := []ContractWriteSet{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, ContractWriteSet{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	case APP_SIGNATURE:
		payload, _ := cpyMsg.Payload.(*SignaturePayload)
		newPayload := SignaturePayload{}
		newPayload.Signatures = payload.Signatures
		msg.Payload = newPayload
	}

	return msg
}

func (msg *Message) CompareMessages(inMsg *Message) bool {
	if inMsg == nil || msg.App != inMsg.App {
		return false
	}
	switch msg.App {
	case APP_CONTRACT_TPL:
		payA, _ := msg.Payload.(*ContractTplPayload)
		payB, _ := inMsg.Payload.(*ContractTplPayload)
		return payA.Equal(payB)
	case APP_CONTRACT_DEPLOY:
		payA, _ := msg.Payload.(*ContractDeployPayload)
		payB, _ := inMsg.Payload.(*ContractDeployPayload)
		return payA.Equal(payB)
	case APP_CONTRACT_INVOKE:
		payA, _ := msg.Payload.(*ContractInvokePayload)
		payB, _ := inMsg.Payload.(*ContractInvokePayload)
		return payA.Equal(payB)
	case APP_CONTRACT_STOP:
		payA, _ := msg.Payload.(*ContractStopPayload)
		payB, _ := inMsg.Payload.(*ContractStopPayload)
		return payA.Equal(payB)
	case APP_SIGNATURE:
		//todo
		//payA, _ := msg.Payload.(*SignaturePayload)
		//payB, _ := inMsg.Payload.(*SignaturePayload)
		return true
	case APP_CONTRACT_TPL_REQUEST:
		payA, _ := msg.Payload.(*ContractInstallRequestPayload)
		payB, _ := inMsg.Payload.(*ContractInstallRequestPayload)
		return payA.Equal(payB)
	case APP_CONTRACT_DEPLOY_REQUEST:
		payA, _ := msg.Payload.(*ContractDeployRequestPayload)
		payB, _ := inMsg.Payload.(*ContractDeployRequestPayload)
		return payA.Equal(payB)
		//return reflect.DeepEqual(payA, payB)
	case APP_CONTRACT_INVOKE_REQUEST:
		payA, _ := msg.Payload.(*ContractInvokeRequestPayload)
		payB, _ := inMsg.Payload.(*ContractInvokeRequestPayload)
		return payA.Equal(payB)
	case APP_CONTRACT_STOP_REQUEST:
		payA, _ := msg.Payload.(*ContractStopRequestPayload)
		payB, _ := inMsg.Payload.(*ContractStopRequestPayload)
		return payA.Equal(payB)
	}
	return true
}

type ContractWriteSet struct {
	IsDelete   bool   `json:"is_delete"`
	Key        string `json:"key"`
	Value      []byte `json:"value"`
	ContractId []byte `json:"contract_id"`
}

func NewWriteSet(key string, value []byte) *ContractWriteSet {
	return &ContractWriteSet{Key: key, Value: value, IsDelete: false}
}
func ToPayloadMapValueBytes(data interface{}) []byte {
	b, err := rlp.EncodeToBytes(data)
	if err != nil {
		return nil
	}
	return b
}

type StateVersion struct {
	Height  *ChainIndex `json:"height"`
	TxIndex uint32      `json:"tx_index"`
}
type ContractStateValue struct {
	Value   []byte        `json:"value"`
	Version *StateVersion `json:"version"`
}

func (version *StateVersion) String() string {
	if version == nil {
		return `null`
	}
	if version.Height == nil {
		return fmt.Sprintf(`StateVersion[AssetId:{null}, Height:{null},TxIdx:{%d}]`, version.TxIndex)
	}
	return fmt.Sprintf(
		"StateVersion[AssetId:{%s}, Height:{%d},TxIdx:{%d}]",
		version.Height.AssetID.String(),
		version.Height.Index,
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

//(16+8)+4=28
func (version *StateVersion) Bytes() []byte {
	b := version.Height.Bytes()
	txIdx := make([]byte, 4)
	binary.LittleEndian.PutUint32(txIdx, version.TxIndex)
	b = append(b, txIdx...)
	return b[:]
}
func (version *StateVersion) SetBytes(b []byte) {
	cidx := &ChainIndex{}
	cidx.SetBytes(b[:24])
	version.Height = cidx
	version.TxIndex = binary.LittleEndian.Uint32(b[24:])
}

func (version *StateVersion) Equal(in *StateVersion) bool {
	rlpA, err := rlp.EncodeToBytes(version)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(in)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

const (
	FIELD_SPLIT_STR     = "^*^"
	FIELD_GENESIS_ASSET = "GenesisAsset"
)

type ContractError struct {
	Code    uint32 `json:"error_code"`    // error code
	Message string `json:"error_message"` // error data
}

//node election
type ElectionInf struct {
	EType     byte        `json:"election_type"` //vrf type, if set to 1, it is the assignation node
	AddrHash  common.Hash `json:"addr_hash"`     //common.Address将地址hash后，返回给请求节点
	Proof     []byte      `json:"proof"`         //vrf proof
	PublicKey []byte      `json:"public_key"`    //alg.PublicKey, rlp not support
}

type ElectionNode struct {
	JuryCount uint64        `json:"jury_count"` //
	EleList   []ElectionInf `json:"ele_list"`   //
}

type ContractReadSet struct {
	Key        string        `json:"key"`
	Version    *StateVersion `json:"version"`
	ContractId []byte        `json:"contract_id"`
}

//请求合约信息
type InvokeInfo struct {
	InvokeAddress common.Address  `json:"invoke_address"` //请求地址
	InvokeTokens  []*InvokeTokens `json:"invoke_tokens"`  //请求数量
	InvokeFees    *AmountAsset    `json:"invoke_fees"`    //请求交易�?
}

//请求的数量
type InvokeTokens struct {
	Amount  uint64 `json:"amount"`  //数量
	Asset   *Asset `json:"asset"`   //资产
	Address string `json:"address"` //接收地址
}

func (i *InvokeTokens) String() string {
	data, _ := json.Marshal(i)
	return string(data)
}

type TokenPayOut struct {
	Asset    *Asset
	Amount   uint64
	PayTo    common.Address
	LockTime uint32
}

//用户钱包发起的合约调用申请
type ContractInstallRequestPayload struct {
	TplName        string        `json:"tpl_name"`
	TplDescription string        `json:"tpl_description"`
	Path           string        `json:"install_path"`
	Version        string        `json:"tpl_version"`
	Abi            string        `json:"abi"`
	Language       string        `json:"language"`
	AddrHash       []common.Hash `json:"addr_hash"`
	Creator        string        `json:"creator"`
}

// Contract template deploy message
// App: contract_template
type ContractTplPayload struct {
	TemplateId []byte        `json:"template_id"`    // contract template id
	Size       uint16        `json:"size"`           // contract template bytecode size(Byte), use to compute transaction fee
	ByteCode   []byte        `json:"byte_code"`      // contract bytecode
	ErrMsg     ContractError `json:"contract_error"` // contract error message
}

// App: contract_deploy
type ContractDeployRequestPayload struct {
	TemplateId []byte   `json:"template_id"`
	Args       [][]byte `json:"args"`
	ExtData    []byte   `json:"extend_data"`
	Timeout    uint32   `json:"timeout"`
}

type ContractDeployPayload struct {
	TemplateId []byte             `json:"template_id"`   // delete--
	ContractId []byte             `json:"contract_id"`   // contract id
	Name       string             `json:"name"`          // the name for contract
	Args       [][]byte           `json:"args"`          // delete--
	EleNode    ElectionNode       `json:"election_node"` // contract jurors node info
	ReadSet    []ContractReadSet  `json:"read_set"`      // the set data of read, and value could be any type
	WriteSet   []ContractWriteSet `json:"write_set"`     // the set data of write, and value could be any type
	DuringTime uint64             `json:"during_time"`
	ErrMsg     ContractError      `json:"contract_error"` // contract error message
}

// Contract invoke message
// App: contract_invoke
type ContractInvokeRequestPayload struct {
	ContractId []byte   `json:"contract_id"` // contract id
	Args       [][]byte `json:"args"`        // contract arguments list
	Timeout    uint32   `json:"timeout"`
}

//如果是用户想修改自己的State信息，那么ContractId可以为空或�?0字节
type ContractInvokePayload struct {
	ContractId []byte             `json:"contract_id"`    // contract id
	Args       [][]byte           `json:"args"`           // delete--
	ReadSet    []ContractReadSet  `json:"read_set"`       // the set data of read, and value could be any type
	WriteSet   []ContractWriteSet `json:"write_set"`      // the set data of write, and value could be any type
	Payload    []byte             `json:"payload"`        // the contract execution result
	ErrMsg     ContractError      `json:"contract_error"` // contract error message
}

// App: contract_stop
type ContractStopRequestPayload struct {
	ContractId  []byte `json:"contract_id"`
	Txid        string `json:"transaction_id"`
	DeleteImage bool   `json:"delete_image"`
}

type ContractStopPayload struct {
	ContractId []byte             `json:"contract_id"`    // contract id
	ReadSet    []ContractReadSet  `json:"read_set"`       // the set data of read, and value could be any type
	WriteSet   []ContractWriteSet `json:"write_set"`      // the set data of write, and value could be any type
	ErrMsg     ContractError      `json:"contract_error"` // contract error message
}

//contract invoke result
type ContractInvokeResult struct {
	ContractId  []byte             `json:"contract_id"` // contract id
	RequestId   common.Hash        `json:"request_id"`
	Args        [][]byte           `json:"args"`           // contract arguments list
	ReadSet     []ContractReadSet  `json:"read_set"`       // the set data of read, and value could be any type
	WriteSet    []ContractWriteSet `json:"write_set"`      // the set data of write, and value could be any type
	Payload     []byte             `json:"payload"`        // the contract execution result
	TokenPayOut []*TokenPayOut     `json:"token_payout"`   //从合约地址付出Token
	TokenSupply []*TokenSupply     `json:"token_supply"`   //增发Token请求产生的结果
	TokenDefine *TokenDefine       `json:"token_define"`   //定义新Token
	ErrMsg      ContractError      `json:"contract_error"` // contract error message
}

type SignaturePayload struct {
	Signatures []SignatureSet `json:"signature_set"` // the array of signature
}

type SignatureSet struct {
	PubKey    []byte `json:"public_key"` //compress public key
	Signature []byte `json:"signature"`  //
}

func (ss SignatureSet) String() string {
	return fmt.Sprintf("SignatureSet: Pubkey[%x],Signature[%x]", ss.PubKey, ss.Signature)
}

// Token exchange message and verify message
// App: text
type DataPayload struct {
	MainData  []byte `json:"main_data"`
	ExtraData []byte `json:"extra_data"`
	Reference []byte `json:"reference"`
}

//一个地址对应的个人StateDB空间
type AccountStateUpdatePayload struct {
	WriteSet []AccountStateWriteSet `json:"write_set"`
}

type AccountStateWriteSet struct {
	IsDelete bool   `json:"is_delete"`
	Key      string `json:"key"`
	Value    []byte `json:"value"`
}

type FileInfo struct {
	UnitHash    common.Hash `json:"unit_hash"`
	UintHeight  uint64      `json:"unit_index"`
	ParentsHash common.Hash `json:"parents_hash"`
	Txid        common.Hash `json:"txid"`
	Timestamp   uint64      `json:"timestamp"`
	MainData    string      `json:"main_data"`
	ExtraData   string      `json:"extra_data"`
	Reference   string      `json:"reference"`
}

func NewContractTplPayload(templateId []byte, memory uint16, bytecode []byte, err ContractError) *ContractTplPayload {
	return &ContractTplPayload{
		TemplateId: templateId,
		Size:       memory,
		ByteCode:   bytecode,
		ErrMsg:     err,
	}
}

func NewContractDeployPayload(templateid []byte, contractid []byte, name string, args [][]byte,
	ele *ElectionNode, readset []ContractReadSet, writeset []ContractWriteSet, err ContractError) *ContractDeployPayload {
	payload := &ContractDeployPayload{
		ContractId: contractid,
		Name:       name,
		ReadSet:    readset,
		WriteSet:   writeset,
		ErrMsg:     err,
	}
	if ele != nil {
		payload.EleNode = *ele
	}
	return payload
}

func NewContractInvokePayload(contractid []byte, readset []ContractReadSet, writeset []ContractWriteSet,
	payload []byte, err ContractError) *ContractInvokePayload {
	return &ContractInvokePayload{
		ContractId: contractid,
		ReadSet:    readset,
		WriteSet:   writeset,
		Payload:    payload,
		ErrMsg:     err,
	}
}

func NewContractStopPayload(contractid []byte, readset []ContractReadSet, writeset []ContractWriteSet,
	err ContractError) *ContractStopPayload {
	return &ContractStopPayload{
		ContractId: contractid,
		ReadSet:    readset,
		WriteSet:   writeset,
		ErrMsg:     err,
	}
}

func (a *ElectionInf) Equal(b *ElectionInf) bool {
	if b == nil {
		return false
	}
	if !bytes.Equal(a.Proof, b.Proof) || !bytes.Equal(a.PublicKey, b.PublicKey) {
		return false
	}
	if !bytes.Equal(a.AddrHash[:], b.AddrHash[:]) {
		return false
	}
	return true
}

func (a *ContractReadSet) Equal(b *ContractReadSet) bool {
	if b == nil {
		return false
	}
	if !strings.EqualFold(a.Key, b.Key) || !bytes.Equal(a.ContractId, b.ContractId) {
		return false
	}
	if a.Version != nil && b.Version != nil {
		if a.Version.TxIndex != b.Version.TxIndex || a.Version.Height != b.Version.Height {
			return false
		}
		if a.Version.Height != nil {
			if !a.Version.Height.Equal(b.Version.Height) {
				return false
			}
		} else if b.Version.Height != nil {
			return false
		}
	} else if a.Version != b.Version {
		return false
	}

	return true
}

func (a *ContractWriteSet) Equal(b *ContractWriteSet) bool {
	if b == nil {
		return false
	}
	if !(a.IsDelete == b.IsDelete) || !strings.EqualFold(a.Key, b.Key) || !bytes.Equal(a.Value, b.Value) {
		return false
	}
	return true
}

func (a *ContractTplPayload) Equal(b *ContractTplPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

func (a *ContractDeployPayload) Equal(b *ContractDeployPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

func (a *ContractInvokePayload) Equal(b *ContractInvokePayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)
}

func (a *ContractStopPayload) Equal(b *ContractStopPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

func (a *ContractInstallRequestPayload) Equal(b *ContractInstallRequestPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

func (a *ContractDeployRequestPayload) Equal(b *ContractDeployRequestPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

func (a *ContractInvokeRequestPayload) Equal(b *ContractInvokeRequestPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

func (a *ContractStopRequestPayload) Equal(b *ContractStopRequestPayload) bool {
	rlpA, err := rlp.EncodeToBytes(a)
	if err != nil {
		return false
	}
	rlpB, err := rlp.EncodeToBytes(b)
	if err != nil {
		return false
	}
	return bytes.Equal(rlpA, rlpB)

}

type SysTokenIDInfo struct {
	CreateAddr     string
	TotalSupply    uint64
	LeastNum       uint64
	AssetID        string
	CreateTime     time.Time
	IsVoteEnd      bool
	SupportResults []*SysSupportResult
}
type SysSupportResult struct {
	TopicIndex  uint64
	TopicTitle  string
	VoteResults []*SysVoteResult
}
type SysVoteResult struct {
	SelectOption string
	Num          uint64
}
