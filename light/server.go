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
	//"sync"

	//"github.com/palletone/go-palletone/dag/modules"
	//"github.com/ethereum/go-ethereum/eth"
	//"github.com/ethereum/go-ethereum/les/flowcontrol"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/light/flowcontrol"
	"github.com/palletone/go-palletone/ptn"
	"math/rand"
	"time"
)

const (
	CodeErr        = "error"
	CodeOK         = "ok"
	CodeTimeout    = "timeout"
	CodeEmptyPeers = "empty peers"
)

type LesServer struct {
	config          *ptn.Config
	protocolManager *ProtocolManager
	fcManager       *flowcontrol.ClientManager // nil if our node is client only
	//fcCostStats     *requestCostStats
	defParams    *flowcontrol.ServerParams
	srv          *p2p.Server
	corss        *p2p.Server
	privateKey   *ecdsa.PrivateKey
	quitSync     chan struct{}
	protocolname string
	fullnode     *ptn.PalletOne
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
		protocolname:    protocolname,
		fullnode:        ptn,
	}

	pm.server = srv

	srv.defParams = &flowcontrol.ServerParams{
		BufLimit:    300000000,
		MinRecharge: 50000,
	}
	srv.fcManager = flowcontrol.NewClientManager(uint64(config.LightServ), 10, 1000000000)
	//srv.fcCostStats = newCostStats(ptn.UnitDb())
	return srv, nil
}

func (s *LesServer) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

func (s *LesServer) CorsProtocols() []p2p.Protocol {
	return nil
}
func (s *LesServer) StartCorsSync() (string, error) {
	return "", nil
}

// Start starts the LES server
func (s *LesServer) Start(srvr *p2p.Server, corss *p2p.Server, syncCh chan bool) {
	s.srv = srvr
	if s.protocolname == configure.CORSProtocol {
		s.corss = corss
	}

	s.protocolManager.Start(s.config.LightPeers, s.corss, syncCh)
	s.privateKey = srvr.PrivateKey
	s.protocolManager.blockLoop()
	s.loopCors()
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

func (s *LesServer) SubscribeCeEvent(ch chan<- *modules.Header) event.Subscription {
	return nil
}

func (s *LesServer) loopCors() {
	headCh := make(chan *modules.Header, 10)
	headSub := s.fullnode.CorsServer().SubscribeCeEvent(headCh)
	go func() {
		for {
			select {
			case header := <-headCh:
				peers := s.protocolManager.peers.AllPeers(s.protocolManager.assetId)
				log.Debug("LesServer loopCors Light recv Cors header", "len(peers)", len(peers), "assetid",
					header.Number.AssetID, "index", header.Number.Index, "hash", header.Hash())
				if len(peers) > 0 {
					announce := announceData{Hash: header.Hash(), Number: *header.Number, Header: *header}
					for _, p := range peers {
						p.announceChn <- announce
						//switch p.announceType {
						//case announceTypeSimple:
						//	select {
						//	case p.announceChn <- announce:
						//	default:
						//		s.protocolManager.removePeer(p.id)
						//	}
						//case announceTypeSigned:
						//}
					}
				}
			case <-s.quitSync:
				headSub.Unsubscribe()
				return
			}
		}
	}()
}

type requestCosts struct {
	baseCost, reqCost uint64
}

type requestCostTable map[uint64]*requestCosts

//type RequestCostList []struct {
//	MsgCode, BaseCost, ReqCost uint64
//}

//func (list RequestCostList) decode() requestCostTable {
//	table := make(requestCostTable)
//	for _, e := range list {
//		table[e.MsgCode] = &requestCosts{
//			baseCost: e.BaseCost,
//			reqCost:  e.ReqCost,
//		}
//	}
//	return table
//}
//
//type linReg struct {
//	sumX, sumY, sumXX, sumXY float64
//	cnt                      uint64
//}
//
//const linRegMaxCnt = 100000
//
//func (l *linReg) add(x, y float64) {
//	if l.cnt >= linRegMaxCnt {
//		sub := float64(l.cnt+1-linRegMaxCnt) / linRegMaxCnt
//		l.sumX -= l.sumX * sub
//		l.sumY -= l.sumY * sub
//		l.sumXX -= l.sumXX * sub
//		l.sumXY -= l.sumXY * sub
//		l.cnt = linRegMaxCnt - 1
//	}
//	l.cnt++
//	l.sumX += x
//	l.sumY += y
//	l.sumXX += x * x
//	l.sumXY += x * y
//}
//
//func (l *linReg) calc() (b, m float64) {
//	if l.cnt == 0 {
//		return 0, 0
//	}
//	cnt := float64(l.cnt)
//	d := cnt*l.sumXX - l.sumX*l.sumX
//	if d < 0.001 {
//		return l.sumY / cnt, 0
//	}
//	m = (cnt*l.sumXY - l.sumX*l.sumY) / d
//	b = (l.sumY / cnt) - (m * l.sumX / cnt)
//	return b, m
//}
//
//func (l *linReg) toBytes() []byte {
//	var arr [40]byte
//	binary.BigEndian.PutUint64(arr[0:8], math.Float64bits(l.sumX))
//	binary.BigEndian.PutUint64(arr[8:16], math.Float64bits(l.sumY))
//	binary.BigEndian.PutUint64(arr[16:24], math.Float64bits(l.sumXX))
//	binary.BigEndian.PutUint64(arr[24:32], math.Float64bits(l.sumXY))
//	binary.BigEndian.PutUint64(arr[32:40], l.cnt)
//	return arr[:]
//}

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
				peers := pm.peers.AllPeers(ev.Unit.UnitHeader.Number.AssetID)
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
	peers := pm.peers.AllPeers(pm.assetId)
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
		return CodeOK
	} else if result == 1 {
		return CodeErr
	} else if result == 2 {
		return CodeTimeout
	}
	return CodeErr
}

func (pm *ProtocolManager) ReqProofByRlptx(rlptx [][]byte) string {
	log.Debug("ReqProofByRlptx", "", rlptx)
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
		return CodeOK
	} else if result == 1 {
		return CodeErr
	} else if result == 2 {
		return CodeTimeout
	}
	return CodeErr
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
	peers := pm.peers.AllPeers(pm.assetId)
	if len(peers) <= 0 {
		return CodeEmptyPeers
	}
	x := rand.Intn(len(peers))
	p := peers[x]
	p.RequestUTXOs(0, 0, req.addr)

	result := req.Wait()
	if result == 0 {
		return CodeOK
	} else if result == 1 {
		return CodeErr
	} else if result == 2 {
		return CodeTimeout
	}
	return CodeErr
}
