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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	TxCoinYearRate     string `json:"txCoinYearRate"`     //交易币天的年利率
	GenerateUnitReward string `json:"generateUnitReward"` //每生产一个单元，奖励多少Dao的PTN
	FoundationAddress  string `json:"foundationAddress"`  //基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等

	DepositRate               string `json:"depositRate"`              //保证金的年利率
	DepositAmountForMediator  string `json:"depositAmountForMediator"` //保证金的数量
	DepositAmountForJury      string `json:"depositAmountForJury"`
	DepositAmountForDeveloper string `json:"depositAmountForDeveloper"`
	DepositPeriod             string `json:"depositPeriod"` //保证金周期

	//对启动用户合约容器的相关资源的限制
	UccMemory     string `json:"ucc_memory"`       //物理内存  104857600  100m
	UccMemorySwap string `json:"ucc_memory_swap"`  //内存交换区，不设置默认为memory的两倍
	UccCpuShares  string `json:"ucc_cpu_shares"`   //CPU占用率，相对的  CPU 利用率权重，默认为 1024
	UccCpuQuota   string `json:"ucc_cpu_quota"`    // 限制CPU --cpu-period=50000 --cpu-quota=25000
	UccCpuPeriod  string `json:"ucc_cpu_period"`   //限制CPU 周期设为 50000，将容器在每个周期内的 CPU 配额设置为 25000，表示该容器每 50ms 可以得到 50% 的 CPU 运行时间
	UccCpuSetCpus string `json:"ucc_cpu_set_cpus"` //限制使用某些CPUS  "1,3"  "0-2"

	//对中间容器的相关资源限制
	TempUccMemory     string `json"temp_ucc_memory"`
	TempUccMemorySwap string `json"temp_ucc_memory_swap"`
	TempUccCpuShares  string `json"temp_ucc_cpu_shares"`
	TempUccCpuQuota   string `json"temp_ucc_cpu_quota"`

	ActiveMediatorCount string `json:"activeMediatorCount"`
}

type DigitalIdentityConfig struct {
	RootCAHolder string `json:"rootCAHolder"` // ROOT CA的持有者
	RootCABytes  string `json:"rootCABytes"`  // ROOT CA证书内容
}

type Genesis struct {
	Version  string `json:"version"`
	GasToken string `json:"gasToken"`
	//TokenAmount  uint64       `json:"tokenAmount"`
	TokenAmount string `json:"tokenAmount"`
	//TokenDecimal              uint32                   `json:"tokenDecimal"`
	//DecimalUnit               string                   `json:"decimal_unit"`
	ChainID                   uint64                   `json:"chainId"`
	TokenHolder               string                   `json:"tokenHolder"`
	Text                      string                   `json:"text"`
	SystemConfig              SystemConfig             `json:"systemConfig"`
	DigitalIdentityConfig     DigitalIdentityConfig    `json:"digitalIdentityConfig"`
	ParentUnitHash            common.Hash              `json:"parentUnitHash"`
	ParentUnitHeight          int64                    `json:"parentUnitHeight"`
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
