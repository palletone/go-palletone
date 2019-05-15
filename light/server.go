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
	"encoding/binary"
	"math"
	//"sync"

	//"github.com/palletone/go-palletone/dag/modules"
	//"github.com/ethereum/go-ethereum/eth"
	//"github.com/ethereum/go-ethereum/les/flowcontrol"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"

	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/light/flowcontrol"
	"github.com/palletone/go-palletone/ptn"
	"math/rand"
	"sync"
	"time"
)

type LesServer struct {
	config          *ptn.Config
	protocolManager *ProtocolManager
	fcManager       *flowcontrol.ClientManager // nil if our node is client only
	fcCostStats     *requestCostStats
	defParams       *flowcontrol.ServerParams
	p2psrv          *p2p.Server
	privateKey      *ecdsa.PrivateKey
	quitSync        chan struct{}
}

func NewLesServer(ptn *ptn.PalletOne, config *ptn.Config, protocolname string) (*LesServer, error) {
	quitSync := make(chan struct{})
	gasToken := config.Dag.GetGasToken()
	genesis, err := ptn.Dag().GetGenesisUnit()
	if err != nil {
		log.Error("Light PalletOne New", "get genesis err:", err)
		return nil, err
	}

	pm, err := NewProtocolManager(false, newPeerSet(), config.NetworkId, gasToken, ptn.TxPool(),
		ptn.Dag(), ptn.EventMux(), genesis, quitSync, protocolname)
	if err != nil {
		log.Error("NewlesServer NewProtocolManager", "err", err)
		return nil, err
	}

	srv := &LesServer{
		config:          config,
		protocolManager: pm,
		quitSync:        quitSync,
	}

	pm.server = srv

	srv.defParams = &flowcontrol.ServerParams{
		BufLimit:    300000000,
		MinRecharge: 50000,
	}
	srv.fcManager = flowcontrol.NewClientManager(uint64(config.LightServ), 10, 1000000000)
	srv.fcCostStats = newCostStats(ptn.UnitDb())
	return srv, nil
}

func (s *LesServer) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start starts the LES server
func (s *LesServer) Start(srvr *p2p.Server) {
	s.p2psrv = srvr
	s.protocolManager.Start(s.config.LightPeers)
	s.privateKey = srvr.PrivateKey
	s.protocolManager.blockLoop()
}

// Stop stops the LES service
func (s *LesServer) Stop() {
	//s.chtIndexer.Close()
	//// bloom trie indexer is closed by parent bloombits indexer
	//s.fcCostStats.store()
	s.fcManager.Stop()
	go func() {
		<-s.protocolManager.noMorePeers
	}()
	s.protocolManager.Stop()
}

type requestCosts struct {
	baseCost, reqCost uint64
}

type requestCostTable map[uint64]*requestCosts

type RequestCostList []struct {
	MsgCode, BaseCost, ReqCost uint64
}

func (list RequestCostList) decode() requestCostTable {
	table := make(requestCostTable)
	for _, e := range list {
		table[e.MsgCode] = &requestCosts{
			baseCost: e.BaseCost,
			reqCost:  e.ReqCost,
		}
	}
	return table
}

type linReg struct {
	sumX, sumY, sumXX, sumXY float64
	cnt                      uint64
}

const linRegMaxCnt = 100000

func (l *linReg) add(x, y float64) {
	if l.cnt >= linRegMaxCnt {
		sub := float64(l.cnt+1-linRegMaxCnt) / linRegMaxCnt
		l.sumX -= l.sumX * sub
		l.sumY -= l.sumY * sub
		l.sumXX -= l.sumXX * sub
		l.sumXY -= l.sumXY * sub
		l.cnt = linRegMaxCnt - 1
	}
	l.cnt++
	l.sumX += x
	l.sumY += y
	l.sumXX += x * x
	l.sumXY += x * y
}

func (l *linReg) calc() (b, m float64) {
	if l.cnt == 0 {
		return 0, 0
	}
	cnt := float64(l.cnt)
	d := cnt*l.sumXX - l.sumX*l.sumX
	if d < 0.001 {
		return l.sumY / cnt, 0
	}
	m = (cnt*l.sumXY - l.sumX*l.sumY) / d
	b = (l.sumY / cnt) - (m * l.sumX / cnt)
	return b, m
}

func (l *linReg) toBytes() []byte {
	var arr [40]byte
	binary.BigEndian.PutUint64(arr[0:8], math.Float64bits(l.sumX))
	binary.BigEndian.PutUint64(arr[8:16], math.Float64bits(l.sumY))
	binary.BigEndian.PutUint64(arr[16:24], math.Float64bits(l.sumXX))
	binary.BigEndian.PutUint64(arr[24:32], math.Float64bits(l.sumXY))
	binary.BigEndian.PutUint64(arr[32:40], l.cnt)
	return arr[:]
}

func linRegFromBytes(data []byte) *linReg {
	if len(data) != 40 {
		return nil
	}
	l := &linReg{}
	l.sumX = math.Float64frombits(binary.BigEndian.Uint64(data[0:8]))
	l.sumY = math.Float64frombits(binary.BigEndian.Uint64(data[8:16]))
	l.sumXX = math.Float64frombits(binary.BigEndian.Uint64(data[16:24]))
	l.sumXY = math.Float64frombits(binary.BigEndian.Uint64(data[24:32]))
	l.cnt = binary.BigEndian.Uint64(data[32:40])
	return l
}

type requestCostStats struct {
	lock  sync.RWMutex
	db    ptndb.Database
	stats map[uint64]*linReg
}

type requestCostStatsRlp []struct {
	MsgCode uint64
	Data    []byte
}

var rcStatsKey = []byte("_requestCostStats")

func newCostStats(db ptndb.Database) *requestCostStats {
	stats := make(map[uint64]*linReg)
	for _, code := range reqList {
		stats[code] = &linReg{cnt: 100}
	}

	if db != nil {
		data, err := db.Get(rcStatsKey)
		var statsRlp requestCostStatsRlp
		if err == nil {
			err = rlp.DecodeBytes(data, &statsRlp)
		}
		if err == nil {
			for _, r := range statsRlp {
				if stats[r.MsgCode] != nil {
					if l := linRegFromBytes(r.Data); l != nil {
						stats[r.MsgCode] = l
					}
				}
			}
		}
	}

	return &requestCostStats{
		db:    db,
		stats: stats,
	}
}

func (s *requestCostStats) store() {
	s.lock.Lock()
	defer s.lock.Unlock()

	statsRlp := make(requestCostStatsRlp, len(reqList))
	for i, code := range reqList {
		statsRlp[i].MsgCode = code
		statsRlp[i].Data = s.stats[code].toBytes()
	}

	if data, err := rlp.EncodeToBytes(statsRlp); err == nil {
		s.db.Put(rcStatsKey, data)
	}
}

func (s *requestCostStats) getCurrentList() RequestCostList {
	s.lock.Lock()
	defer s.lock.Unlock()

	list := make(RequestCostList, len(reqList))
	//fmt.Println("RequestCostList")
	for idx, code := range reqList {
		b, m := s.stats[code].calc()
		//fmt.Println(code, s.stats[code].cnt, b/1000000, m/1000000)
		if m < 0 {
			b += m
			m = 0
		}
		if b < 0 {
			b = 0
		}

		list[idx].MsgCode = code
		list[idx].BaseCost = uint64(b * 2)
		list[idx].ReqCost = uint64(m * 2)
	}
	return list
}

func (s *requestCostStats) update(msgCode, reqCnt, cost uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	c, ok := s.stats[msgCode]
	if !ok || reqCnt == 0 {
		return
	}
	c.add(float64(reqCnt), float64(cost))
}

func (pm *ProtocolManager) blockLoop() {
	pm.wg.Add(1)
	headCh := make(chan modules.ChainHeadEvent, 10)
	headSub := pm.dag.SubscribeChainHeadEvent(headCh)
	go func() {
		var lastHead *modules.Header
		//lastBroadcastTd := common.Big0
		for {
			select {
			case ev := <-headCh:
				peers := pm.peers.AllPeers()
				if len(peers) > 0 {
					header := ev.Unit.Header()
					hash := header.Hash()
					number := header.Number.Index
					//td := core.GetTd(pm.chainDb, hash, number)
					if lastHead == nil || (header.Number.Index > lastHead.Number.Index) {
						lastHead = header
						log.Debug("Announcing block to peers", "number", number, "hash", hash)

						announce := announceData{Hash: hash, Number: *lastHead.Number, Header: *lastHead}
						var (
							signed         bool
							signedAnnounce announceData
						)

						for _, p := range peers {
							log.Debug("Light Palletone", "ProtocolManager->blockLoop p.announceType", p.announceType)
							switch p.announceType {

							case announceTypeSimple:
								select {
								case p.announceChn <- announce:
								default:
									pm.removePeer(p.id)
								}

							case announceTypeSigned:
								if !signed {
									signedAnnounce = announce
									signedAnnounce.sign(pm.server.privateKey)
									signed = true
								}

								select {
								case p.announceChn <- signedAnnounce:
								default:
									pm.removePeer(p.id)
								}
							}
						}
					}
				}
			case <-pm.quitSync:
				headSub.Unsubscribe()
				pm.wg.Done()
				return
			}
		}
	}()
}

func (pm *ProtocolManager) ReqProofByTxHash(strhash string) string {
	peers := pm.peers.AllPeers()
	vreq, err := pm.validation.AddSpvReq(strhash)
	if err != nil {
		return err.Error()
	}

	var req ProofReq
	req.BHash = vreq.txhash
	req.FromLevel = uint(0)
	req.Index = vreq.strindex
	//for _, p := range peers {
	//	p.RequestProofs(0, 0, []ProofReq{req})
	//}

	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(len(peers))
	p := peers[x]
	p.RequestProofs(0, 0, []ProofReq{req})

	result := vreq.Wait()
	if result == 0 {
		return "OK"
	} else if result == 1 {
		return "error"
	} else if result == 2 {
		return "timeout"
	}
	return "errors"
}

func (pm *ProtocolManager) ReqProofByRlptx(rlptx [][]byte) string {
	log.Debug("===========ReqProofByRlptx===========", "", rlptx)
	resp := proofsRespData{}
	resp.headerhash.SetBytes(rlptx[0])
	resp.key = rlptx[1]
	if err := rlp.DecodeBytes(rlptx[2], &resp.pathData); err != nil {
		return err.Error()
	}
	result, err := pm.validation.Check(&resp)
	if err != nil {
		return err.Error()
	}
	if result == 0 {
		return "OK"
	} else if result == 1 {
		return "error"
	} else if result == 2 {
		return "timeout"
	}
	return "errors"
}

func (pm *ProtocolManager) SyncUTXOByAddr(addr string) string {
	log.Debug("Light PalletOne", "ProtocolManager->SyncUTXOByAddr addr:", addr)
	_, err := common.StringToAddress(addr)
	if err != nil {
		log.Debug("Light PalletOne", "ProtocolManager->SyncUTXOByAddr err:", err)
		return err.Error()
	}

	req, err := pm.utxosync.AddUtxoSyncReq(addr)
	if err != nil {
		log.Debug("Light PalletOne", "ProtocolManager->SyncUTXOByAddr err:", err)
		return err.Error()
	}

	//random select peer to download GetAddrUtxos(addr)
	rand.Seed(time.Now().UnixNano())
	peers := pm.peers.AllPeers()
	x := rand.Intn(len(peers))
	p := peers[x]
	p.RequestUTXOs(0, 0, req.addr)

	result := req.Wait()
	if result == 0 {
		return "OK"
	} else if result == 1 {
		return "error"
	} else if result == 2 {
		return "timeout"
	}
	return "errors"
}

func (pm *ProtocolManager) AddPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	if pm.server.p2psrv == nil {
		return false, nil
	}
	// Try to add the url as a static peer and return
	node, err := discover.ParseNode(url)
	if err != nil {
		return false, fmt.Errorf("invalid pnode: %v", err)
	}
	pm.server.p2psrv.AddPeer(node)
	return true, nil
}

/*
type MainChain struct {
	GenesisHash common.Hash
	Status      byte //Active:1 ,Terminated:0,Suspended:2
	SyncModel   byte //Push:1 , Pull:2, Push+Pull:0
	GasToken    AssetId
	Peers       []string // IP:port format string
}
*/
func (pm *ProtocolManager) GetMainChain() (*modules.MainChain, error) {
	//contract.ccquery("PCGTta3M4t3yXu8uRgkKvaWd2d8DRxVdGDZ",["getMainChain"])
	return nil, fmt.Errorf("this is not cors protocol")
	//TODO
	if pm.protocolname != configure.CORSProtocol {
		return nil, fmt.Errorf("this is not cors protocol")
	}
	mainchain := &modules.MainChain{}
	mainchain.NetworkId = 1
	mainchain.Version = 1
	mainchain.GenesisHash.SetHexString("0x927c94780c89b450cf2d9bcb3febea8457bcb830f5867b9d85c74ce4df3d2ac4")
	return mainchain, nil
}
