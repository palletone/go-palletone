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

	"github.com/btcsuite/btcutil/base58"
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common/log"
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	DepositRate float64 `json:"depositRate"`
	//基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等
	FoundationAddress string `json:"foundationAddress"`
	//保证金的数量
	DepositAmount uint64 `json:"depositAmount"`
	//这是保证金的地址
	DepositContractAddress string `json:"depositContractAddress"`
}

type Genesis struct {
	Version      string       `json:"version"`
	Alias        string       `json:"alias"`
	TokenAmount  uint64       `json:"tokenAmount"`
	TokenDecimal uint32       `json:"tokenDecimal"`
	DecimalUnit  string       `json:"decimal_unit"`
	ChainID      uint64       `json:"chainId"`
	TokenHolder  string       `json:"tokenHolder"`
	Text         string       `json:"text"`
	SystemConfig SystemConfig `json:"systemConfig"`

	InitialParameters         ChainParameters          `json:"initialParameters"`
	ImmutableParameters       ImmutableChainParameters `json:"immutableChainParameters"`
	InitialTimestamp          int64                    `json:"initialTimestamp"`
	InitialActiveMediators    uint16                   `json:"initialActiveMediators"`
	InitialMediatorCandidates []*MediatorInfo          `json:"initialMediatorCandidates"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	return g.TokenAmount
}

type MediatorInfo struct {
	Address     string
	InitPartPub string
	Node        string
	//WebsiteUrl  string
}

// author Albert·Gou
func ScalarToStr(sec kyber.Scalar) string {
	secB, err := sec.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return base58.Encode(secB)
}

// author Albert·Gou
func PointToStr(pub kyber.Point) string {
	pubB, err := pub.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return base58.Encode(pubB)
}

func (medInfo *MediatorInfo) InfoToMediator() Mediator {
	// 1. 解析 mediator 账户地址
	add := StrToMedAdd(medInfo.Address)

	// 2. 解析 mediator 的 DKS 初始公钥
	pub := StrToPoint(medInfo.InitPartPub)

	// 3. 解析mediator 的 node 节点信息
	node := StrToMedNode(medInfo.Node)

	md := Mediator{
		Address:     add,
		InitPartPub: pub,
		Node:        node,
	}

	return md
}
