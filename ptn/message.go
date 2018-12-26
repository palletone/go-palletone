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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package ptn

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/consensus/jury"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/ptn/downloader"
	"github.com/palletone/go-palletone/tokenengine"
)

type Tag uint64

func (pm *ProtocolManager) StatusMsg(msg p2p.Msg, p *peer) error {
	// Status messages should never arrive after the handshake
	return errResp(ErrExtraStatusMsg, "uncontrolled status message")
}

// Block header query, collect the requested headers and reply
func (pm *ProtocolManager) GetBlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// Decode the complex header query
	log.Debug("===Enter GetBlockHeadersMsg===")
	defer log.Debug("===End GetBlockHeadersMsg===")

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

	for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
		// Retrieve the next header satisfying the query
		var origin *modules.Header
		if hashMode {
			origin = pm.dag.GetHeaderByHash(query.Origin.Hash)
		} else {
			log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Origin.Number:", query.Origin.Number.Index)
			origin = pm.dag.GetHeaderByNumber(query.Origin.Number)
		}

		if origin == nil {
			break
		}
		log.Debug("ProtocolManager", "GetBlockHeadersMsg origin index:", origin.Number.Index)

		number := origin.Number.Index
		headers = append(headers, origin)
		bytes += estHeaderRlpSize

		// Advance to the next header of the query
		switch {
		case hashMode && query.Reverse:
			// Hash based traversal towards the genesis block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the genesis block")
			for i := 0; i < int(query.Skip)+1; i++ {
				if header, err := pm.dag.GetHeader(query.Origin.Hash, number); err == nil && header != nil {
					if number != 0 {
						query.Origin.Hash = header.ParentsHash[0]
					}
					number--
				} else {
					//log.Info("========GetBlockHeadersMsg========", "number", number, "err:", err)
					unknown = true
					break
				}
			}
		case hashMode && !query.Reverse:
			// Hash based traversal towards the leaf block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the leaf block")
			var (
				current = origin.Number.Index
				next    = current + query.Skip + 1
				index   = origin.Number
			)
			log.Debug("ProtocolManager", "GetBlockHeadersMsg next", next, "current:", current)
			if next <= current {
				infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
				log.Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
				unknown = true
			} else {
				index.Index = next
				log.Debug("ProtocolManager", "GetBlockHeadersMsg index.Index:", index.Index)
				if header := pm.dag.GetHeaderByNumber(index); header != nil {
					if pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip] == query.Origin.Hash {
						query.Origin.Hash = header.Hash()
					} else {
						log.Debug("ProtocolManager", "GetBlockHeadersMsg unknown = true; pm.dag.GetUnitHashesFromHash not equal origin hash.", "")
						log.Debug("ProtocolManager", "GetBlockHeadersMsg header.Hash()", header.Hash(), "query.Skip+1:", query.Skip+1, "query.Origin.Hash:", query.Origin.Hash)
						log.Debug("ProtocolManager", "GetBlockHeadersMsg pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip]:", pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip])
						unknown = true
					}
				} else {
					log.Debug("ProtocolManager", "GetBlockHeadersMsg unknown = true; pm.dag.GetHeaderByNumber not found. Index:", index.Index)
					unknown = true
				}
			}
		case query.Reverse:
			// Number based traversal towards the genesis block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the genesis block")
			if query.Origin.Number.Index >= query.Skip+1 {
				query.Origin.Number.Index -= query.Skip + 1
			} else {
				log.Info("ProtocolManager", "GetBlockHeadersMsg query.Reverse", "unknown is true")
				unknown = true
			}

		case !query.Reverse:
			// Number based traversal towards the leaf block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the leaf block")
			query.Origin.Number.Index += query.Skip + 1
		}
	}
	start := uint64(0)
	end := uint64(0)
	number := len(headers)
	if number > 0 {
		start = uint64(headers[0].Number.Index)
		end = uint64(headers[number-1].Number.Index)
	}
	log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Amount", query.Amount, "send number:", len(headers), "start:", start, "end:", end, " getBlockHeadersData:", query)
	return p.SendUnitHeaders(headers)
}

func (pm *ProtocolManager) BlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// A batch of headers arrived to one of our previous requests
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	log.Debug("ProtocolManager", "BlockHeadersMsg len(headers):", len(headers))
	// Filter out any explicitly requested headers, deliver the rest to the downloader
	filter := len(headers) == 1
	if filter {
		// Irrelevant of the fork checks, send the header to the fetcher just in case
		headers = pm.fetcher.FilterHeaders(p.id, headers, time.Now())
	}
	if len(headers) > 0 || !filter {
		log.Debug("===BlockHeadersMsg ===", "len(headers):", len(headers))
		err := pm.downloader.DeliverHeaders(p.id, headers)
		if err != nil {
			log.Debug("Failed to deliver headers", "err", err.Error())
		}
	}
	return nil
}

func (pm *ProtocolManager) GetBlockBodiesMsg(msg p2p.Msg, p *peer) error {
	// Decode the retrieval message
	log.Debug("Enter GetBlockBodiesMsg")
	defer log.Debug("End GetBlockBodiesMsg")
	msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if _, err := msgStream.List(); err != nil {
		log.Debug("msgStream.List() err:", err)
		return err
	}
	// Gather blocks until the fetch or network limits is reached
	var (
		hash   common.Hash
		bytes  int
		bodies [][]byte //rlp.RawValue
		//bodies []blockBody
	)

	for bytes < softResponseLimit && len(bodies) < downloader.MaxBlockFetch {
		// Retrieve the hash of the next block
		if err := msgStream.Decode(&hash); err == rlp.EOL {
			break
		} else if err != nil {
			log.Debug("msgStream.Decode", "err", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		log.Debug("GetBlockBodiesMsg", "hash", hash)
		// Retrieve the requested block body, stopping if enough was found
		txs, err := pm.dag.GetUnitTransactions(hash)
		if err != nil || len(txs) == 0 {
			log.Debug("GetBlockBodiesMsg", "hash:", hash, "GetUnitTransactions err:", err)
			//return errResp(ErrDecode, "msg %v: %v", msg, err)
			continue
		}

		//data, err := rlp.EncodeToBytes(txs)
		//if err != nil {
		//	log.Debug("Get body rlp when rlp encode", "unit hash", hash.String(), "error", err.Error())
		//	return errResp(ErrDecode, "msg %v: %v", msg, err)
		//}
		data, err := json.Marshal(txs)
		if err != nil {
			log.Debug("Get body Marshal encode", "error", err.Error(), "unit hash", hash.String())
			//return errResp(ErrDecode, "msg %v: %v", msg, err)
			continue
		}
		log.Debug("Get body Marshal", "data:", string(data))
		bytes += len(data)
		bodies = append(bodies, data)

		//log.Debug("GetBlockBodiesMsg", "hash", hash, "txs size:", len(txs))

		//body := blockBody{Transactions: txs}
		//bodies = append(bodies, body)
	}
	log.Debug("GetBlockBodiesMsg", "len(bodies):", len(bodies), "bytes:", bytes)
	//return p.SendBlockBodies(bodies)
	return p.SendBlockBodiesRLP(bodies)
}

func (pm *ProtocolManager) BlockBodiesMsg(msg p2p.Msg, p *peer) error {
	// A batch of block bodies arrived to one of our previous requests
	log.Debug("Enter ProtocolManager BlockBodiesMsg")
	defer log.Debug("End ProtocolManager BlockBodiesMsg")
	var request [][]byte //rlp.RawValue
	if err := msg.Decode(&request); err != nil {
		log.Info("BlockBodiesMsg msg Decode", "err:", err, "msg:", msg)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	//log.Debug("===BlockBodiesMsg===", "len(request:)", len(request))
	transactions := make([][]*modules.Transaction, len(request))
	for i, body := range request {
		if len(body) == 0 {
			continue
		}
		var txs modules.Transactions
		//log.Debug("BlockBodiesMsg", "have body:", string(body))
		if err := json.Unmarshal(body, &txs); err != nil {
			log.Debug("have body Unmarshal encode", "error", err.Error(), "body", string(body))
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		var temptxs modules.Transactions
		for _, tx := range txs {
			msgs, err1 := storage.ConvertMsg(tx)
			if err1 != nil {
				log.Error("tx comvertmsg failed......", "err:", err1, "tx:", tx)
				break
			}
			tx.TxMessages = msgs
			temptxs = append(temptxs, tx)
		}

		transactions[i] = temptxs
		log.Debug("BlockBodiesMsg", "i", i, "txs size:", len(temptxs))
	}
	log.Debug("===BlockBodiesMsg===", "len(transactions:)", len(transactions))
	// Filter out any explicitly requested bodies, deliver the rest to the downloader
	filter := len(transactions) > 0
	if filter {
		log.Debug("===BlockBodiesMsg->FilterBodies===")
		transactions = pm.fetcher.FilterBodies(p.id, transactions, time.Now())
	}
	if len(transactions) > 0 || !filter {
		log.Debug("===BlockBodiesMsg->DeliverBodies===")
		err := pm.downloader.DeliverBodies(p.id, transactions)
		if err != nil {
			log.Debug("Failed to deliver bodies", "err", err.Error())
		}
	}
	/*
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver them all to the downloader for queuing
		transactions := make([][]*modules.Transaction, len(request))
		for i, body := range request {
			transactions[i] = body.Transactions
			log.Info("BlockBodiesMsg", "i", i, "txs size:", len(body.Transactions))
		}

		log.Debug("===BlockBodiesMsg===", "len(transactions:)", len(transactions))
		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		filter := len(transactions) > 0
		if filter {
			log.Debug("===BlockBodiesMsg->FilterBodies===")
			transactions = pm.fetcher.FilterBodies(p.id, transactions, time.Now())
		}
		if len(transactions) > 0 || !filter {
			log.Debug("===BlockBodiesMsg->DeliverBodies===")
			err := pm.downloader.DeliverBodies(p.id, transactions)
			if err != nil {
				log.Debug("Failed to deliver bodies", "err", err.Error())
			}
		}
	*/
	return nil
}

func (pm *ProtocolManager) GetNodeDataMsg(msg p2p.Msg, p *peer) error {
	// Decode the retrieval message
	msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if _, err := msgStream.List(); err != nil {
		return err
	}
	// Gather state data until the fetch or network limits is reached
	var (
		hash  common.Hash
		bytes int
		data  [][]byte
	)
	for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
		// Retrieve the hash of the next state entry
		if err := msgStream.Decode(&hash); err == rlp.EOL {
			break
		} else if err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Retrieve the requested state entry, stopping if enough was found
		//if entry, err := pm.blockchain.TrieNode(hash); err == nil {
		//	data = append(data, entry)
		//	bytes += len(entry)
		//}
	}
	return p.SendNodeData(data)
}

func (pm *ProtocolManager) NodeDataMsg(msg p2p.Msg, p *peer) error {
	// A batch of node state data arrived to one of our previous requests
	var data [][]byte
	if err := msg.Decode(&data); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Deliver all to the downloader
	if err := pm.downloader.DeliverNodeData(p.id, data); err != nil {
		log.Debug("Failed to deliver node state data", "err", err.Error())
	}
	return nil
}

func (pm *ProtocolManager) NewBlockHashesMsg(msg p2p.Msg, p *peer) error {
	log.Debug("===NewBlockHashesMsg===")
	var announces newBlockHashesData
	if err := msg.Decode(&announces); err != nil {
		log.Debug("===NewBlockHashesMsg===", "Decode err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	// Mark the hashes as present at the remote node
	for _, block := range announces {
		p.MarkUnit(block.Hash)
	}
	// Schedule all the unknown hashes for retrieval
	unknown := make(newBlockHashesData, 0, len(announces))
	for _, block := range announces {
		if !pm.dag.HasUnit(block.Hash) {
			unknown = append(unknown, block)
		}
	}
	log.Debug("===NewBlockHashesMsg===", "len(unknown):", len(unknown))
	for _, block := range unknown {
		pm.fetcher.Notify(p.id, block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
	}
	return nil
}

func (pm *ProtocolManager) NewBlockMsg(msg p2p.Msg, p *peer) error {
	// Retrieve and decode the propagated block
	//unit := &modules.Unit{}
	data := []byte{}
	if err := msg.Decode(&data); err != nil {
		log.Info("ProtocolManager", "NewBlockMsg msg:", msg.String())
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	unit := &modules.Unit{}
	if err := json.Unmarshal(data, &unit); err != nil {
		log.Info("ProtocolManager", "NewBlockMsg json ummarshal err:", err)
		return err
	}

	var temptxs modules.Transactions
	for _, tx := range unit.Txs {
		msgs, err1 := storage.ConvertMsg(tx)
		if err1 != nil {
			log.Error("tx comvertmsg failed......", "err:", err1, "tx:", tx)
			return err1
		}
		tx.TxMessages = msgs
		temptxs = append(temptxs, tx)
	}
	unit.Txs = temptxs

	unit.ReceivedAt = msg.ReceivedAt
	unit.ReceivedFrom = p
	log.Debug("===NewBlockMsg===", "unit:", *unit, "index:", unit.Number().Index, "peer id:", p.id)

	// Mark the peer as owning the block and schedule it for import
	p.MarkUnit(unit.UnitHash)
	pm.fetcher.Enqueue(p.id, unit)

	requestNumber := unit.UnitHeader.Number
	hash, number := p.Head(unit.Number().AssetID)
	if common.EmptyHash(hash) || (!common.EmptyHash(hash) && requestNumber.Index > number.Index) {
		log.Debug("ProtocolManager", "NewBlockMsg SetHead request.Index:", unit.UnitHeader.ChainIndex().Index,
			"local peer index:", number.Index)
		trueHead := unit.Hash()
		p.SetHead(trueHead, requestNumber)
		requestIndex := requestNumber.Index
		currentUnitIndex := pm.dag.GetCurrentUnit(unit.Number().AssetID).UnitHeader.Number.Index

		if requestIndex > currentUnitIndex {
			log.Debug("ProtocolManager", "NewBlockMsg synchronise request.Index:", unit.UnitHeader.ChainIndex().Index,
				"current unit index:", currentUnitIndex)
			go func() {
				time.Sleep(100 * time.Millisecond)
				pm.synchronise(p, unit.Number().AssetID)
			}()
		}
	}
	return nil
}

func (pm *ProtocolManager) TxMsg(msg p2p.Msg, p *peer) error {
	log.Debug("Enter ProtocolManager TxMsg")
	defer log.Debug("End ProtocolManager TxMsg")
	// Transactions arrived, make sure we have a valid and fresh chain to handle them
	if atomic.LoadUint32(&pm.acceptTxs) == 0 {
		log.Debug("ProtocolManager handlmsg TxMsg pm.acceptTxs==0")
		return nil
	}
	// Transactions can be processed, parse all of them and deliver to the pool
	var txs []*modules.Transaction
	if err := msg.Decode(&txs); err != nil {
		log.Debug("ProtocolManager handlmsg TxMsg", "Decode err:", err, "msg:", msg)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	//TODO VerifyTX
	log.Debug("===============ProtocolManager", "TxMsg txs:", txs)
	for i, tx := range txs {
		// Validate and mark the remote transaction
		if tx == nil {
			return errResp(ErrDecode, "transaction %d is nil", i)
		}

		if tx.IsContractTx() {
			if !pm.contractProc.CheckContractTxValid(tx) {
				log.Debug("TxMsg", "CheckContractTxValid is false")
				return nil //errResp(ErrDecode, "msg %v: Contract transaction valid fail", msg)
			}
		}

		for msgIndex, msg := range tx.TxMessages {
			payload, ok := msg.Payload.(*modules.PaymentPayload)
			if ok == false {
				continue
			}
			for inputIndex, txin := range payload.Inputs {
				st, err := pm.dag.GetUtxoEntry(txin.PreviousOutPoint)
				if st == nil || err != nil {
					return err
				}
				err = tokenengine.ScriptValidate(st.PkScript, nil, tx, msgIndex, inputIndex)
				if err != nil {
					return err
				}
			}
		}
		p.MarkTransaction(tx.Hash())
		txHash := tx.Hash()
		txHash = txHash
		_, err := pm.txpool.ProcessTransaction(tx, true, true, 0 /*pm.txpool.Tag(peer.ID())*/)
		//acceptedTxs = acceptedTxs
		if err != nil {
			return errResp(ErrDecode, "transaction %d not accepteable ", i, "err:", err)
		}
	}

	log.Debug("===============ProtocolManager TxMsg AddRemotes====================")
	pm.txpool.AddRemotes(txs)
	return nil
}

func (pm *ProtocolManager) ConsensusMsg(msg p2p.Msg, p *peer) error {
	var consensusmsg string
	if err := msg.Decode(&consensusmsg); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Info("ConsensusMsg recv:", consensusmsg)
	if consensusmsg == "A" {
		p.SendConsensus("Hello I received A")
	}
	return nil
}

func (pm *ProtocolManager) NewProducedUnitMsg(msg p2p.Msg, p *peer) error {
	// Retrieve and decode the propagated new produced unit
	var unit modules.Unit
	if err := msg.Decode(&unit); err != nil {
		log.Info("===NewProducedUnitMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	pm.producer.ToUnitTBLSSign(&unit)
	return nil
}

func (pm *ProtocolManager) SigShareMsg(msg p2p.Msg, p *peer) error {
	var sigShare mp.SigShareEvent
	if err := msg.Decode(&sigShare); err != nil {
		log.Info("===SigShareMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	pm.producer.ToTBLSRecover(&sigShare)
	return nil
}

func (pm *ProtocolManager) VSSDealMsg(msg p2p.Msg, p *peer) error {
	// comment by Albert·Gou
	//var vssmsg vssMsg
	//if err := msg.Decode(&vssmsg); err != nil {

	var deal mp.VSSDealEvent
	if err := msg.Decode(&deal); err != nil {
		log.Info("===VSSDealMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	pm.producer.ToProcessDeal(&deal)

	// comment by Albert·Gou
	////TODO vssmark
	//if !pm.peers.PeersWithoutVss(vssmsg.NodeId) {
	//	pm.producer.ToProcessDeal(vssmsg.Deal)
	//	pm.peers.MarkVss(vssmsg.NodeId)
	//	pm.BroadcastVss(vssmsg.NodeId, vssmsg.Deal)
	//}
	return nil
}

func (pm *ProtocolManager) VSSResponseMsg(msg p2p.Msg, p *peer) error {
	var resp mp.VSSResponseEvent
	if err := msg.Decode(&resp); err != nil {
		log.Info("===VSSResponseMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	pm.producer.ToProcessResponse(&resp)
	return nil
}

//GroupSigMsg
func (pm *ProtocolManager) GroupSigMsg(msg p2p.Msg, p *peer) error {
	var gSign mp.GroupSigEvent
	if err := msg.Decode(&gSign); err != nil {
		log.Info("===GroupSigMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	pm.dag.SetUnitGroupSign(gSign.UnitHash, gSign.GroupSig, pm.txpool)
	return nil
}

func (pm *ProtocolManager) ContractExecMsg(msg p2p.Msg, p *peer) error {
	var event jury.ContractExeEvent
	if err := msg.Decode(&event); err != nil {
		log.Info("===ContractExecMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	pm.contractProc.ProcessContractEvent(&event)
	return nil
}

func (pm *ProtocolManager) ContractSigMsg(msg p2p.Msg, p *peer) error {
	var event jury.ContractSigEvent
	if err := msg.Decode(&event); err != nil {
		log.Info("===ContractExecMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	//pm.contractProc.ProcessContractSigEvent(&event)
	return nil
}

//local test
func (pm *ProtocolManager) ContractReqLocalSend(event jury.ContractExeEvent) {
	log.Info("ContractReqLocalSend", "event", event.Tx.Hash())
	pm.contractExecCh <- event
}

func (pm *ProtocolManager) ContractSigLocalSend(event jury.ContractSigEvent) {
	log.Info("ContractSigLocalSend", "event", event.Tx.Hash())
	pm.contractSigCh <- event
}

func (pm *ProtocolManager) ContractBroadcast(event jury.ContractExeEvent) {
	log.Debug("ContractBroadcast", "event", event.Tx.Hash())
	//peers := pm.peers.PeersWithoutUnit(event.Tx.TxHash)
	peers := pm.peers.GetPeers()
	for _, peer := range peers {
		peer.SendContractExeTransaction(event)
	}
}

func (pm *ProtocolManager) ContractSigBroadcast(event jury.ContractSigEvent) {
	log.Info("ContractSigBroadcast", "event", event.Tx.Hash())
	//peers := pm.peers.PeersWithoutUnit(event.Tx.TxHash)
	peers := pm.peers.GetPeers()
	for _, peer := range peers {
		peer.SendContractSigTransaction(event)
	}
}
