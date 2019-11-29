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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"strings"
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

func (m *Message) GetApp() MessageType     { return m.App }
func (m *Message) GetPayload() interface{} { return m.Payload }
func (m *Message) GetPaymentPayLoad() *PaymentPayload {
	if m.App != APP_PAYMENT {
		return nil
	}
	return m.Payload.(*PaymentPayload)
}

// return message struct
func NewMessage(app MessageType, payload interface{}) *Message {
	m := new(Message)
	m.App = app
	m.Payload = payload
	return m
}

// message深拷贝
func CopyMessage(cpyMsg *Message) *Message {
	msg := *cpyMsg

	switch cpyMsg.App {
	default:
		msg.Payload = cpyMsg.Payload
	case APP_PAYMENT:
		payload, _ := cpyMsg.Payload.(*PaymentPayload)
		payment := *payload
		if len(payload.Inputs) > 0 {
			payment.Inputs = make([]*Input, 0)
			for _, in := range payload.Inputs {
				temp := *in
				temp.Extra = common.CopyBytes(in.Extra)
				temp.SignatureScript = common.CopyBytes(in.SignatureScript)
				if in.PreviousOutPoint != nil { // 创币message的previous outpoint 为空
					temp.PreviousOutPoint = &OutPoint{TxHash: in.PreviousOutPoint.TxHash, MessageIndex:
					in.PreviousOutPoint.MessageIndex, OutIndex: in.PreviousOutPoint.OutIndex}
				}
				payment.Inputs = append(payment.Inputs, &temp)
			}
		}
		if len(payload.Outputs) > 0 {
			payment.Outputs = make([]*Output, 0)
			for _, out := range payload.Outputs {
				temp := *out
				temp.Asset = &Asset{AssetId: out.Asset.AssetId, UniqueId: out.Asset.UniqueId}
				temp.PkScript = common.CopyBytes(out.PkScript)
				payment.Outputs = append(payment.Outputs, &temp)
			}
		}
		msg.Payload = &payment
	case APP_CONTRACT_TPL:
		payload, _ := cpyMsg.Payload.(*ContractTplPayload)
		temp_load := *payload
		temp_load.TemplateId = common.CopyBytes(payload.TemplateId)
		temp_load.ByteCode = common.CopyBytes(payload.ByteCode)
		msg.Payload = &temp_load

	case APP_CONTRACT_DEPLOY:
		payload, _ := cpyMsg.Payload.(*ContractDeployPayload)
		newPayload := ContractDeployPayload{
			TemplateId: common.CopyBytes(payload.TemplateId),
			ContractId: common.CopyBytes(payload.ContractId),
			Name:       payload.Name,
			DuringTime: payload.DuringTime,
			EleNode:    payload.EleNode,
			ErrMsg:     payload.ErrMsg,
		}
		if len(payload.Args) > 0 {
			newPayload.Args = make([][]byte, 0)
			newPayload.Args = append(newPayload.Args, payload.Args...)
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			version := *rs.Version
			if rs.Version != nil && rs.Version.Height != nil {
				version.Height.AssetID = rs.Version.Height.AssetID
				version.Height.Index = rs.Version.Height.Index
			}

			cs := ContractReadSet{Key: rs.Key, Version: &version, ContractId: common.CopyBytes(rs.ContractId)}
			readSet = append(readSet, cs)
		}

		writeSet := []ContractWriteSet{}
		for _, ws := range payload.WriteSet {
			cs := ContractWriteSet{Key: ws.Key, Value: common.CopyBytes(ws.Value), IsDelete: ws.IsDelete,
				ContractId: common.CopyBytes(ws.ContractId)}
			writeSet = append(writeSet, cs)
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = &newPayload
	case APP_CONTRACT_INVOKE:
		payload, _ := cpyMsg.Payload.(*ContractInvokePayload)
		newPayload := ContractInvokePayload{
			ContractId: common.CopyBytes(payload.ContractId),
			Payload:    common.CopyBytes(payload.Payload),
			ErrMsg:     payload.ErrMsg,
		}
		if len(payload.Args) > 0 {
			newPayload.Args = make([][]byte, 0)
			newPayload.Args = append(newPayload.Args, payload.Args...)
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			var version StateVersion
			if rs.Version != nil {
				version = *(rs.Version)
				if rs.Version != nil && rs.Version.Height != nil {
					version.Height.AssetID = rs.Version.Height.AssetID
					version.Height.Index = rs.Version.Height.Index
				}
			}

			cs := ContractReadSet{Key: rs.Key, Version: &version, ContractId: common.CopyBytes(rs.ContractId)}
			readSet = append(readSet, cs)
		}

		writeSet := []ContractWriteSet{}
		for _, ws := range payload.WriteSet {
			cs := ContractWriteSet{Key: ws.Key, Value: common.CopyBytes(ws.Value), IsDelete: ws.IsDelete,
				ContractId: common.CopyBytes(ws.ContractId)}
			writeSet = append(writeSet, cs)
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = &newPayload

	case APP_CONTRACT_STOP:
		payload, _ := cpyMsg.Payload.(*ContractStopPayload)
		temp_load := *payload
		temp_load.ContractId = common.CopyBytes(payload.ContractId)
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			version := *rs.Version
			if rs.Version != nil && rs.Version.Height != nil {
				version.Height.AssetID = rs.Version.Height.AssetID
				version.Height.Index = rs.Version.Height.Index
			}

			cs := ContractReadSet{Key: rs.Key, Version: &version, ContractId: common.CopyBytes(rs.ContractId)}
			readSet = append(readSet, cs)
		}

		writeSet := []ContractWriteSet{}
		for _, ws := range payload.WriteSet {
			cs := ContractWriteSet{Key: ws.Key, Value: common.CopyBytes(ws.Value), IsDelete: ws.IsDelete,
				ContractId: common.CopyBytes(ws.ContractId)}
			writeSet = append(writeSet, cs)
		}
		temp_load.ReadSet = readSet
		temp_load.WriteSet = writeSet
		msg.Payload = &temp_load

	case APP_SIGNATURE:
		payload, _ := cpyMsg.Payload.(*SignaturePayload)
		newPayload := SignaturePayload{}
		newPayload.Signatures = make([]SignatureSet, len(payload.Signatures))
		copy(newPayload.Signatures, payload.Signatures)
		msg.Payload = &newPayload
	case APP_DATA:
		payload, _ := cpyMsg.Payload.(*DataPayload)
		temp_load := *payload
		temp_load.MainData = common.CopyBytes(payload.MainData)
		temp_load.ExtraData = common.CopyBytes(payload.ExtraData)
		temp_load.Reference = common.CopyBytes(payload.Reference)
		msg.Payload = &temp_load
	case APP_ACCOUNT_UPDATE:
		payload, _ := cpyMsg.Payload.(*AccountStateUpdatePayload)
		temp_load := *payload
		temp_load.WriteSet = make([]AccountStateWriteSet, 0)
		for _, set := range payload.WriteSet {
			temp_load.WriteSet = append(temp_load.WriteSet, set)
		}
		msg.Payload = &temp_load
	case APP_CONTRACT_TPL_REQUEST:
		payload, _ := cpyMsg.Payload.(*ContractTplPayload)
		temp_load := *payload
		temp_load.TemplateId = common.CopyBytes(payload.TemplateId)
		temp_load.ByteCode = common.CopyBytes(payload.ByteCode)
		msg.Payload = &temp_load
	case APP_CONTRACT_DEPLOY_REQUEST:
		payload, _ := cpyMsg.Payload.(*ContractDeployRequestPayload)
		temp_load := *payload
		temp_load.TemplateId = common.CopyBytes(payload.TemplateId)
		temp_load.ExtData = common.CopyBytes(payload.ExtData)
		if len(payload.Args) > 0 {
			temp_load.Args = make([][]byte, 0)
			temp_load.Args = append(temp_load.Args, payload.Args...)
		}
		msg.Payload = &temp_load
	case APP_CONTRACT_INVOKE_REQUEST:
		payload, _ := cpyMsg.Payload.(*ContractInvokeRequestPayload)
		newPayload := ContractInvokeRequestPayload{
			ContractId: common.CopyBytes(payload.ContractId),
			Args:       payload.Args,
			Timeout:    payload.Timeout,
		}
		if len(payload.Args) > 0 {
			newPayload.Args = make([][]byte, 0)
			newPayload.Args = append(newPayload.Args, payload.Args...)
		}
		msg.Payload = &newPayload

	case APP_CONTRACT_STOP_REQUEST:
		payload, _ := cpyMsg.Payload.(*ContractStopRequestPayload)
		temp_load := *payload
		temp_load.ContractId = common.CopyBytes(payload.ContractId)
		msg.Payload = &temp_load
	}

	return &msg
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
	Version    *StateVersion `json:"version" rlp:"-"`
	ContractId []byte        `json:"contract_id"`
}
type contractReadSetTemp struct {
	Key        string  `json:"key"`
	AssetId    AssetId `json:"asset_id"`
	Index      uint64  `json:"chain_index"`
	TxIndex    uint32  `json:"tx_index"`
	ContractId []byte  `json:"contract_id"`
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
	CreateTime     int64
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
