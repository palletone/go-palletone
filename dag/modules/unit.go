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

// unit package, unit structure and storage api
package modules

import (
	"encoding/json"
	"math/big"
	"time"
	"unsafe"

	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"strings"
)

type Header struct {
	ParentsHash  []common.Hash   `json:"parents_hash"`
	AssetIDs     []IDType16      `json:"assets"`
	Authors      *Authentifier   `json:"author" rlp:"-"`  // the unit creation authors
	Witness      []*Authentifier `json:"witness" rlp:"-"` // 群签名
	TxRoot       common.Hash     `json:"root"`
	Number       ChainIndex      `json:"index"`
	Extra        []byte          `json:"extra"`
	Creationdate int64           `json:"creation_time"` // unit create time
	//FeeLimit    uint64        `json:"fee_limit"`
	//FeeUsed     uint64        `json:"fee_used"`
}

func (cpy *Header) CopyHeader(h *Header) {
	cpy = h
	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i] = h.ParentsHash[i]
		}
	}

	if len(h.AssetIDs) > 0 {
		cpy.AssetIDs = make([]IDType16, len(h.AssetIDs))
		for i := 0; i < len(h.AssetIDs); i++ {
			cpy.AssetIDs[i] = h.AssetIDs[i]
		}
	}

}

func NewHeader(parents []common.Hash, asset []IDType16, used uint64, extra []byte) *Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...)}
}

func HeaderEqual(oldh, newh *Header) bool {
	if oldh.ParentsHash[0] == newh.ParentsHash[0] && oldh.ParentsHash[1] == newh.ParentsHash[1] {
		return true
	} else if oldh.ParentsHash[0] == newh.ParentsHash[1] && oldh.ParentsHash[1] == newh.ParentsHash[0] {
		return true
	}
	return false
}

func (h *Header) Index() uint64 {
	return h.Number.Index
}
func (h *Header) ChainIndex() ChainIndex {
	return h.Number
}

func (h *Header) Hash() common.Hash {
	emptyHeader := CopyHeader(h)
	emptyHeader.Authors = nil
	emptyHeader.Witness = []*Authentifier{}
	return rlp.RlpHash(emptyHeader)
}

func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)/8)
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h

	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i].Set(h.ParentsHash[i])
		}
	}

	if len(h.AssetIDs) > 0 {
		copy(cpy.AssetIDs, h.AssetIDs)
	}

	if len(h.Witness) > 0 {
		copy(cpy.Witness, h.Witness)
	}

	if len(h.TxRoot) > 0 {
		cpy.TxRoot.Set(h.TxRoot)
	}

	return &cpy
}

func (u *Unit) CopyBody(txs Transactions) Transactions {
	if len(txs) > 0 {
		u.Txs = make([]*Transaction, len(txs))
		for i, pTx := range txs {
			tx := Transaction{
				AccountNonce: pTx.AccountNonce,
				TxHash:       pTx.TxHash,
				From:         pTx.From,
				Excutiontime: pTx.Excutiontime,
				Memery:       pTx.Memery,
				CreationDate: pTx.CreationDate,
				TxFee:        pTx.TxFee,
				Txsize:       pTx.Txsize,
			}
			if len(pTx.TxMessages) > 0 {
				tx.TxMessages = make([]Message, len(pTx.TxMessages))
				for j := 0; j < len(pTx.TxMessages); j++ {
					tx.TxMessages[j] = pTx.TxMessages[j]
				}
			}
			u.Txs[i] = &tx
		}
	}
	return u.Txs
}

//wangjiyou add for ptn/fetcher.go
type Units []*Unit

// key: unit.UnitHash(unit)
type Unit struct {
	UnitHeader *Header            `json:"unit_header"`  // unit header
	Txs        Transactions       `json:"transactions"` // transaction list
	UnitHash   common.Hash        `json:"unit_hash"`    // unit hash
	UnitSize   common.StorageSize `json:"UnitSize"`     // unit size
	// These fields are used by package ptn to track
	// inter-peer block relay.
	ReceivedAt   time.Time
	ReceivedFrom interface{}
}

func (unit *Unit) IsEmpty() bool {
	if unit == nil || len(unit.Txs) <= 0 {
		return true
	}
	return false
}

type Transactions []*Transaction

type Transaction struct {
	AccountNonce uint64
	TxHash       common.Hash        `json:"txhash" rlp:""`
	TxMessages   []Message          `json:"messages"` //
	From         *Authentifier      `json:"authors"`  // the issuers of the transaction
	Excutiontime uint               `json:"excution_time"`
	Memery       uint               `json:"memory"`
	CreationDate string             `json:"creation_date"`
	TxFee        *big.Int           `json:"txfee"` // user set total transaction fee.
	Txsize       common.StorageSize `json:"txsize" rlp:""`
	Locktime     uint32             `json:"lock_time"`
	Priority_lvl float64            `json:"priority_lvl"` // 打包的优先级
}

type ChainIndex struct {
	AssetID IDType16
	IsMain  bool
	Index   uint64
}

func (height ChainIndex) String() string {
	data, err := rlp.EncodeToBytes(height)
	if err != nil {
		return ""
	}
	return string(data)
}

var (
	APP_PAYMENT         = "payment"
	APP_CONTRACT_TPL    = "contract_template"
	APP_CONTRACT_DEPLOY = "contract_deploy"
	APP_CONTRACT_INVOKE = "contract_invoke"
	APP_CONFIG          = "config"
	APP_TEXT            = "text"
)

// key: message.UnitHash(message+timestamp)
type Message struct {
	App         string      `json:"app"`          // message type
	PayloadHash common.Hash `json:"payload_hash"` // payload hash
	Payload     interface{} `json:"payload"`      // the true transaction data
}

func (msg *Message) CopyMessages(cpyMsg *Message) *Message {
	msg.App = cpyMsg.App
	msg.PayloadHash = cpyMsg.PayloadHash
	msg.Payload = cpyMsg.Payload
	switch cpyMsg.App {
	case APP_PAYMENT, APP_CONTRACT_TPL, APP_TEXT:
		msg.Payload = cpyMsg.Payload
	case APP_CONFIG:
		payload, _ := cpyMsg.Payload.(ConfigPayload)
		newPayload := ConfigPayload{
			ConfigSet: []PayloadMapStruct{},
		}
		for _, p := range payload.ConfigSet {
			newPayload.ConfigSet = append(newPayload.ConfigSet, PayloadMapStruct{Key: p.Key, Value: p.Value})
		}
		msg.Payload = newPayload
	case APP_CONTRACT_DEPLOY:
		payload, _ := cpyMsg.Payload.(ContractDeployPayload)
		newPayload := ContractDeployPayload{
			TemplateId: payload.TemplateId,
			ContractId: payload.ContractId,
			Config:     payload.Config,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Value: &StateVersion{Height: rs.Value.Height, TxIndex: rs.Value.TxIndex}})
		}
		writeSet := []PayloadMapStruct{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	case APP_CONTRACT_INVOKE:
		payload, _ := cpyMsg.Payload.(ContractInvokePayload)
		newPayload := ContractInvokePayload{
			ContractId: payload.ContractId,
			Function:   payload.Function,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Value: &StateVersion{Height: rs.Value.Height, TxIndex: rs.Value.TxIndex}})
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

/************************** Payload Details ******************************************/
type PayloadMapStruct struct {
	Key   string
	Value interface{}
}

// Token exchange message and verify message
// App: payment
type PaymentPayload struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type StateVersion struct {
	Height  ChainIndex
	TxIndex uint32
}

func (version *StateVersion) String() string {
	data, err := rlp.EncodeToBytes(*version)
	if err != nil {
		return ""
	}
	return string(data)
}

func (version *StateVersion) ParseStringKey(key string) {
	ss := strings.Split(key, "_")
	if len(ss) != 2 {
		return
	}
	var v StateVersion
	if err := rlp.DecodeBytes([]byte(ss[1]), &v); err != nil {
		return
	}
	version = &v
}

// Contract template deploy message
// App: contract_template
type ContractTplPayload struct {
	TemplateId common.Hash `json:"template_id"` // configure xml file of contract
	Bytecode   []byte      `json:"bytecode"`    // contract bytecode
}

type DelContractState struct {
	IsDelete bool
}

type ContractReadSet struct {
	Key   string
	Value *StateVersion
}

// Contract instance message
// App: contract_deploy
type ContractDeployPayload struct {
	TemplateId common.Hash        `json:"template_id"` // contract template id
	ContractId string             `json:"contract_id"` // contract id
	Config     []byte             `json:"config"`      // configure xml file of contract instance parameters
	ReadSet    []ContractReadSet  `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   []PayloadMapStruct `json:"write_set"`   // the set data of write, and value could be any type
}

// Contract invoke message
// App: contract_invoke
type ContractInvokePayload struct {
	ContractId string             `json:"contract_id"` // contract id
	Function   []byte             `json:"function"`    // serialized value of invoked function with call parameters
	ReadSet    []ContractReadSet  `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   []PayloadMapStruct `json:"write_set"`   // the set data of write, and value could be any type
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

/************************** End of Payload Details ******************************************/

type Author struct {
	Address        common.Address `json:"address"`
	Pubkey         []byte/*common.Hash*/ `json:"pubkey"`
	TxAuthentifier Authentifier `json:"authentifiers"`
}

type Authentifier struct {
	Address string `json:"address"`
	R       []byte `json:"r"`
	S       []byte `json:"s"`
	V       []byte `json:"v"`
}

func (a *Authentifier) ToDB() ([]byte, error) {
	return json.Marshal(a)
}
func (a *Authentifier) FromDB(info []byte) error {
	return json.Unmarshal(info, a)
}

func NewUnit(header *Header, txs Transactions) *Unit {
	u := &Unit{
		UnitHeader: CopyHeader(header),
		Txs:        CopyTransactions(txs),
	}
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}

func CopyTransactions(txs Transactions) Transactions {
	cpy := txs
	return cpy
}

type UnitNonce [8]byte

/************************** Unit Members  *****************************/
func (u *Unit) Header() *Header { return CopyHeader(u.UnitHeader) }

// transactions
func (u *Unit) Transactions() []*Transaction {
	return u.Txs
}

// return transaction
func (u *Unit) Transaction(hash common.Hash) *Transaction {
	for _, transaction := range u.Txs {
		if transaction.TxHash == hash {
			return transaction
		}
	}
	return nil
}

// return  unit'UnitHash
func (u *Unit) Hash() common.Hash {
	return u.UnitHeader.Hash()
}

func (u *Unit) Size() common.StorageSize {
	emptyUnit := Unit{}
	emptyUnit.UnitHeader = CopyHeader(u.UnitHeader)
	emptyUnit.UnitHeader.Authors = nil
	emptyUnit.UnitHeader.Witness = []*Authentifier{}
	emptyUnit.CopyBody(u.Txs)

	b, err := rlp.EncodeToBytes(emptyUnit)
	if err != nil {
		return common.StorageSize(0)
	} else {
		return common.StorageSize(len(b))
	}
}

// return Creationdate
// comment by Albert·Gou
//func (u *Unit) CreationDate() time.Time {
//	return u.UnitHeader.Creationdate
//}

//func (u *Unit) NumberU64() uint64 { return u.Head.Number.Uint64() }
func (u *Unit) Number() ChainIndex {
	return u.UnitHeader.Number
}
func (u *Unit) NumberU64() uint64 {
	return u.UnitHeader.Number.Index
}

// return unit's parents UnitHash
func (u *Unit) ParentHash() []common.Hash {
	return u.UnitHeader.ParentsHash
}

type ErrUnit float64

func (e ErrUnit) Error() string {
	switch e {
	case -1:
		return "Unit size error"
	case -2:
		return "Unit signature error"
	case -3:
		return "Unit header save error"
	case -4:
		return "Unit tx size error"
	case -5:
		return "Save create token transaction error"
	case -6:
		return "Save config transaction error"
	default:
		return ""
	}
	return ""
}

/************************** Unit Members  *****************************/

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewUnitWithHeader(header *Header) *Unit {
	return &Unit{UnitHeader: CopyHeader(header)}
}

// WithBody returns a new block with the given transaction and uncle contents.
func (b *Unit) WithBody(transactions []*Transaction) *Unit {
	// check transactions merkle root
	txs := b.CopyBody(transactions)
	root := core.DeriveSha(txs)
	if strings.Compare(root.String(), b.UnitHeader.TxRoot.String()) != 0 {
		return nil
	}
	// set unit body
	b.Txs = b.CopyBody(txs)
	return b
}

func (u *Unit) ContainsParent(pHash common.Hash) bool {
	ps := pHash.String()
	for _, hash := range u.UnitHeader.ParentsHash {
		if strings.Compare(hash.String(), ps) == 0 {
			return true
		}
	}
	return false
}

func RSVtoAddress(tx *Transaction) common.Address {
	sig := make([]byte, 65)
	copy(sig[32-len(tx.From.R):32], tx.From.R)
	copy(sig[64-len(tx.From.S):64], tx.From.S)
	copy(sig[64:], tx.From.V)
	pub, _ := crypto.SigToPub(tx.TxHash[:], sig)
	address := crypto.PubkeyToAddress(*pub)
	return address
}

func RSVtoPublicKey(hash, r, s, v []byte) (*ecdsa.PublicKey, error) {
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	copy(sig[64:], v)
	return crypto.SigToPub(hash, sig)
}

type TxValidationCode int32

const (
	TxValidationCode_VALID                        TxValidationCode = 0
	TxValidationCode_INVALID_CONTRACT_TEMPLATE    TxValidationCode = 1
	TxValidationCode_INVALID_FEE                  TxValidationCode = 2
	TxValidationCode_BAD_COMMON_HEADER            TxValidationCode = 3
	TxValidationCode_BAD_CREATOR_SIGNATURE        TxValidationCode = 4
	TxValidationCode_INVALID_ENDORSER_TRANSACTION TxValidationCode = 5
	TxValidationCode_INVALID_CONFIG_TRANSACTION   TxValidationCode = 6
	TxValidationCode_UNSUPPORTED_TX_PAYLOAD       TxValidationCode = 7
	TxValidationCode_BAD_PROPOSAL_TXID            TxValidationCode = 8
	TxValidationCode_DUPLICATE_TXID               TxValidationCode = 9
	TxValidationCode_ENDORSEMENT_POLICY_FAILURE   TxValidationCode = 10
	TxValidationCode_MVCC_READ_CONFLICT           TxValidationCode = 11
	TxValidationCode_PHANTOM_READ_CONFLICT        TxValidationCode = 12
	TxValidationCode_UNKNOWN_TX_TYPE              TxValidationCode = 13
	TxValidationCode_TARGET_CHAIN_NOT_FOUND       TxValidationCode = 14
	TxValidationCode_MARSHAL_TX_ERROR             TxValidationCode = 15
	TxValidationCode_NIL_TXACTION                 TxValidationCode = 16
	TxValidationCode_EXPIRED_CHAINCODE            TxValidationCode = 17
	TxValidationCode_CHAINCODE_VERSION_CONFLICT   TxValidationCode = 18
	TxValidationCode_BAD_HEADER_EXTENSION         TxValidationCode = 19
	TxValidationCode_BAD_CHANNEL_HEADER           TxValidationCode = 20
	TxValidationCode_BAD_RESPONSE_PAYLOAD         TxValidationCode = 21
	TxValidationCode_BAD_RWSET                    TxValidationCode = 22
	TxValidationCode_ILLEGAL_WRITESET             TxValidationCode = 23
	TxValidationCode_INVALID_WRITESET             TxValidationCode = 24
	TxValidationCode_NOT_VALIDATED                TxValidationCode = 254
	TxValidationCode_NOT_COMPARE_SIZE             TxValidationCode = 255
	TxValidationCode_INVALID_OTHER_REASON         TxValidationCode = 256
)

var TxValidationCode_name = map[int32]string{
	0:   "VALID",
	1:   "INVALID_CONTRACT_TEMPLATE",
	2:   "INVALID_FEE",
	3:   "BAD_COMMON_HEADER",
	4:   "BAD_CREATOR_SIGNATURE",
	5:   "INVALID_ENDORSER_TRANSACTION",
	6:   "INVALID_CONFIG_TRANSACTION",
	7:   "UNSUPPORTED_TX_PAYLOAD",
	8:   "BAD_PROPOSAL_TXID",
	9:   "DUPLICATE_TXID",
	10:  "ENDORSEMENT_POLICY_FAILURE",
	11:  "MVCC_READ_CONFLICT",
	12:  "PHANTOM_READ_CONFLICT",
	13:  "UNKNOWN_TX_TYPE",
	14:  "TARGET_CHAIN_NOT_FOUND",
	15:  "MARSHAL_TX_ERROR",
	16:  "NIL_TXACTION",
	17:  "EXPIRED_CHAINCODE",
	18:  "CHAINCODE_VERSION_CONFLICT",
	19:  "BAD_HEADER_EXTENSION",
	20:  "BAD_CHANNEL_HEADER",
	21:  "BAD_RESPONSE_PAYLOAD",
	22:  "BAD_RWSET",
	23:  "ILLEGAL_WRITESET",
	24:  "INVALID_WRITESET",
	254: "NOT_VALIDATED",
	255: "NOT_COMPARE_SIZE",
	256: "INVALID_OTHER_REASON",
}
