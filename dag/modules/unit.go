// unit package, unit structure and storage api
package modules

import (
<<<<<<< HEAD
	"time"
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
	UnitHeader  	Header			`json:"unit_header"`	// unit header
	Transactions 	Transactions	`json:"transactions"`	// transaction list

	UnitHash 		Hash			`json:"unit_hash"`		// unit hash
	Size 			uint64			`json:"size"`			// unit size
	CreateTime   	time.Time		`json:"create_time"`	// unit create time
}

type Header struct {
	ParentUnits		Hash			`json:"parent_units"`
	AssetIDs		[]IDType		`json:"assets"`
	Authors			[]Author		`json:"authors"`		// the unit creation authors
	Witness			[]Author		`json:"witness"`
	GasLimit    	uint64         	`json:"gasLimit"`
	GasUsed    	 	uint64         	`json:"gasUsed"`
	Root			Hash			`json:"root"`
	Index      		ChainIndex      `json:"index"`
}

type ChainIndex struct {
	AssetID			IDType
	IsMain			bool
	Index			uint64
}

type Transactions []*Transaction

type Transaction struct {
	TxHash			Hash		  `json:"tx_hash"`
	TxMessages		[]Message	  `json:"messages"`			//
	Authors			[]Author	  `json:"authors"`			// the issuers of the transaction
	CreationDate    time.Time     `json:"creation_date"`
}

// key: message.hash(message+timestamp)
type Message struct {
	App             string        `json:"app"`              // message type
	PayloadHash     Hash    	  `json:"payload_hash"`     // payload hash
	Payload         interface{}   `json:"payload"`			// the true transaction data
}

/************************** Payload Details ******************************************/
=======
	"encoding/json"
	"math/big"
	"time"
	"unsafe"

	"github.com/palletone/go-palletone/common"
)

type UnitNonce [8]byte
type Header struct {
	ParentHash  common.Hash    `json:"parent_hash"`
	Coinbase    common.Address `json:"coinbase"`
	Root        common.Hash    `json:"state_root"`
	TxHash      common.Hash    `json:"transactions_root"`
	ReceiptHash common.Hash    `json:"receipt_root"`
	Price       uint64         `json:"tx_price"`
	Alias       string         `json:""`
	Difficulty  *big.Int       `json:"difficulty"`
	Number      *big.Int       `json:"number"`
	Time        *big.Int       `json:"timestamp"`
	Extra       []byte         `json:"extraData"`
	MixDigest   common.Hash    `json:"mixHash"`
	Nonce       UnitNonce      `json:"nonce"`
}

func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)+(h.Difficulty.BitLen()+h.Number.BitLen()+h.Time.BitLen())/8)
}

// key: unit.hash(unit)
type Unit struct {
	Unit     string    `json:"unit"`    // unit hash
	Version  string    `json:"version"` // 版本号
	Alt      string    `json:"alt"`     // 资产号
	Head     *Header   `json:"header"`
	Messages []Message `json:"messages"` // 消息
	Authors  []Author  `json:"authors"`  // 发起人
	// ParentUnits     []string  `json:"parent_units"`      // 父单元数组
	CreationDate    time.Time `json:"creation_date"`     // 创建时间
	LastPacket      string    `json:"last_packet"`       // 最后一个packet
	LastPacketUnit  string    `json:"last_packet_unit"`  // 最后一个packet对应的unit
	WitnessListUnit string    `json:"witness_list_unit"` // 上一个稳定见证单元的hash
	ContentHash     string    `json:"content_hash"`      // 内容hash
	// 头佣金和净载荷酬劳在我们的项目可能暂时没有用到
	// fields of  'headers_commission' and 'headers_commission' may not be used in our project for the now
	// HeadersCommission     int             `json:"headers_commission"`       // 头佣金
	// PayloadCommission     int             `json:"headers_commission"`       // 净载荷酬劳
	// 改为单链
	// IsFree                bool   `json:"is_free"`                  // 顶端单元
	// IsOnMainChain         bool   `json:"is_on_main_chain"`         // 是在主链上
	// MainChainIndex        int64  `json:"main_chain_index"`         // 主链序号
	// LatestIncludedMcIndex uint64 `json:"latest_included_mc_index"` // 最新的主链序号
	Level          uint64 `json:"level"`            // 单元级别
	WitnessedLevel uint64 `json:"witness_level"`    // 见证级别
	IsStable       bool   `json:"is_stable"`        // 是否稳定
	Sequence       string `json:"sequence"`         // {枚举：'good' 'temp-bad' 'final-bad', default:'good'}
	BestParentUnit string `json:"best_parent_unit"` // 最优父单元

	// 与杨杰沟通，这两个字段表示前驱和后继，但是从DAG网络和数据库update两方面考虑，暂时不需要这两个字段
	// In communication with Yang Jie, these two fields represent the precursor and successor, but considering the DAG network and the database update, the two fields are not needed temporarily.
	// ToUnit                map[string]bool `json:"to_unit"`                  // parents
	// FromUnit              map[string]bool `json:"from_unit"`                // childs

	// 与杨杰沟通，当时在未确定数据库的时候考虑外键、模糊查询的情况
	// Communicating with Yang Jie, then ha has considered foreign keys and fuzzy queries at the time when the database was not determined
	// Key                   string          `json:"key"`                      // index: key
}

func NewUnit(header *Header, msgs []*Message, receipt []*Receipt) *Unit {
	u := &Unit{
		Head: CopyHeader(header),
	}
	u.Version = "1.0"
	u.Alt = "1"
	u.CreationDate = time.Now()
	u.Sequence = "good"
	return u
}

// return  unit'hash
func (u *Unit) Hash() common.Hash {
	v := rlpHash(u)
	return v
}

//
func (u *Unit) Header() *Header { return CopyHeader(u.Head) }

//
func CurrentUnit() *Unit {
	return &Unit{CreationDate: time.Now()}
}

//
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
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

//
func (u *Unit) NumberU64() uint64 { return u.Head.Number.Uint64() }

// transactions
func (u *Unit) Transactions() []*Transaction {
	ts := make([]*Transaction, 0)
	return ts
}

//
func (u *Unit) ParentHash() common.Hash {
	return u.Head.ParentHash
}
func (u *Unit) Is_stable() bool {
	// 判断候选主链的交点是否是这个unit
	return true
}
func (u *Unit) GetBestParentUnit() string {
	return ""
}
func (u *Unit) GetLastBall() string {
	ball := getlastball(u.Unit)

	return ball.Ball
}
func getlastball(unit string) *Ball {
	return &Ball{Ball: ""}
}

// type Message struct {
// 	App             string      `json:"app"`              // 消息类型
// 	PayloadLocation string      `json:"payload_location"` // 负载位置
// 	PayloadHash     string      `json:"payload_hash"`     // payload hash
// 	Payload         interface{} `json:"payload"`          // should be the choice of Payload, ContractTplPayload, ContractDeployPayload, ContractInvokePayload
// 	CreationDate    time.Time   `json:"creation_date"`
// }

>>>>>>> 284ec7719faa69448b92037bb670e9f5723b513d
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
<<<<<<< HEAD
	TemplateId	string 					`json:"template_id"`	// configure xml file of contract
	Bytecode	[]byte 					`json:"bytecode"`		// contract bytecode
	ReadSet		map[string]interface{}  `json:"read_set"`		// the set data of read, and value could be any type
	WriteSet	map[string]interface{}	`json:"write_set"`		// the set data of write, and value could be any type
=======
	TemplateId string `json:"template_id"` // configure xml file of contract
	Bytecode   []byte `json:"bytecode"`    // contract bytecode
>>>>>>> 284ec7719faa69448b92037bb670e9f5723b513d
}

// Contract instance message
// App: contract_deploy
type ContractDeployPayload struct {
<<<<<<< HEAD
	TemplateId	string 				 	`json:"template_id"`	// contract template id
	Config		[]byte 	             	`json:"config"`		// configure xml file of contract instance parameters
	ReadSet		map[string]interface{}  `json:"read_set"`		// the set data of read, and value could be any type
	WriteSet	map[string]interface{}	`json:"write_set"`		// the set data of write, and value could be any type
=======
	TemplateId string `json:"template_id"` // contract template id
	Config     []byte `json:"config"`      // configure xml file of contract instance parameters
>>>>>>> 284ec7719faa69448b92037bb670e9f5723b513d
}

// Contract invoke message
// App: contract_invoke
type ContractInvokePayload struct {
<<<<<<< HEAD
	ContractId	string               	`json:"contract_id"`	// contract id
	Function	[]byte               	`json:"function"`		// serialized value of invoked function with call parameters
	ReadSet		map[string]interface{}  `json:"read_set"`		// the set data of read, and value could be any type
	WriteSet	map[string]interface{}	`json:"write_set"`		// the set data of write, and value could be any type
}

// Token exchange message and verify message
// App: config	// update global config
type ConfigPayload struct {
	ConfigSet 	map[string]interface{}	`json:"config_set"`		// the array of global config
}

// Token exchange message and verify message
// App: text
type TextPayload struct {
	Text	[]byte		`json:"text"`	// Textdata
}
/************************** End of Payload Details ******************************************/

type Author struct {
	Address       	Hash        	`json:"address"`
	Pubkey		  	Hash 			`json:"pubkey"`
	TxAuthentifier 	Authentifier  	`json:"authentifiers"`
}

type Authentifier struct {
	R string `json:"r"`
}

//func (a *Authentifier) ToDB() ([]byte, error) {
//	return json.Marshal(a)
//}
//
//func (a *Authentifier) FromDB(info []byte) error {
//	return json.Unmarshal(info, a)
//}
//
//func NewUnit() *Unit {
//	return &Unit{}
//}
//
//func (u *Unit) Is_stable() bool {
//	// 判断候选主链的交点是否是这个unit
//	return true
//}
//
//func (u *Unit) GetBestParentUnit() string {
//	return ""
//}
//
//func (u *Unit) GetLastBall() string {
//	ball := getlastball(u.Unit)
//	return ball.Ball
//}
//
//func getlastball(unit string) *Ball {
//	return &Ball{Ball: ""}
//}
//
//type UnStableUnitsList []*Unit
//
//func (ulist UnStableUnitsList) Len() int { return len(ulist) }
//
//func (ulist UnStableUnitsList) Less(i, j int) bool {
//	if ulist[i].Unit < ulist[j].Unit {
//		return true
//	}
//	return false
//}
//
//func (ulist UnStableUnitsList) Swap(i, j int) {
//	var temp *Unit = ulist[i]
//	ulist[i] = ulist[j]
//	ulist[j] = temp
//}
//
//func (ulist UnStableUnitsList) GetMaxMainChainIndex() (max int64) {
//	//max = ulist[0].MainChainIndex
//	for _, v := range ulist {
//		if max < v.MainChainIndex && !v.IsStable {
//			max = v.MainChainIndex
//		}
//	}
//	return
//}
//
//type StableUnitslist []*Unit
//
//func (list StableUnitslist) Len() int { return len(list) }
//
//func (list StableUnitslist) Less(i, j int) bool {
//	if list[i].Unit < list[j].Unit {
//		return true
//	}
//	return false
//}
//
//func (list StableUnitslist) Swap(i, j int) {
//	var temp *Unit = list[i]
//	list[i] = list[j]
//	list[j] = temp
//}
//
//func (list StableUnitslist) GetMaxMainChainIndex() (max int64) {
//	//max = list[0].MainChainIndex
//	for _, v := range list {
//		if max < v.MainChainIndex && v.IsStable {
//			max = v.MainChainIndex
//		}
//	}
//	return
//}
//
//// return  unit'hash
//func (u *Unit) Hash() common.Hash {
//	v := common.RlpHash(u)
//	return v
//}
=======
	ContractId string `json:"contract_id"` // contract id
	Function   []byte `json:"function"`    // serialized value of invoked function with call parameters
}

// type Input struct {
// 	Unit         string `json:"unit"`          // 未花费unit
// 	MessageIndex int    `json:"message_index"` // 消息序号
// 	OutputIndex  int    `json:"output_index"`  // 输出序号
// }
// type Output struct {
// 	Address string `json:"address"` // 输出地址
// 	Amount  int    `json:"amount"`  // 输出金额

// }
type Author struct {
	Address       string       `json:"address"`
	Pubkey        string       `json:"pubkey"`
	Authentifiers Authentifier `json:"authentifiers"`
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

type UnStableUnitsList []*Unit

func (ulist UnStableUnitsList) Len() int { return len(ulist) }
func (ulist UnStableUnitsList) Less(i, j int) bool {
	if ulist[i].Unit < ulist[j].Unit {
		return true
	}
	return false
}
func (ulist UnStableUnitsList) Swap(i, j int) {
	var temp *Unit = ulist[i]
	ulist[i] = ulist[j]
	ulist[j] = temp
}

type StableUnitslist []*Unit

func (list StableUnitslist) Len() int { return len(list) }
func (list StableUnitslist) Less(i, j int) bool {
	if list[i].Unit < list[j].Unit {
		return true
	}
	return false
}
func (list StableUnitslist) Swap(i, j int) {
	var temp *Unit = list[i]
	list[i] = list[j]
	list[j] = temp
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}
>>>>>>> 284ec7719faa69448b92037bb670e9f5723b513d
