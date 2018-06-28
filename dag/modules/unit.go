// unit package, unit structure and storage api
package modules

import (
	"encoding/json"
	"sync/atomic"
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

type Header struct {
	ParentUnits []common.Hash `json:"parent_units"`
	AssetIDs    []IDType      `json:"assets"`
	Authors     []Author      `json:"authors"` // the unit creation authors
	Witness     []Author      `json:"witness"`
	GasLimit    uint64        `json:"gasLimit"`
	GasUsed     uint64        `json:"gasUsed"`
	Root        common.Hash   `json:"root"`
	Number      ChainIndex    `json:"index"`
	Extra       []byte        `json:"extra"`
}

func NewHeader(parents []common.Hash, asset []IDType, gas, used uint64, extra []byte) *Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &Header{ParentUnits: hashs, AssetIDs: asset, GasLimit: gas, GasUsed: gas, Extra: append(b, extra...)}
}
func HeaderEqual(oldh, newh *Header) bool {
	if oldh.ParentUnits[0] == newh.ParentUnits[0] && oldh.ParentUnits[1] == newh.ParentUnits[1] {
		return true
	} else if oldh.ParentUnits[0] == newh.ParentUnits[1] && oldh.ParentUnits[1] == newh.ParentUnits[0] {
		return true
	}
	return false
}
func (h *Header) Index() uint64 {
	return h.Number.Index
}

// func (h *Header) ChainIndex() ChainIndex {
// 	return h.Number
// }

// key: unit.hash(unit)
type Unit struct {
	header *Header      `json:"unit_header"`  // unit header
	txs    Transactions `json:"transactions"` // transaction list

	hash         common.Hash        `json:"unit_hash"`     // unit hash
	size         common.StorageSize `json:"size"`          // unit size
	creationdate time.Time          `json:"creation_time"` // unit create time
}

type ChainIndex struct {
	AssetID IDType
	IsMain  bool
	Index   uint64
}

type Transactions []*Transaction

type Transaction struct {
	TxHash       common.Hash   `json:"tx_hash"`
	TxMessages   []UnitMessage `json:"messages"` //
	Authors      []Author      `json:"authors"`  // the issuers of the transaction
	CreationDate time.Time     `json:"creation_date"`
}

// key: message.hash(message+timestamp)
type UnitMessage struct {
	App          string       `json:"app"`          // message type
	PayloadHash  common.Hash  `json:"payload_hash"` // payload hash
	Memery       atomic.Value `json:"momery"`
	Excutiontime uint         `json:"excution_time"`
	Payload      interface{}  `json:"payload"` // the true transaction data
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

func NewUnit(header *Header, txs Transactions) *Unit {
	u := &Unit{
		header: CopyHeader(header),
		txs:    CopyTransactions(txs),
	}
	u.hash = u.Hash()
	u.size = u.Size()
	u.creationdate = time.Now()
	return u
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h

	if len(h.ParentUnits) > 0 {
		cpy.ParentUnits = make([]common.Hash, len(h.ParentUnits))
		for i := 0; i < len(h.ParentUnits); i++ {
			cpy.ParentUnits[i].Set(h.ParentUnits[i])
		}
	}

	if len(h.AssetIDs) > 0 {
		copy(cpy.AssetIDs, h.AssetIDs)
	}

	if len(h.Authors) > 0 {
		copy(cpy.Authors, h.Authors)
	}

	if len(h.Witness) > 0 {
		copy(cpy.Witness, h.Witness)
	}

	if len(h.Root) > 0 {
		cpy.Root.Set(h.Root)
	}

	return &cpy
}

func CopyTransactions(txs Transactions) Transactions {
	cpy := txs
	return cpy
}

type UnitNonce [8]byte

func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)/8)
}

// return  unit'hash
func (u *Unit) Hash() common.Hash {
	v := rlpHash(u)
	return v
}

func (u *Unit) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*u)) + common.StorageSize(len(u.hash)/8)
}

//
func (u *Unit) Header() *Header { return CopyHeader(u.header) }

//
//func (u *Unit) NumberU64() uint64 { return u.Head.Number.Uint64() }
func (u *Unit) Number() ChainIndex {
	return u.header.Number
}
func (u *Unit) NumberU64() uint64 {
	return u.header.Number.Index
}

// transactions
func (u *Unit) Transactions() []*Transaction {
	return u.txs
}
func txsTOPooltxs(txs Transactions) PoolTransactions {
	ptxs := make(PoolTransactions, 0)
	return ptxs

}
func (u *Unit) PoolTransactions() []*PoolTransaction {
	return txsTOPooltxs(u.txs)
}

//
func (u *Unit) ParentHash() []common.Hash {
	return u.header.ParentUnits
}
