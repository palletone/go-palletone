/*
 *
 * 	This file is part of go-palletone.
 * 	go-palletone is free software: you can redistribute it and/or modify
 * 	it under the terms of the GNU General Public License as published by
 * 	the Free Software Foundation, either version 3 of the License, or
 * 	(at your option) any later version.
 * 	go-palletone is distributed in the hope that it will be useful,
 * 	but WITHOUT ANY WARRANTY; without even the implied warranty of
 * 	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * 	GNU General Public License for more details.
 * 	You should have received a copy of the GNU General Public License
 * 	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018-2020
 *
 */

package packetcc

import (
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

type Packet struct {
	PubKey          []byte         //红包对应的公钥，也是红包的唯一标识
	Creator         common.Address //红包发放人员地址
	Token           *modules.Asset //红包中的TokenID
	Amount          uint64         //红包总金额
	Count           uint32         //红包数，为0表示可以无限领取
	MinPacketAmount uint64         //单个红包最小额
	MaxPacketAmount uint64         //单个红包最大额,最大额最小额相同，则说明不是随机红包,0则表示完全随机
	ExpiredTime     uint64         //红包过期时间，0表示永不过期
	Remark          string         //红包的备注
}

func (p *Packet) IsFixAmount() bool {
	return p.MinPacketAmount == p.MaxPacketAmount && p.MaxPacketAmount > 0
}
func (p *Packet) GetPullAmount(seed int64, amount uint64, count uint32) uint64 {
	if p.IsFixAmount() {
		return p.MaxPacketAmount
	}
	if count == 1 {
		return amount
	}
	expect := amount / uint64(count)
	return NormalRandom(seed, expect, p.MinPacketAmount, p.MaxPacketAmount)
}
func NormalRandom(seed int64, expect uint64, min, max uint64) uint64 {
	if expect < min {
		return min
	}
	if expect > max {
		return max
	}
	//计算标准差
	bzc1 := max - expect
	bzc2 := expect - min
	bzc := bzc1
	if bzc2 < bzc1 {
		bzc = bzc2
	}
	bzc = bzc / 3 //正态分布，3标准差内的概率>99%

	for i := int64(0); i < 100; i++ {
		rand.Seed(seed + i)
		number := rand.NormFloat64()*float64(bzc) + float64(expect)
		if number <= 0 {
			continue
		}
		n64 := uint64(number)
		if n64 >= min && n64 <= max {
			return n64
		}
	}
	return expect
}

func savePacket(stub shim.ChaincodeStubInterface, p *Packet) error {
	key := PacketPrefix + hex.EncodeToString(p.PubKey)
	value, err := rlp.EncodeToBytes(p)
	if err != nil {
		return err
	}
	return stub.PutState(key, value)
}
func getPacket(stub shim.ChaincodeStubInterface, pubKey []byte) (*Packet, error) {
	key := PacketPrefix + hex.EncodeToString(pubKey)
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	p := Packet{}
	err = rlp.DecodeBytes(value, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

//红包余额
type PacketBalance struct {
	Amount uint64
	Count  uint32
}

func savePacketBalance(stub shim.ChaincodeStubInterface, pubKey []byte, balanceAmt uint64, balanceCount uint32) error {
	key := PacketBalancePrefix + hex.EncodeToString(pubKey)
	value, err := rlp.EncodeToBytes(PacketBalance{Amount: balanceAmt, Count: balanceCount})
	if err != nil {
		return err
	}
	return stub.PutState(key, value)
}
func getPacketBalance(stub shim.ChaincodeStubInterface, pubKey []byte) (uint64, uint32, error) {
	key := PacketBalancePrefix + hex.EncodeToString(pubKey)
	value, err := stub.GetState(key)
	b := PacketBalance{}
	err = rlp.DecodeBytes(value, &b)
	if err != nil {
		return 0, 0, err
	}
	return b.Amount, b.Count, nil
}

type PacketJson struct {
	PubKey          string          //红包对应的公钥，也是红包的唯一标识
	Creator         common.Address  //红包发放人员地址
	Token           string          //红包中的TokenID
	TotalAmount     decimal.Decimal //红包总金额
	PacketCount     uint32          //红包数，为0表示可以无限领取
	MinPacketAmount decimal.Decimal //单个红包最小额
	MaxPacketAmount decimal.Decimal //单个红包最大额,最大额最小额相同，则说明不是随机红包
	ExpiredTime     string          //红包过期时间，0表示永不过期
	Remark          string          //红包的备注
	BalanceAmount   decimal.Decimal //红包剩余额度
	BalanceCount    uint32          //红包剩余次数
}

func convertPacket2Json(packet *Packet, balanceAmount uint64, balanceCount uint32) *PacketJson {
	js := &PacketJson{
		PubKey:          hex.EncodeToString(packet.PubKey),
		Creator:         packet.Creator,
		Token:           packet.Token.String(),
		TotalAmount:     packet.Token.DisplayAmount(packet.Amount),
		MinPacketAmount: packet.Token.DisplayAmount(packet.MinPacketAmount),
		MaxPacketAmount: packet.Token.DisplayAmount(packet.MaxPacketAmount),
		PacketCount:     packet.Count,
		Remark:          packet.Remark,
		BalanceAmount:   packet.Token.DisplayAmount(balanceAmount),
		BalanceCount:    balanceCount,
	}
	if packet.ExpiredTime != 0 {
		js.ExpiredTime = time.Unix(int64(packet.ExpiredTime), 0).String()
	}
	return js
}
