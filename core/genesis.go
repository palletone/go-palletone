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
	"strconv"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common/log"
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	//年利率
	DepositRate string `json:"depositRate"`
	//基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等
	FoundationAddress string `json:"foundationAddress"`
	//保证金的数量
	DepositAmountForMediator  string `json:"depositAmountForMediator"`
	DepositAmountForJury      string `json:"depositAmountForJury"`
	DepositAmountForDeveloper string `json:"depositAmountForDeveloper"`
	//保证金周期
	DepositPeriod string `json:"depositPeriod"`

	// ROOT CA的持有者
	RootCAHolder string `json:"rootCAHolder"`
	// ROOT CA证书内容
	RootCABytes string `json:"rootCABytes"`
}

type Genesis struct {
	Version string `json:"version"`
	Alias   string `json:"alias"`
	//TokenAmount  uint64       `json:"tokenAmount"`
	TokenAmount  string       `json:"tokenAmount"`
	TokenDecimal uint32       `json:"tokenDecimal"`
	DecimalUnit  string       `json:"decimal_unit"`
	ChainID      uint64       `json:"chainId"`
	TokenHolder  string       `json:"tokenHolder"`
	Text         string       `json:"text"`
	RootCA       string       `json:"rootCA"`
	SystemConfig SystemConfig `json:"systemConfig"`

	InitialParameters         ChainParameters          `json:"initialParameters"`
	ImmutableParameters       ImmutableChainParameters `json:"immutableChainParameters"`
	InitialTimestamp          int64                    `json:"initialTimestamp"`
	InitialActiveMediators    uint16                   `json:"initialActiveMediators"`
	InitialMediatorCandidates []*InitialMediator       `json:"initialMediatorCandidates"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	amount, err := strconv.ParseInt(g.TokenAmount, 10, 64)
	if err != nil {
		log.Error("genesis", "get token amount err:", err)
		return uint64(0)
	}
	return uint64(amount)
}

type MediatorInfoBase struct {
	AddStr     string `json:"account"`
	InitPubKey string `json:"initPubKey"`
	Node       string `json:"node"`
}

func NewMediatorInfoBase() *MediatorInfoBase {
	return &MediatorInfoBase{
		AddStr:     "",
		InitPubKey: "",
		Node:       "",
	}
}

type InitialMediator struct {
	*MediatorInfoBase
}

func NewInitialMediator() *InitialMediator {
	return &InitialMediator{
		MediatorInfoBase: NewMediatorInfoBase(),
	}
}

// author Albert·Gou
func ScalarToStr(sec kyber.Scalar) string {
	secB, err := sec.MarshalBinary()
	if err != nil {
		log.Error(err.Error())
	}

	return base58.Encode(secB)
}

// author Albert·Gou
func PointToStr(pub kyber.Point) string {
	pubB, err := pub.MarshalBinary()
	if err != nil {
		log.Error(err.Error())
	}

	return base58.Encode(pubB)
}

// author Albert·Gou
func CreateInitDKS() (secStr, pubStr string) {
	sec, pub := GenInitPair()

	secStr = ScalarToStr(sec)
	pubStr = PointToStr(pub)

	return
}

// this is for root ca
type RootCA struct {
	// 组织/公司
	Organization string
	// 部门/单位
	Department string
	// 城市
	Location string
	// 省份
	State string
	// 国家
	Country string
	// 加密算法
	Encption string
	// 哈希签名算法ry
	SignatureAlgorithm string
	// 加密位数
	EncryptionBits int8
	// 邮箱
	Email string
	// 域名
	Domain string
}
