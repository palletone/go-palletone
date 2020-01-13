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

package tokenallocation

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
)

type Packet struct {
	PubKey          []byte         //红包对应的公钥，也是红包的唯一标识
	Creator         common.Address //红包发放人员地址
	Token           *modules.Asset //红包中的TokenID
	Amount          uint64         //红包总金额
	Count           uint32         //红包数，为0表示可以无限领取
	MinPacketAmount uint64         //单个红包最小额
	MaxPacketAmount uint64         //单个红包最大额,最大额最小额相同，则说明不是随机红包
	ExpiredTime     uint64         //红包过期时间，0表示永不过期
	Remark          string         //红包的备注
}

func (p *Packet) IsFixAmount() bool {
	return p.MinPacketAmount == p.MaxPacketAmount
}
func (p *Packet) GetPullAmount(seed []byte) uint64 {
	return p.MaxPacketAmount
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
