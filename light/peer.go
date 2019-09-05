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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/light/flowcontrol"
	"github.com/palletone/go-palletone/ptn"
	"sync"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

const (
	announceTypeSimple = iota + 1
	announceTypeSigned
)

type peer struct {
	*p2p.Peer
	pubKey *ecdsa.PublicKey

	rw p2p.MsgReadWriter

	version int    // Protocol version negotiated
	network uint64 // Network ID being on

	announceType, requestAnnounceType uint64

	id string

	//headInfo *announceData
	//lock     sync.RWMutex

	lightpeermsg map[modules.AssetId]*announceData
	lightlock    sync.RWMutex

	announceChn chan announceData
	sendQueue   *execQueue

	//poolEntry *poolEntry
	hasBlock func(common.Hash, uint64) bool
	//	responseErrors int

	fcClient       *flowcontrol.ClientNode // nil if the peer is server only
	fcServer       *flowcontrol.ServerNode // nil if the peer is client only
	fcServerParams *flowcontrol.ServerParams
	fcCosts        requestCostTable
	fullnode       bool
}

func newPeer(version int, network uint64, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	id := p.ID()
	pubKey, _ := id.Pubkey()

	return &peer{
		Peer:         p,
		pubKey:       pubKey,
		rw:           rw,
		version:      version,
		network:      network,
		id:           fmt.Sprintf("%x", id[:8]),
		announceChn:  make(chan announceData, 20),
		lightpeermsg: map[modules.AssetId]*announceData{},
	}
}

// Info gathers and returns a collection of metadata known about a peer.
func (p *peer) Info(assetId modules.AssetId) *ptn.PeerInfo {
	hash, number := p.HeadAndNumber(assetId)
	return &ptn.PeerInfo{
		Version: p.version,
		Index:   number.Index, //uint64(0),
		Head:    fmt.Sprintf("%x", hash),
	}
}

func (p *peer) SetHead(data *announceData) {
	p.lightlock.Lock()
	defer p.lightlock.Unlock()
	data.Number = *data.Header.Number
	data.Hash = data.Header.Hash()
	p.lightpeermsg[data.Number.AssetID] = data

}

// Head retrieves a copy of the current head (most recent) hash of the peer.
func (p *peer) Head(assetid modules.AssetId) (hash common.Hash) {
	p.lightlock.RLock()
	defer p.lightlock.RUnlock()
	if v, ok := p.lightpeermsg[assetid]; ok {
		copy(hash[:], v.Hash[:])
		return hash
	}
	return hash
}

func (p *peer) HeadAndNumber(assetid modules.AssetId) (hash common.Hash, number *modules.ChainIndex) {
	p.lightlock.RLock()
	defer p.lightlock.RUnlock()

	if v, ok := p.lightpeermsg[assetid]; ok {
		copy(hash[:], v.Hash[:])
		return hash, &v.Number
	}
	return hash, nil
}

func sendRequest(w p2p.MsgWriter, msgcode, reqID uint64, data interface{}) error {
	type req struct {
		ReqID uint64
		Data  interface{}
	}
	return p2p.Send(w, msgcode, req{reqID, data})
}

func sendResponse(w p2p.MsgWriter, msgcode, reqID, bv uint64, data interface{}) error {
	type resp struct {
		ReqID, BV uint64
		Data      interface{}
	}
	return p2p.Send(w, msgcode, resp{reqID, bv, data})
}

func (p *peer) GetRequestCost(msgcode uint64, amount int) uint64 {
	p.lightlock.RLock()
	defer p.lightlock.RUnlock()

	cost := p.fcCosts[msgcode].baseCost + p.fcCosts[msgcode].reqCost*uint64(amount)
	if cost > p.fcServerParams.BufLimit {
		cost = p.fcServerParams.BufLimit
	}
	return cost
}

// HasBlock checks if the peer has a given block
func (p *peer) HasBlock(hash common.Hash, number uint64) bool {
	p.lightlock.RLock()
	defer p.lightlock.RUnlock()
	hasBlock := p.hasBlock
	return hasBlock != nil && hasBlock(hash, number)
}

func (p *peer) SendRawAnnounce(request []byte /*announceData*/) error {
	return p2p.Send(p.rw, AnnounceMsg, request)
}

// SendBlockHeaders sends a batch of block headers to the remote peer.
func (p *peer) SendUnitHeaders(reqID, bv uint64, headers []*modules.Header) error {
	return sendResponse(p.rw, BlockHeadersMsg, reqID, bv, headers)
}

func (p *peer) SendLeafNodes(reqID, bv uint64, headers []*modules.Header) error {
	return sendResponse(p.rw, LeafNodesMsg, reqID, bv, headers)
}

func (p *peer) SendRawUTXOs(reqID, bv uint64, utxos [][][]byte) error {
	return p2p.Send(p.rw, UTXOsMsg, utxos)
}

// SendProofs sends a batch of legacy LES/1 merkle proofs, corresponding to the ones requested.
func (p *peer) SendRawProofs(reqID, bv uint64, proofs [][][]byte) error {
	//log.Debug("Light PalleOne SendProofs", "len", len(proofs))
	return p2p.Send(p.rw, ProofsMsg, proofs)
	//return sendResponse(p.rw, ProofsV1Msg, reqID, bv, proofs)
}
func (p *peer) SendProofs(reqID, bv uint64, proof proofsRespData) error {
	//log.Debug("Light PalleOne SendProofs", "len", len(proofs))
	return p2p.Send(p.rw, ProofsMsg, proof)
	//return sendResponse(p.rw, ProofsV1Msg, reqID, bv, proofs)
}

// RequestHeadersByHash fetches a batch of blocks' headers corresponding to the
// specified header query, based on the hash of an origin block.
func (p *peer) RequestHeadersByHash(reqID, cost uint64, origin common.Hash, amount int, skip int, reverse bool) error {
	log.Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip,
		"reverse", reverse, "cost", cost)
	return sendRequest(p.rw, GetBlockHeadersMsg, reqID, &getBlockHeadersData{Origin: hashOrNumber{Hash: origin},
		Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// RequestHeadersByNumber fetches a batch of blocks' headers corresponding to the
// specified header query, based on the number of an origin block.
func (p *peer) RequestHeadersByNumber(reqID, cost uint64, origin modules.ChainIndex, amount int, skip int,
	reverse bool) error {
	log.Debug("Fetching batch of headers", "count", amount, "fromnum", origin, "skip", skip,
		"reverse", reverse, "cost", cost)
	//return nil
	return sendRequest(p.rw, GetBlockHeadersMsg, reqID, &getBlockHeadersData{Origin: hashOrNumber{Number: origin},
		Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

func (p *peer) RequestUTXOs(reqID, cost uint64, addr string) error {
	log.Debug("Fetching batch of utxos", "addr", addr)
	return p2p.Send(p.rw, GetUTXOsMsg, addr)
}

// RequestProofs fetches a batch of merkle proofs from a remote node.
func (p *peer) RequestProofs(reqID, cost uint64, reqs []ProofReq) error {
	log.Debug("Fetching batch of proofs", "count", len(reqs))
	return p2p.Send(p.rw, GetProofsMsg, reqs)
}

// SendTxStatus sends a batch of transactions to be added to the remote transaction pool.
func (p *peer) SendTxs(reqID, cost uint64, txs modules.Transactions) error {
	log.Debug("Fetching batch of transactions", "count", len(txs))
	switch p.version {
	case lpv1:
		return p2p.Send(p.rw, SendTxMsg, txs) // old message format does not include reqID
	//case lpv2:
	//return sendRequest(p.rw, SendTxV2Msg, reqID, cost, txs)
	default:
		panic(nil)
	}
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
func (p *peer) Handshake(number *modules.ChainIndex, genesis common.Hash, server *LesServer, headhash common.Hash,
	assetids [][][]byte) error {
	p.lightlock.Lock()
	defer p.lightlock.Unlock()

	var send keyValueList
	send = send.add("protocolVersion", uint64(p.version))
	send = send.add("networkId", p.network)
	send = send.add("headNum", *number)
	send = send.add("headHash", headhash)
	send = send.add("genesisHash", genesis)
	send = send.add("assetids", assetids)

	if server != nil {
		send = send.add("serveHeaders", nil)
		send = send.add("serveChainSince", uint64(0))
		send = send.add("serveStateSince", uint64(0))
		send = send.add("txRelay", nil)
		send = send.add("flowControl/BL", server.defParams.BufLimit)
		send = send.add("flowControl/MRR", server.defParams.MinRecharge)
		//list := server.fcCostStats.getCurrentList()
		//send = send.add("flowControl/MRC", list)
		send = send.add("fullnode", nil)
		//p.fcCosts = list.decode()
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
		// until we have a proper peer connectivity API, allow LES connection to other servers
		//if recv.get("serveStateSince", nil) == nil {
		//	return errResp(ErrUselessPeer, "wanted client, got server")
		//}
		if recv.get("announceType", &p.announceType) != nil {
			p.announceType = announceTypeSimple
		}
		p.fcClient = flowcontrol.NewClientNode(server.fcManager, server.defParams)
	} else {
		if recv.get("serveChainSince", nil) != nil {
			return errResp(ErrUselessPeer, "peer cannot serve chain")
		}
		if recv.get("serveStateSince", nil) != nil {
			return errResp(ErrUselessPeer, "peer cannot serve state")
		}
		if recv.get("txRelay", nil) != nil {
			return errResp(ErrUselessPeer, "peer cannot relay transactions")
		}
		params := &flowcontrol.ServerParams{}
		if err := recv.get("flowControl/BL", &params.BufLimit); err != nil {
			return err
		}
		if err := recv.get("flowControl/MRR", &params.MinRecharge); err != nil {
			return err
		}
		//var MRC RequestCostList
		//if err := recv.get("flowControl/MRC", &MRC); err != nil {
		//	return err
		//}
		p.fcServerParams = params
		p.fcServer = flowcontrol.NewServerNode(params)
		//p.fcCosts = MRC.decode()
	}

	if err := recv.get("fullnode", nil); err != nil {
		p.fullnode = false
	} else {
		p.fullnode = true
	}
	var rAssetIds [][][]byte
	if err := recv.get("assetids", &rAssetIds); err != nil {
		return err
	}
	for _, data := range rAssetIds {
		var hash common.Hash
		var number modules.ChainIndex
		hash.SetBytes(data[0])
		number.SetBytes(data[1])
		log.Debug("Light PalletOne Handshake all assetids", "assetid", number.AssetID, "index", number.Index)
		p.lightpeermsg[number.AssetID] = &announceData{Hash: hash, Number: number}
	}
	p.lightpeermsg[rNum.AssetID] = &announceData{Hash: rHash, Number: rNum}
	return nil
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("les/%d", p.version),
	)
}

// peerSetNotify is a callback interface to notify services about added or
// removed peers
type peerSetNotify interface {
	registerPeer(*peer)
	unregisterPeer(*peer)
}

// peerSet represents the collection of active peers currently participating in
// the Light Ethereum sub-protocol.
type peerSet struct {
	peers      map[string]*peer
	lock       sync.RWMutex
	notifyList []peerSetNotify
	closed     bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

// notify adds a service to be notified about added or removed peers
func (ps *peerSet) notify(n peerSetNotify) {
	ps.lock.Lock()
	ps.notifyList = append(ps.notifyList, n)
	peers := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		peers = append(peers, p)
	}
	ps.lock.Unlock()

	for _, p := range peers {
		n.registerPeer(p)
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	if ps.closed {
		ps.lock.Unlock()
		return errClosed
	}
	if _, ok := ps.peers[p.id]; ok {
		ps.lock.Unlock()
		return errAlreadyRegistered
	}
	ps.peers[p.id] = p
	p.sendQueue = newExecQueue(100)
	peers := make([]peerSetNotify, len(ps.notifyList))
	copy(peers, ps.notifyList)
	ps.lock.Unlock()

	for _, n := range peers {
		n.registerPeer(p)
	}
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity. It also initiates disconnection at the networking layer.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	if p, ok := ps.peers[id]; !ok {
		ps.lock.Unlock()
		return errNotRegistered
	} else {
		delete(ps.peers, id)
		peers := make([]peerSetNotify, len(ps.notifyList))
		copy(peers, ps.notifyList)
		ps.lock.Unlock()

		for _, n := range peers {
			n.unregisterPeer(p)
		}
		p.sendQueue.quit()
		p.Peer.Disconnect(p2p.DiscUselessPeer)
		return nil
	}
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
func (ps *peerSet) BestPeer(assetid modules.AssetId) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	var (
		bestPeer *peer
		bestTd   uint64
	)
	for _, p := range ps.peers {
		p.lightlock.RLock()
		defer p.lightlock.RUnlock()
		if v, ok := p.lightpeermsg[assetid]; ok {
			if number := v.Number; bestPeer == nil || number.Index > bestTd {
				bestPeer, bestTd = p, number.Index
			}
		}
	}
	return bestPeer
}

// AllPeers returns all peers in a list
func (ps *peerSet) AllPeers(assetid modules.AssetId) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := []*peer{}
	for _, peer := range ps.peers {
		peer.lightlock.RLock()
		defer peer.lightlock.RUnlock()
		if _, ok := peer.lightpeermsg[assetid]; ok {
			list = append(list, peer)
		}
	}
	return list
}

func (ps *peerSet) PeersWithoutHeader(assetid modules.AssetId, hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, len(ps.peers))
	i := 0
	for _, peer := range ps.peers {
		if peer.Head(assetid).String() != hash.String() {
			list[i] = peer
			i++
		}
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
