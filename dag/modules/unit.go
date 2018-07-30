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
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
)

/*****************************27 June, 2018 update unit struct type*****************************************/
//type Unit struct {
//	Unit                  string          `json:"unit"`                     // unit hash
//	Version               string          `json:"version"`                  // 版本号
//	Alt                   string          `json:"alt"`                      // 资产号
//	Messages              []Message       `json:"messages"`                 // 消息
//	Authors               []Author        `json:"authors"`                  // 发起人
//	ParentsHash           []string        `json:"parents_hash"`             // 父单元数组
//	CreationDate          time.Time       `json:"creation_date"`            // 创建时间
//	LastPacket            string          `json:"last_packet"`              // 最后一个packet
//	LastPacketUnit        string          `json:"last_packet_unit"`         // 最后一个packet对应的unit
//	WitnessListUnit       string          `json:"witness_list_unit"`        // 上一个稳定见证单元的hash
//	ContentHash           string          `json:"content_hash"`             // 内容hash
//	IsFree                bool            `json:"is_free"`                  // 顶端单元
//	IsOnMainChain         bool            `json:"is_on_main_chain"`         // 是在主链上
//	MainChainIndex        uint64          `json:"main_chain_index"`         // 主链序号
//	LatestIncludedMcIndex uint64          `json:"latest_included_mc_index"` // 最新的主链序号
//	Level                 uint64          `json:"level"`                    // 单元级别
//	WitnessedLevel        uint64          `json:"witness_level"`            // 见证级别
//	IsStable              bool            `json:"is_stable"`                // 是否稳定
//	Sequence              string          `json:"sequence"`                 // {枚举：'good' 'temp-bad' 'final-bad', default:'good'}
//	BestParentUnit        string          `json:"best_parent_unit"`         // 最优父单元
//
//	// 头佣金和净载荷酬劳在我们的项目可能暂时没有用到
//	// fields of  'headers_commission' and 'headers_commission' may not be used in our project for the now
//	// HeadersCommission     int             `json:"headers_commission"`       // 头佣金
//	// PayloadCommission     int             `json:"headers_commission"`       // 净载荷酬劳
//
//	// 与杨杰沟通，这两个字段表示前驱和后继，但是从DAG网络和数据库update两方面考虑，暂时不需要这两个字段
//	// In communication with Yang Jie, these two fields represent the precursor and successor, but considering the DAG network and the database update, the two fields are not needed temporarily.
//	// ToUnit                map[string]bool `json:"to_unit"`                  // parents
//	// FromUnit              map[string]bool `json:"from_unit"`                // childs
//
//	// 与杨杰沟通，当时在未确定 数据库的时候考虑外键、模糊查询的情况
//	// Communicating with Yang Jie, then ha has considered foreign keys and fuzzy queries at the time when the database was not determined
//	// Key                   string          `json:"key"`                      // index: key
//}
/***************************** end of update **********************************************/

type Header struct {
	ParentsHash  []common.Hash   `json:"parents_hash"`
	AssetIDs     []IDType16      `json:"assets"`
	Authors      *Authentifier   `json:"author" rlp:"-"`  // the unit creation authors
	Witness      []*Authentifier `json:"witness" rlp:"-"` // 群签名
	TxRoot       common.Hash     `json:"root"`
	Number       ChainIndex      `json:"index"`
	Extra        []byte          `json:"extra"`
	Creationdate int64       `json:"creation_time"` // unit create time
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
	return rlp.RlpHash(h)
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

// key: unit.UnitHash(unit)
type Unit struct {
	UnitHeader *Header            `json:"unit_header"`  // unit header
	Txs        Transactions       `json:"transactions"` // transaction list
	UnitHash   common.Hash        `json:"unit_hash"`    // unit hash
	UnitSize   common.StorageSize `json:"UnitSize"`     // unit size
	//Gasprice     uint64             `json:"gas_price"`     // user set total gas
	//Gasused      uint64             `json:"gas_used"`      // the actually used gas, mediator set
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
	Priority_lvl float64            `json:"priority_lvl"`
}

type ChainIndex struct {
	AssetID IDType16
	IsMain  bool
	Index   uint64
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

/************************** Payload Details ******************************************/
// Token exchange message and verify message
// App: payment
type PaymentPayload struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

// Contract template deploy message
// App: contract_template
type ContractTplPayload struct {
	TemplateId common.Hash            `json:"template_id"` // configure xml file of contract
	Bytecode   []byte                 `json:"bytecode"`    // contract bytecode
	ReadSet    map[string]interface{} `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   map[string]interface{} `json:"write_set"`   // the set data of write, and value could be any type

}

// Contract instance message
// App: contract_deploy
type ContractDeployPayload struct {
	TemplateId common.Hash            `json:"template_id"` // contract template id
	Config     []byte                 `json:"config"`      // configure xml file of contract instance parameters
	ReadSet    map[string]interface{} `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   map[string]interface{} `json:"write_set"`   // the set data of write, and value could be any type
}

// Contract invoke message
// App: contract_invoke
type ContractInvokePayload struct {
	ContractId string                 `json:"contract_id"` // contract id
	Function   []byte                 `json:"function"`    // serialized value of invoked function with call parameters
	ReadSet    map[string]interface{} `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   map[string]interface{} `json:"write_set"`   // the set data of write, and value could be any type
}

// Token exchange message and verify message
// App: config	// update global config
type ConfigPayload struct {
	ConfigSet map[string]interface{} `json:"config_set"` // the array of global config
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
	u.UnitSize = header.Size()
	u.UnitHash = u.Hash()
	return u
}

// comment by Albert·Gou
//func NewGenesisUnit(genesisConf *core.Genesis, txs Transactions) (*Unit, error) {
//	//test
//	unit := Unit{Txs: txs}
//	return &unit, nil
//}

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
	v := rlp.RlpHash(u)
	return v
}

func (u *Unit) Size() common.StorageSize {
	//u.UnitSize = common.StorageSize(unsafe.Sizeof(*u)) + common.StorageSize(len(u.UnitHash)/8)
	//return u.UnitSize

	b, err := rlp.EncodeToBytes(u)
	if err != nil {
		return common.StorageSize(0)
	} else {
		return common.StorageSize(len(b))
	}
	// if UnitSize := b.UnitSize.Load(); UnitSize != nil {
	// 	return UnitSize.(common.StorageSize)
	// }
	// c := writeCounter(0)
	// rlp.Encode(&c, b)
	// b.UnitSize.Store(common.StorageSize(c))
	// return common.StorageSize(c)
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
