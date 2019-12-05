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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"go.dedis.ch/kyber/v3"
)

type DigitalIdentityConfig struct {
	RootCAHolder string `json:"rootCAHolder"` // ROOT CA的持有者
	RootCABytes  string `json:"rootCABytes"`  // ROOT CA证书内容
}

func DefaultDigitalIdentityConfig() DigitalIdentityConfig {
	return DigitalIdentityConfig{
		// default root ca holder, 默认是基金会地址
		RootCAHolder: DefaultFoundationAddress,
		RootCABytes:  DefaultRootCABytes,
	}
}

type Genesis struct {
	Version      string `json:"version"`
	GasToken     string `json:"gasToken"`
	TokenAmount  string `json:"tokenAmount"`
	PublicKey    string `json:"mediator_pubkey"`
	MediatorSign string `json:"mediator_sign"`
	//TokenAmount uint64 `json:"tokenAmount"`
	//TokenDecimal              uint32                   `json:"tokenDecimal"`
	//DecimalUnit               string                   `json:"decimal_unit"`
	ChainID     uint64 `json:"chainId"`
	TokenHolder string `json:"tokenHolder"`
	Text        string `json:"text"`
	//SystemConfig          SystemConfig             `json:"systemConfig"`
	DigitalIdentityConfig DigitalIdentityConfig    `json:"digitalIdentityConfig"`
	ParentUnitHash        common.Hash              `json:"parentUnitHash"`
	ParentUnitHeight      int64                    `json:"parentUnitHeight"`
	InitialParameters     ChainParameters          `json:"initialParameters"`
	ImmutableParameters   ImmutableChainParameters `json:"immutableChainParameters"`
	InitialTimestamp      int64                    `json:"initialTimestamp"`
	//InitialActiveMediators    uint16                   `json:"initialActiveMediators"`
	InitialMediatorCandidates []*InitialMediator `json:"initialMediatorCandidates"`
	SystemContracts           []SysContract      `json:"systemContracts"`
}

type SysContract struct {
	Address common.Address `json:"address"`
	Name    string         `json:"name"`
	Active  bool           `json:"active"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	amount, err := strconv.ParseInt(g.TokenAmount, 10, 64)
	if err != nil {
		log.Error("genesis", "get token amount err:", err)
		return uint64(0)
	}
	return uint64(amount)
}

// mediator基本信息
type MediatorInfoBase struct {
	AddStr     string `json:"account"`    // mediator账户地址，主要用于产块签名
	RewardAdd  string `json:"rewardAdd"`  // mediator奖励地址，主要用于接收产块奖励
	InitPubKey string `json:"initPubKey"` // mediator的群签名初始公钥
	Node       string `json:"node"`       // mediator节点网络信息，包括ip和端口等
}

func NewMediatorInfoBase() *MediatorInfoBase {
	return &MediatorInfoBase{
		AddStr:     "",
		RewardAdd:  "",
		InitPubKey: "",
		Node:       "",
	}
}

func (mib *MediatorInfoBase) Validate() (common.Address, error) {
	addr, err := StrToMedAdd(mib.AddStr)
	if err != nil {
		return addr, err
	}

	if mib.RewardAdd != "" {
		_, err := StrToMedAdd(mib.RewardAdd)
		if err != nil {
			return addr, err
		}
	}

	_, err = StrToPoint(mib.InitPubKey)
	if err != nil {
		return addr, err
	}

	node, err := StrToMedNode(mib.Node)
	if err != nil {
		return addr, err
	}

	err = node.ValidateComplete()
	if err != nil {
		return addr, err
	}

	return addr, nil
}

// genesis 文件定义的mediator结构体
type InitialMediator struct {
	*MediatorInfoBase
	JurorDepositExtraJson
}

func (im *InitialMediator) Validate() (common.Address, JurorDepositExtra, error) {
	var jde JurorDepositExtra

	addr, err := im.MediatorInfoBase.Validate()
	if err != nil {
		return addr, jde, err
	}

	jde, err = im.JurorDepositExtraJson.Validate(im.AddStr)
	if err != nil {
		return addr, jde, err
	}

	return addr, jde, nil
}

func NewInitialMediator() *InitialMediator {
	return &InitialMediator{
		MediatorInfoBase: NewMediatorInfoBase(),
		JurorDepositExtraJson: NewJurorDepositExtraJson(),
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
