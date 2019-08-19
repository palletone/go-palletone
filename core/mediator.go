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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package core

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
)

var Suite = bn256.NewSuiteG2()

func GenInitPair() (kyber.Scalar, kyber.Point) {
	sc := Suite.Scalar().Pick(Suite.RandomStream())

	return sc, Suite.Point().Mul(sc, nil)
}

// mediator 结构体 和具体的账户模型有关
type Mediator struct {
	Address    common.Address `json:"address"`    // mediator账户地址，主要用于产块签名
	RewardAdd  common.Address `json:"rewardAdd"`  // mediator奖励地址，主要用于接收产块奖励
	InitPubKey kyber.Point    `json:"initPubKey"` // mediator的群签名初始公钥
	Node       *discover.Node `json:"node"`       // mediator节点网络信息，包括ip和端口等
	*MediatorApplyInfo
	*MediatorInfoExpand
}

// mediator扩展信息
type MediatorInfoExpand struct {
	TotalMissed          uint64 `json:"totalMissed"`          // 当前mediator未能按照调度生产区块的总个数
	LastConfirmedUnitNum uint32 `json:"lastConfirmedUnitNum"` // 当前mediator最新生产的区块编号
	TotalVotes           uint64 `json:"totalVotes"`           // 当前mediator的总共得票数量
}

func NewMediatorInfoExpand() *MediatorInfoExpand {
	return &MediatorInfoExpand{
		TotalMissed:          0,
		LastConfirmedUnitNum: 0,
		TotalVotes:           0,
	}
}

func NewMediator() *Mediator {
	return &Mediator{
		Address:            common.Address{},
		RewardAdd:          common.Address{},
		InitPubKey:         nil,
		Node:               nil,
		MediatorApplyInfo:  NewMediatorApplyInfo(),
		MediatorInfoExpand: NewMediatorInfoExpand(),
	}
}

func (med *Mediator) GetRewardAdd() common.Address {
	if med.RewardAdd != (common.Address{}) {
		return med.RewardAdd
	}

	return med.Address
}

// Mediator申请信息
type MediatorApplyInfo struct {
	Logo        string `json:"logo"`      // 节点图标url
	Name        string `json:"name"`      // 节点名称
	Location    string `json:"loc"`       // 节点所在地区
	Url         string `json:"url"`       // 节点宣传网站
	Description string `json:"applyInfo"` // 节点详细信息描述
}

func NewMediatorApplyInfo() *MediatorApplyInfo {
	return &MediatorApplyInfo{
		Logo:        "",
		Name:        "",
		Location:    "",
		Url:         "",
		Description: "",
	}
}

func StrToMedNode(medNode string) (*discover.Node, error) {
	node, err := discover.ParseNode(medNode)
	if err != nil {
		err = fmt.Errorf("invalid mediator node \"%v\" : %v", medNode, err)
		return nil, err
	}

	return node, nil
}

func StrToMedAdd(addStr string) (common.Address, error) {
	address := strings.TrimSpace(addStr)
	address = strings.Trim(address, "\"")
	if address == "" {
		err := fmt.Errorf("mediator account address is empty string")
		return common.Address{}, err
	}

	addr, err := common.StringToAddress(address)
	if err != nil || addr.GetType() != common.PublicKeyHash {
		err = fmt.Errorf("invalid mediator account address \"%v\" : %v", address, err)
		return addr, err
	}

	return addr, nil
}

func StrToScalar(secStr string) (kyber.Scalar, error) {
	secB := base58.Decode(secStr)
	sec := Suite.Scalar()

	err := sec.UnmarshalBinary(secB)
	if err != nil {
		err = fmt.Errorf("invalid init mediator private key \"%v\" : %v", secStr, err)
		return nil, err
	}

	return sec, nil
}

func StrToPoint(pubStr string) (kyber.Point, error) {
	pubB := base58.Decode(pubStr)
	pub := Suite.Point()

	err := pub.UnmarshalBinary(pubB)
	if err != nil {
		err = fmt.Errorf("invalid init mediator public key \"%v\" : %v", pubStr, err)
		return nil, err
	}

	return pub, nil
}
