// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light Ethereum Subprotocol.
package light

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
)

/*
type PartitionChain struct {
	GenesisHash    common.Hash
	GenesisHeight  uint64
	ForkUnitHash   common.Hash
	ForkUnitHeight uint64
	GasToken       AssetId
	Status         byte     //Active:1 ,Terminated:0,Suspended:2
	SyncModel      byte     //Push:1 , Pull:2, Push+Pull:3
	Peers          []string // IP:port format string
}

//作为一个分区，我会维护我链接到的主链
type MainChain struct {
	GenesisHash common.Hash
	Status      byte //Active:1 ,Terminated:0,Suspended:2
	SyncModel   byte //Push:1 , Pull:2, Push+Pull:0
	GasToken    AssetId
	NetworkId   uint64
	Version     int
	Peers       []string // IP:port format string
}
*/
// Handshake executes the les protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *peer) CorsHandshake(number *modules.ChainIndex, genesis common.Hash, server *LesServer, headhash common.Hash,
	mainchain *modules.MainChain) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	var send keyValueList
	send = send.add("protocolVersion", uint64(mainchain.Version))
	send = send.add("networkId", mainchain.NetworkId)
	send = send.add("headNum", *number)
	send = send.add("headHash", headhash)
	send = send.add("genesisHash", genesis)

	if server != nil {
		//send = send.add("flowControl/BL", server.defParams.BufLimit)
		//send = send.add("flowControl/MRR", server.defParams.MinRecharge)

		send = send.add("fullnode", nil)
	} else {
		p.requestAnnounceType = announceTypeSimple // set to default until "very light" client mode is implemented
		send = send.add("announceType", p.requestAnnounceType)
	}
	recvList, err := p.sendReceiveHandshake(send)
	if err != nil {
		return err
	}
	recv := recvList.decode()

	var rGenesis, rHash common.Hash
	var rVersion, rNetwork uint64
	//var rTd *big.Int
	var rNum modules.ChainIndex

	if err := recv.get("protocolVersion", &rVersion); err != nil {
		return err
	}
	if err := recv.get("networkId", &rNetwork); err != nil {
		return err
	}
	if err := recv.get("headHash", &rHash); err != nil {
		return err
	}
	if err := recv.get("headNum", &rNum); err != nil {
		return err
	}
	if err := recv.get("genesisHash", &rGenesis); err != nil {
		return err
	}

	if rGenesis != genesis {
		return errResp(ErrGenesisBlockMismatch, "%x (!= %x)", rGenesis[:8], genesis[:8])
	}
	if rNetwork != p.network {
		return errResp(ErrNetworkIdMismatch, "%d (!= %d)", rNetwork, p.network)
	}
	if int(rVersion) != p.version {
		return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", rVersion, p.version)
	}
	if server != nil {
		if recv.get("announceType", &p.announceType) != nil {
			p.announceType = announceTypeSimple
		}
	}

	if err := recv.get("fullnode", nil); err != nil {
		p.fullnode = false
		log.Debug("Light Palletone peer->Handshake peer is light node")
	} else {
		p.fullnode = true
		log.Debug("Light Palletone peer->Handshake peer is full node")
	}
	//
	p.headInfo = &announceData{Hash: rHash, Number: rNum}
	return nil
}

//func AAA()  {
//	contract.ccquery("PCGTta3M4t3yXu8uRgkKvaWd2d8DRxVdGDZ",["listPartition"])
//}
