/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package modules

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
)

//作为主链，我会维护我上面支持的分区
type PartitionChain struct {
	GenesisHeaderRlp []byte
	//GenesisHeight   uint64
	ForkUnitHash     common.Hash
	ForkUnitHeight   uint64
	GasToken         AssetId
	Status           byte //Active:1 ,Terminated:0,Suspended:2
	SyncModel        byte //Push:1 , Pull:2, Push+Pull:3
	NetworkId        uint64
	Version          uint64
	StableThreshold  uint32    //需要多少个签名才能是稳定单元
	Peers            []string  //pnode://publickey@IP:port format string
	CrossChainTokens []AssetId // 哪些Token可以跨链转移
}

func (p *PartitionChain) GetGenesisHeader() *Header {
	header := &Header{}
	err := rlp.DecodeBytes(p.GenesisHeaderRlp, header)
	if err != nil {
		return nil
	}
	return header
}

//作为一个分区，我会维护我链接到的主链
type MainChain struct {
	GenesisHeaderRlp []byte
	Status           byte //Active:1 ,Terminated:0,Suspended:2
	SyncModel        byte //Push:1 , Pull:2, Push+Pull:0
	GasToken         AssetId
	NetworkId        uint64
	Version          uint64
	StableThreshold  uint32    //需要多少个签名才能是稳定单元
	Peers            []string  // pnode://publickey@IP:port format string
	CrossChainTokens []AssetId // 哪些Token可以跨链转移
}

func (p *MainChain) GetGenesisHeader() *Header {
	header := &Header{}
	err := rlp.DecodeBytes(p.GenesisHeaderRlp, header)
	if err != nil {
		return nil
	}
	return header
}
