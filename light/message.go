package light

import (
	"encoding/json"
	"fmt"

	"bytes"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

func (pm *ProtocolManager) StatusMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received status message")
	// Status messages should never arrive after the handshake
	return errResp(ErrExtraStatusMsg, "uncontrolled status message")
}

// Block header query, collect the requested headers and reply
func (pm *ProtocolManager) AnnounceMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received announce message")
	var req announceData
	var data []byte
	if err := msg.Decode(&data); err != nil {
		log.Error("AnnounceMsg", "Decode err", err, "msg", msg)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	if err := json.Unmarshal(data, &req.Header); err != nil {
		log.Error("AnnounceMsg", "Unmarshal err", err, "data", data)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	if p.requestAnnounceType == announceTypeSigned {
		if err := req.checkSignature(p.pubKey); err != nil {
			log.Trace("Invalid announcement signature", "err", err)
			return err
		}
		log.Trace("Valid announcement signature")
	}

	if pm.IsExistInCache(req.Header.Hash().Bytes()) {
		//log.Debugf("Received unit(%v) again, ignore it", unitHash.TerminalString())
		p.SetHead(&req)
		return nil
	}

	log.Debug("Light PalletOne Announce message content", "p.id", p.id, "assetid", req.Header.Number.AssetID,
		"index", req.Header.Number.Index, "hash", req.Header.Hash())

	if pm.lightSync || pm.assetId != req.Header.Number.AssetID {
		pm.fetcher.Enqueue(p, &req.Header)
		localheader := pm.dag.CurrentHeader(req.Header.Number.AssetID)
		if localheader != nil {
			log.Debug("Light PalletOne Announce message pre synchronize", "localheader index",
				localheader.Number.Index, "req.Header.Number.Index-1", req.Header.Number.Index-1)
		} else {
			log.Debug("Light PalletOne Announce message pre synchronize",
				"localheader is ni.req.Header.Number.Index-1", req.Header.Number.Index-1)
		}

		if localheader == nil || localheader.Number.Index < req.Header.Number.Index-1 {
			p.SetHead(&req)
			go func() {
				log.Debug("Light PalletOne Announce message Enter synchronize")
				pm.synchronize(p, req.Header.Number.AssetID)
			}()
		}
	}

	return nil
}

func (pm *ProtocolManager) GetBlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// Decode the complex header query
	log.Debug("===Enter Light GetBlockHeadersMsg===")
	defer log.Debug("===End Ligth GetBlockHeadersMsg===")

	var query getBlockHeadersData
	if err := msg.Decode(&query); err != nil {
		log.Info("GetBlockHeadersMsg Decode", "err:", err, "msg:", msg)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	log.Debug("ProtocolManager", "GetBlockHeadersMsg getBlockHeadersData:", query)

	hashMode := query.Origin.Hash != (common.Hash{})
	log.Debug("ProtocolManager", "GetBlockHeadersMsg hashMode:", hashMode)
	// Gather headers until the fetch or network limits is reached
	var (
		bytes   common.StorageSize
		headers []*modules.Header
		unknown bool
	)

	for !unknown && len(headers) < int(query.Amount) &&
		bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
		// Retrieve the next header satisfying the query
		var origin *modules.Header
		if hashMode {
			origin, _ = pm.dag.GetHeaderByHash(query.Origin.Hash)
		} else {
			log.Debug("ProtocolManager GetBlockHeadersMsg", "assetid", query.Origin.Number.AssetID,
				" index", query.Origin.Number.Index)
			origin, _ = pm.dag.GetHeaderByNumber(&query.Origin.Number)
		}

		if origin == nil {
			break
		}
		log.Debug("ProtocolManager", "GetBlockHeadersMsg origin index:", origin.Number.Index)

		number := origin.Number.Index
		headers = append(headers, origin)
		bytes += estHeaderRlpSize

		// Advance to the next header of the query in light
		switch {
		case hashMode && query.Reverse:
			// Hash based traversal towards the genesis block
			log.Debug("Light ProtocolManager GetBlockHeadersMsg Hash based towards the genesis block")
			for i := 0; i < int(query.Skip)+1; i++ {
				if header, err := pm.dag.GetHeaderByHash(query.Origin.Hash); err == nil && header != nil {
					if number != 0 {
						query.Origin.Hash = header.ParentsHash[0]
					}
					number--
				} else {
					unknown = true
					break
				}
			}
		case hashMode && !query.Reverse:
			// Hash based traversal towards the leaf block
			log.Debug("Light ProtocolManager GetBlockHeadersMsg Hash based towards the leaf block")
			var (
				current = origin.Number.Index
				next    = current + query.Skip + 1
				index   = origin.Number
			)
			log.Debug("Light ProtocolManager GetBlockHeadersMsg", "next", next, "current:", current)
			if next <= current {
				infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
				log.Debug("Light ProtocolManager GetBlockHeaders skip overflow attack", "current", current,
					"skip", query.Skip, "next", next, "attacker", infos)
				unknown = true
			} else {
				index.Index = next
				log.Debug("Light ProtocolManager GetBlockHeadersMsg", "Index:", index.Index)
				if header, _ := pm.dag.GetHeaderByNumber(index); header != nil {
					hashs := pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)
					log.Debug("Light ProtocolManager", "GetUnitHashesFromHash len(hashs):", len(hashs),
						"header.index:", header.Number.Index, "header.hash:", header.Hash().String(),
						"query.Skip+1", query.Skip+1)
					if len(hashs) > int(query.Skip) && (hashs[query.Skip] == query.Origin.Hash) {
						query.Origin.Hash = header.Hash()
					} else {
						log.Debug("Light ProtocolManager GetBlockHeadersMsg ", "unknown = true; "+
							"pm.dag.GetUnitHashesFromHash not equal origin hash.", "")
						log.Debug("ProtocolManager", "GetBlockHeadersMsg header.Hash()", header.Hash(),
							"query.Skip+1:", query.Skip+1, "query.Origin.Hash:", query.Origin.Hash)
						unknown = true
					}
				} else {
					log.Debug("Light ProtocolManager GetBlockHeadersMsg ", "unknown = true; "+
						"pm.dag.GetHeaderByNumber not found. Index:", index.Index)
					unknown = true
				}
			}
		case query.Reverse:
			// Number based traversal towards the genesis block
			log.Debug("Light ProtocolManager GetBlockHeadersMsg Number based towards the genesis block")
			if query.Origin.Number.Index >= query.Skip+1 {
				query.Origin.Number.Index -= query.Skip + 1
			} else {
				log.Info("Light ProtocolManager GetBlockHeadersMsg query.Reverse unknown is true")
				unknown = true
			}

		case !query.Reverse:
			// Number based traversal towards the leaf block
			log.Debug("Light ProtocolManager GetBlockHeadersMsg Number based towards the leaf block")
			query.Origin.Number.Index += query.Skip + 1
		}
	}
	number := len(headers)
	if number > 0 {
		log.Debug("Light ProtocolManager GetBlockHeadersMsg", "query.Amount", query.Amount, "send number:",
			number, "start:", headers[0].Number.Index, "end:", headers[number-1].Number.Index,
			" getBlockHeadersData:", query)
	}
	log.Debug("Light ProtocolManager GetBlockHeadersMsg", "query.Amount", query.Amount, "send number:", 0,
		" getBlockHeadersData:", query)
	return p.SendUnitHeaders(0, 0, headers)
}

func (pm *ProtocolManager) BlockHeadersMsg(msg p2p.Msg, p *peer) error {
	if pm.downloader == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received block header response message")
	// A batch of headers arrived to one of our previous requests
	var resp struct {
		ReqID, BV uint64
		Headers   []*modules.Header
	}
	if err := msg.Decode(&resp); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	err := pm.downloader.DeliverHeaders(p.id, resp.Headers)
	if err != nil {
		log.Debug(fmt.Sprint(err))
	}
	return nil
}

func (pm *ProtocolManager) GetProofsMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received proofs request")
	// Decode the retrieval message
	var reqs []ProofReq
	if err := msg.Decode(&reqs); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	datas := [][][]byte{}
	for _, req := range reqs {
		log.Debug("Light PalletOne ProtocolManager GetProofsMsg", "req", req)
		//Get txRootHash and indexer in tx array and validation path
		//pm.dag.gettx
		//core.GetTrieInfo()
		tx, err := pm.dag.GetTransaction(req.BHash)
		if err != nil {
			return err
		}
		unit, err := pm.dag.GetUnitByHash(tx.UnitHash)
		if err != nil {
			return err
		}
		index := 0
		for _, tx := range unit.Txs {
			if tx.Hash() == req.BHash {
				break
			}
			index++
		}

		tri, trieRootHash := core.GetTrieInfo(unit.Txs)
		log.Debug("Light PalletOne ProtocolManager GetProofsMsg", "unit TxRoot", unit.UnitHeader.TxRoot,
			"trieRootHash", trieRootHash)
		//var proof les.NodeList
		keybuf := new(bytes.Buffer)
		rlp.Encode(keybuf, uint(index))

		resp := new(proofsRespData)
		resp.index = req.Index
		resp.txhash = req.BHash
		resp.headerhash = unit.Hash()
		resp.txroothash = trieRootHash
		resp.key = keybuf.Bytes()

		if err := tri.Prove(resp.key, 0, &resp.pathData); err != nil {
			log.Debug("Light PalletOne", "Prove err", err, "key", resp.key, "level", req.FromLevel,
				"proof", resp.pathData)
		}
		log.Debug("Light PalletOne", "key", resp.key, "level", req.FromLevel, "proof", resp.pathData)
		log.Debugf("Light PalletOne GetProofsMsg recv %v", resp)

		data, err := resp.encode()
		if err != nil {
			log.Debug("====", "err", err)
		}
		log.Debug("", "data", data)
		datas = append(datas, data)
	}
	return p.SendRawProofs(0, 0, datas)
}

func (pm *ProtocolManager) ProofsMsg(msg p2p.Msg, p *peer) error {
	log.Trace("Received proofs response")
	datas := [][][]byte{}

	//var resp []les.NodeList
	if err := msg.Decode(&datas); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	for _, data := range datas {
		resp := new(proofsRespData)
		if err := resp.decode(data); err != nil {
			log.Debug("Light PalletOne ProtocolManager ProofsMsg", "err", err)
			return err
		}
		log.Debug("Light PalletOne ProtocolManager ProofsMsg", "resp.key", resp.key)
		pm.validation.AddSpvResp(resp)
	}
	return nil
}

func (pm *ProtocolManager) SendTxMsg(msg p2p.Msg, p *peer) error {
	if pm.txpool == nil {
		return errResp(ErrRequestRejected, "")
	}
	// Transactions arrived, parse all of them and deliver to the pool
	var txs []*modules.Transaction
	if err := msg.Decode(&txs); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	//TODO must recover
	//reqCnt := len(txs)
	//if reject(uint64(reqCnt), MaxTxSend) {
	//	return errResp(ErrRequestRejected, "")
	//}
	pm.txpool.AddRemotes(txs)

	//_, rcost := p.fcClient.RequestProcessed(costs.baseCost + uint64(reqCnt)*costs.reqCost)
	//pm.server.fcCostStats.update(msg.Code, uint64(reqCnt), rcost)
	return nil
}

func (pm *ProtocolManager) GetUTXOsMsg(msg p2p.Msg, p *peer) error {
	if pm.server == nil {
		return errors.New("this node can not service with download utxo server")
	}

	var addr string
	if err := msg.Decode(&addr); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	address, err := common.StringToAddress(addr)
	if err != nil {
		log.Error("Light PalletOne", "ProtocolManager->GetUTXOsMsg addr err", err, "addr:", addr)
		return err
	}
	respdata := NewUtxosRespData()
	utxos, err := pm.dag.GetAddrUtxos(address)
	if err != nil {
		log.Error("Light PalletOne", "ProtocolManager->GetUTXOsMsg GetAddrUtxos err", err, "addr:", addr)
		return err
	}
	respdata.addr = addr
	respdata.utxos = utxos

	datas, err := respdata.encode()
	if err != nil {
		log.Error("Light PalletOne", "ProtocolManager->GetUTXOsMsg GetAddrUtxos err", err,
			"respdata:", respdata)
		return err
	}
	log.Debug("Light PalletOne", "ProtocolManager->GetUTXOsMsg GetAddrUtxos respdata.addr:", respdata.addr,
		"len(datas)", len(datas), "datas", datas)
	return p.SendRawUTXOs(0, 0, datas)
}
func (pm *ProtocolManager) UTXOsMsg(msg p2p.Msg, p *peer) error {
	if pm.server != nil {
		return errors.New("this is server node")
	}
	var datas [][][]byte
	respdata := NewUtxosRespData()
	if err := msg.Decode(&datas); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	log.Debug("Light PalletOne", "ProtocolManager->UTXOsMsg respdata", datas)
	if err := respdata.decode(datas); err != nil {
		log.Error("Light PalletOne", "ProtocolManager->UTXOsMsg respdata.decode err", err, "datas:", datas)
		return err
	}

	return pm.utxosync.SaveUtxoView(respdata)
}

func (pm *ProtocolManager) GetLeafNodesMsg(msg p2p.Msg, p *peer) error {
	//log.Debug("Light PalletOne ProtocolManager->GetLeafNodesMsg", "p.id", p.id)
	headers, err := pm.dag.GetAllLeafNodes()
	if err != nil {
		log.Error("Light PalletOne ProtocolManager->GetLeafNodesMsg", "p.id", p.id, "GetAllLeafNodes err", err)
		return err
	}
	headers = append(headers, pm.dag.CurrentHeader(pm.assetId))
	log.Debug("Light PalletOne ProtocolManager->GetLeafNodesMsg", "p.id", p.id, "len(headers)", len(headers),
		"headers", headers)

	p.SendLeafNodes(0, 0, headers)
	return nil
}

func (pm *ProtocolManager) LeafNodesMsg(msg p2p.Msg, p *peer) error {
	//var headers []*modules.Header
	var resp struct {
		ReqID, BV uint64
		Headers   []*modules.Header
	}
	if err := msg.Decode(&resp); err != nil {
		log.Debug("Light PalletOne ProtocolManager->LeafNodesMsg rlp", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Debug("Light PalletOne ProtocolManager->LeafNodesMsg", "p.id", p.id, "len(headers)", len(resp.Headers),
		"headers", resp.Headers)
	err := pm.downloader.DeliverAllToken(p.id, resp.Headers)
	if err != nil {
		log.Debug("Failed to deliver headers", "err", err.Error())
	}
	//err := pm.downloader.DeliverHeaders(p.id, resp.Headers)
	//if err != nil {
	//	log.Debug(fmt.Sprint(err))
	//}
	return err
}
