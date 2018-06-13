// unit package, unit structure and storage api
package modules

import (
	"encoding/json"
	"time"

	"github.com/palletone/go-palletone/dag/common"
)

// key: unit.hash(unit)
type Unit struct {
	Unit                  string          `json:"unit"`                     // unit hash
	Version               string          `json:"version"`                  // 版本号
	Alt                   string          `json:"alt"`                      // 资产号
	Messages              []Message       `json:"messages"`                 // 消息
	Authors               []Author        `json:"authors"`                  // 发起人
	ParentUnits           []string        `json:"parent_units"`             // 父单元数组
	CreationDate          time.Time       `json:"creation_date"`            // 创建时间
	LastBall              string          `json:"last_ball"`                // 最后一个球
	LastBallUnit          string          `json:"last_ball_unit"`           // 最后一个球的unit
	WitnessListUnit       string          `json:"witness_list_unit"`        // 稳定点hash
	ContentHash           string          `json:"content_hash"`             // 内容hash
	HeadersCommission     int             `json:"headers_commission"`       // 头佣金
	PayloadCommission     int             `json:"payload_commission"`       // 酬劳
	IsFree                bool            `json:"is_free"`                  // 自由单元
	IsOnMainChain         bool            `json:"is_on_main_chain"`         // 是在主链上
	MainChainIndex        int64           `json:"main_chain_index"`         // 主链序号
	LatestIncludedMcIndex int64           `json:"latest_included_mc_index"` // 最新的主链序号
	Level                 int             `json:"level"`                    // 单元级别
	WitnessedLevel        int             `json:"witness_level"`            // 见证级别
	IsStable              bool            `json:"is_stable"`                // 是否稳定
	Sequence              string          `json:"sequence"`                 // {枚举：'good' 'temp-bad' 'final-bad', default:'good'}
	BestParentUnit        string          `json:"best_parent_unit"`         // 最优父单元
	ToUnit                map[string]bool `json:"to_unit"`                  // parents
	FromUnit              map[string]bool `json:"from_unit"`                // childs
	Key                   string          `json:"key"`                      // index: key
}

// key: message.hash(message+timestamp)
type Message struct {
	App             string    `json:"app"`              // 消息类型
	PayloadLocation string    `json:"payload_location"` // 负载位置
	PayloadHash     string    `json:"payload_hash"`     // payload hash
	Payload         Payload   `json:"payload"`
	CreationDate    time.Time `json:"creation_date"`
}

type Payload struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
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
func NewUnit() *Unit {
	return &Unit{Version: "1.0", Alt: "1", CreationDate: time.Now(), IsFree: true, Sequence: "good"}
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
func (ulist UnStableUnitsList) GetMaxMainChainIndex() (max int64) {
	//max = ulist[0].MainChainIndex
	for _, v := range ulist {
		if max < v.MainChainIndex && !v.IsStable {
			max = v.MainChainIndex
		}
	}
	return
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
func (list StableUnitslist) GetMaxMainChainIndex() (max int64) {
	//max = list[0].MainChainIndex
	for _, v := range list {
		if max < v.MainChainIndex && v.IsStable {
			max = v.MainChainIndex
		}
	}
	return
}

// return  unit'hash
func (u *Unit) Hash() common.Hash {
	v := common.RlpHash(u)
	return v
}
