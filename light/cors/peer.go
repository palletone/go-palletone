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
 * @author PalletOne core developer Jiyou Wang <dev@pallet.one>
 * @date 2018
 */
package cors

import (
	"crypto/ecdsa"
	//"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn"
	"sync"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

type peerMsg struct {
	head   common.Hash
	number *modules.ChainIndex
}

type peer struct {
	*p2p.Peer
	pubKey *ecdsa.PublicKey

	rw p2p.MsgReadWriter

	version int    // Protocol version negotiated
	network uint64 // Network ID being on

	//requestAnnounceType uint64

	id string

	headInfo peerMsg
	lock     sync.RWMutex

	announceChn chan announceData

	hasBlock func(common.Hash, uint64) bool
	//responseErrors int

	//fcClient       *flowcontrol.ClientNode // nil if the peer is server only
	//fcServer       *flowcontrol.ServerNode // nil if the peer is client only
	//fcServerParams *flowcontrol.ServerParams
	//fcCosts        requestCostTable
	//fullnode bool
}

func newPeer(version int, network uint64, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	id := p.ID()
	pubKey, _ := id.Pubkey()

	return &peer{
		Peer:        p,
		pubKey:      pubKey,
		rw:          rw,
		version:     version,
		network:     network,
		id:          fmt.Sprintf("%x", id[:8]),
		announceChn: make(chan announceData, 20),
	}
}

//
//func (p *peer) canQueue() bool {
//	//return p.sendQueue.canQueue()
//	return false
//}
//
//func (p *peer) queueSend(f func()) {
//	//p.sendQueue.queue(f)
//}

// Info gathers and returns a collection of metadata known about a peer.
func (p *peer) Info(asssetId modules.AssetId) *ptn.PeerInfo {
	hash, number := p.HeadAndNumber(asssetId)
	var index uint64
	if number != nil {
		index = number.Index
	}
	return &ptn.PeerInfo{
		Version: p.version,
		Index:   index,
		Head:    fmt.Sprintf("%x", hash),
	}
}

// Head retrieves a copy of the current head (most recent) hash of the peer.
func (p *peer) Head(assetId modules.AssetId) (hash common.Hash, number *modules.ChainIndex) {
	return p.HeadAndNumber(assetId)
}

func (p *peer) HeadAndNumber(assetId modules.AssetId) (hash common.Hash, number *modules.ChainIndex) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	copy(hash[:], p.headInfo.head[:])
	return hash, p.headInfo.number
}

func (p *peer) SetHead(header *modules.Header) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.headInfo = peerMsg{head: header.Hash(), number: header.Number}
}

// HasBlock checks if the peer has a given block
func (p *peer) HasBlock(hash common.Hash, number uint64) bool {
	p.lock.RLock()
	hasBlock := p.hasBlock
	p.lock.RUnlock()
	return hasBlock != nil && hasBlock(hash, number)
}

func (p *peer) SendSingleHeader(headers []*modules.Header) error {
	return p2p.Send(p.rw, CorsHeaderMsg, headers)
}

func (p *peer) SendHeaders(headers []*modules.Header) error {
	return p2p.Send(p.rw, CorsHeadersMsg, headers)
}

func (p *peer) RequestCurrentHeader(number modules.ChainIndex) error {
	return p2p.Send(p.rw, GetCurrentHeaderMsg, number)
}

func (p *peer) SendCurrentHeader(headers []*modules.Header) error {
	return p2p.Send(p.rw, CurrentHeaderMsg, headers)
}

func (p *peer) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	log.Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip,
		"reverse", reverse)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Hash: origin},
		Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}
func (p *peer) RequestHeadersByNumber(origin *modules.ChainIndex, amount int, skip int, reverse bool) error {
	log.Debug("Fetching batch of headers", "count", amount, "index", origin.Index, "skip", skip,
		"reverse", reverse)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Number: *origin},
		Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// SendBlockHeaders sends a batch of block headers to the remote peer.
func (p *peer) SendUnitHeaders(headers []*modules.Header) error {
	return p2p.Send(p.rw, BlockHeadersMsg, headers)
}

//interface
func (p *peer) RequestBodies([]common.Hash) error {
	return nil
}
func (p *peer) RequestNodeData([]common.Hash) error {
	return nil
}

func (p *peer) RequestNodeDataHead(modules.AssetId) (common.Hash, *modules.ChainIndex) {
	return common.Hash{}, nil
}
func (p *peer) RequestNodeDataRequestHeadersByHash(common.Hash, int, int, bool) error {
	return nil
}
func (p *peer) RequestNodeDataRequestHeadersByNumber(*modules.ChainIndex, int, int, bool) error {
	return nil
}
func (p *peer) RequestNodeDataRequestDagHeadersByHash(common.Hash, int, int, bool) error {
	return nil
}
func (p *peer) RequestNodeDataRequestLeafNodes() error {
	return nil
}
func (p *peer) RequestLeafNodes() error {
	return nil
}

type keyValueEntry struct {
	Key   string
	Value rlp.RawValue
}
type keyValueList []keyValueEntry
type keyValueMap map[string]rlp.RawValue

func (l keyValueList) add(key string, val interface{}) keyValueList {
	var entry keyValueEntry
	entry.Key = key
	if val == nil {
		val = uint64(0)
	}
	enc, err := rlp.EncodeToBytes(val)
	if err == nil {
		entry.Value = enc
	}
	return append(l, entry)
}

func (l keyValueList) decode() keyValueMap {
	m := make(keyValueMap)
	for _, entry := range l {
		m[entry.Key] = entry.Value
	}
	return m
}

func (m keyValueMap) get(key string, val interface{}) error {
	enc, ok := m[key]
	if !ok {
		return errResp(ErrMissingKey, "%s", key)
	}
	if val == nil {
		return nil
	}
	return rlp.DecodeBytes(enc, val)
}

func (p *peer) sendReceiveHandshake(sendList keyValueList) (keyValueList, error) {
	// Send out own handshake in a new thread
	errc := make(chan error, 1)
	go func() {
		errc <- p2p.Send(p.rw, StatusMsg, sendList)
	}()
	// In the mean time retrieve the remote status message
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return nil, err
	}
	if msg.Code != StatusMsg {
		return nil, errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, StatusMsg)
	}
	if msg.Size > ProtocolMaxMsgSize {
		return nil, errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// Decode the handshake
	var recvList keyValueList
	if err := msg.Decode(&recvList); err != nil {
		return nil, errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if err := <-errc; err != nil {
		return nil, err
	}
	return recvList, nil
}

// Handshake executes the les protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *peer) Handshake(number *modules.ChainIndex, genesis common.Hash, headhash common.Hash, assetId modules.AssetId,
	pcs []*modules.PartitionChain) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	var send keyValueList
	send = send.add("protocolVersion", uint64(p.version))
	send = send.add("networkId", p.network)
	send = send.add("headNum", *number)
	send = send.add("headHash", headhash)
	send = send.add("genesisHash", genesis)
	send = send.add("gastoken", assetId)

	recvList, err := p.sendReceiveHandshake(send)
	if err != nil {
		return err
	}
	recv := recvList.decode()

	var rGenesis, rHash common.Hash
	var rVersion, rNetwork uint64
	var rNum modules.ChainIndex
	var rGastoken modules.AssetId

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
	if err := recv.get("gastoken", &rGastoken); err != nil {
		return err
	}

	if pcs != nil {
		flag := 0
		for _, pc := range pcs {
			pcHash := pc.GetGenesisHeader().Hash()
			if rGenesis != pcHash {
				log.Debugf("ErrGenesisBlockMismatch , %x (!= %x)", rGenesis[:8], pcHash[:8])
				continue
			}
			if rNetwork != pc.NetworkId {
				log.Debugf("ErrNetworkIdMismatch, %d (!= %d)", rNetwork, pc.NetworkId)
				continue
			}
			if rVersion != pc.Version {
				log.Debugf("ErrProtocolVersionMismatch %d (!= %d)", rVersion, pc.Version)
				continue
			}

			for _, peer := range pc.Peers {
				node, err := discover.ParseNode(peer)
				if err != nil {
					continue
				}
				if p.id == node.ID.TerminalString() {
					flag = 1
					break
				}
			}
			break
		}
		if flag != 1 && len(pcs) > 0 {
			return errResp(ErrRequestRejected, "Not Accessed,p.id:%v", p.id)
		}
	} else {
		return errResp(ErrRequestRejected, "Not Registered,p.id:%v", p.id)
	}

	log.Debug("Cors Handshake", "p.ID()", p.ID(), "genesis", rGenesis, "network", rNetwork,
		"version", rVersion, "gastoken", rGastoken)
	p.headInfo = peerMsg{head: rHash, number: &rNum}
	return nil
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("cors/%d", p.version),
	)
}

// peerSet represents the collection of active peers currently participating in
// the Light Ethereum sub-protocol.
type peerSet struct {
	peers map[string]*peer
	lock  sync.RWMutex
	//notifyList []peerSetNotify
	closed bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errClosed
	}
	if _, ok := ps.peers[p.id]; ok {
		return errAlreadyRegistered
	}
	ps.peers[p.id] = p
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if _, ok := ps.peers[id]; !ok {
		return errNotRegistered
	}
	delete(ps.peers, id)
	return nil
}

// AllPeerIDs returns a list of all registered peer IDs
func (ps *peerSet) AllPeerIDs() []string {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	res := make([]string, len(ps.peers))
	idx := 0
	for id := range ps.peers {
		res[idx] = id
		idx++
	}
	return res
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// BestPeer retrieves the known peer with the currently highest total difficulty.
func (ps *peerSet) BestPeer() *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	var (
		bestPeer *peer
		bestTd   uint64
	)
	for _, p := range ps.peers {
		if number := p.headInfo.number; bestPeer == nil || number.Index > bestTd {
			bestPeer, bestTd = p, number.Index
		}
	}
	return bestPeer
}

// AllPeers returns all peers in a list
func (ps *peerSet) AllPeers() []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, len(ps.peers))
	i := 0
	for _, peer := range ps.peers {
		list[i] = peer
		i++
	}
	return list
}

// Close disconnects all peers.
// No new peers can be registered after Close has returned.
func (ps *peerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.Disconnect(p2p.DiscQuitting)
	}
	ps.closed = true
}
