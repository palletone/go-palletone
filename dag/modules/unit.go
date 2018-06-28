// unit package, unit structure and storage api
package modules

import (
	"encoding/json"
	"math/big"
	"time"
	"unsafe"

	"github.com/palletone/go-palletone/common"
)

/*****************************27 June, 2018 update unit struct type*****************************************/
//type Unit struct {
//	Unit                  string          `json:"unit"`                     // unit hash
//	Version               string          `json:"version"`                  // 版本号
//	Alt                   string          `json:"alt"`                      // 资产号
//	Messages              []Message       `json:"messages"`                 // 消息
//	Authors               []Author        `json:"authors"`                  // 发起人
//	ParentUnits           []string        `json:"parent_units"`             // 父单元数组
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

// key: unit.hash(unit)
type Unit struct {
	UnitHeader *Header      `json:"unit_header"`  // unit header
	Txs        Transactions `json:"transactions"` // transaction list

	UnitHash     common.Hash `json:"unit_hash"`   // unit hash
	Size         uint64      `json:"size"`        // unit size
	CreationDate time.Time   `json:"create_time"` // unit create time
}

type Header struct {
	ParentUnit  common.Hash `json:"parent_unit"`
	AssetIDs    []IDType    `json:"assets"`
	Authors     []Author    `json:"authors"` // the unit creation authors
	Witness     []Author    `json:"witness"`
	GasLimit    uint64      `json:"gasLimit"`
	GasUsed     uint64      `json:"gasUsed"`
	Root        common.Hash `json:"state_root"`
	TxHash      common.Hash `json:"transactions_root"`
	ReceiptHash common.Hash `json:"receipt_root"`
	Time        *big.Int
	Index       ChainIndex `json:"index"`
	Extra       []byte     `json:"extra"`
}

type ChainIndex struct {
	AssetID IDType
	IsMain  bool
	Index   uint64
}

type Transactions []*Transaction

type Transaction struct {
	TxHash       common.Hash `json:"tx_hash"`
	TxMessages   []Message   `json:"messages"` //
	Authors      []Author    `json:"authors"`  // the issuers of the transaction
	CreationDate time.Time   `json:"creation_date"`
}

// key: message.hash(message+timestamp)
type Message struct {
	App         string      `json:"app"`          // message type
	PayloadHash common.Hash `json:"payload_hash"` // payload hash
	Payload     interface{} `json:"payload"`      // the true transaction data
}

/************************** Payload Details ******************************************/
// type Payload struct {
// 	ppay   PaymentPayload
// 	ctpay  ContractTplPayload
// 	cdpay  ContractDeployPayload
// 	cipay  ContractInvokePayload
// 	conpay ConfigPayload
// 	txtpay TextPayload
// }

// Token exchange message and verify message
// App: payment
// App: verify

type PaymentPayload struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

// Contract template deploy message
// App: contract_template
type ContractTplPayload struct {
	TemplateId string                 `json:"template_id"` // configure xml file of contract
	Bytecode   []byte                 `json:"bytecode"`    // contract bytecode
	ReadSet    map[string]interface{} `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   map[string]interface{} `json:"write_set"`   // the set data of write, and value could be any type

}

// Contract instance message
// App: contract_deploy
type ContractDeployPayload struct {
	TemplateId string                 `json:"template_id"` // contract template id
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
	Address        common.Hash  `json:"address"`
	Pubkey         common.Hash  `json:"pubkey"`
	TxAuthentifier Authentifier `json:"authentifiers"`
}

type Authentifier struct {
	R string `json:"r"`
}

func (a *Authentifier) ToDB() ([]byte, error) {
	return json.Marshal(a)
}
func (a *Authentifier) FromDB(info []byte) error {
	return json.Unmarshal(info, a)
}

type UnitNonce [8]byte

func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)/8)
}

func NewUnit(header *Header, msgs []*Message, receipt []*Receipt) *Unit {
	u := &Unit{
		UnitHeader: CopyHeader(header),
	}

	u.CreationDate = time.Now()
	return u
}

// return  unit'hash
func (u *Unit) Hash() common.Hash {
	v := rlpHash(u)
	return v
}

//
func (u *Unit) Header() *Header { return CopyHeader(u.UnitHeader) }

//  last unit
func CurrentUnit() *Unit {
	return &Unit{CreationDate: time.Now()}
}

// get unit  with hash & number or level.
func GetUnit(hash common.Hash, number uint64) *Unit {
	return &Unit{CreationDate: time.Now()}
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(big.Int); h.Time != nil {
		cpy.Time.Set(h.Time)
	}
	// if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
	// 	cpy.Difficulty.Set(h.Difficulty)
	// }
	// if cpy.Number = new(big.Int); h.Number != nil {
	// 	cpy.Number.Set(h.Number)
	// }
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

//
// func (u *Unit) NumberU64() uint64 { return u.UnitHeader.Number.Uint64() }

// transactions
func (u *Unit) Transactions() []*Transaction {
	ts := make([]*Transaction, 0)
	return ts
}

//
func (u *Unit) ParentUnit() common.Hash {
	return u.UnitHeader.ParentUnit
}
