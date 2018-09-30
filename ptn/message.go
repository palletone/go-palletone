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
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

func (pm *ProtocolManager) StatusMsg(msg p2p.Msg, p *peer) error {
	// Status messages should never arrive after the handshake
	return errResp(ErrExtraStatusMsg, "uncontrolled status message")
}

func (pm *ProtocolManager) GetBlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// Block header query, collect the requested headers and reply

	// Decode the complex header query
	var query getBlockHeadersData
	if err := msg.Decode(&query); err != nil {
		log.Info("GetBlockHeadersMsg Decode", "err:", err, "msg:", msg)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	hashMode := query.Origin.Hash != (common.Hash{})

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
			//origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
			origin = pm.dag.GetHeaderByHash(query.Origin.Hash)
		} else {
			//index *modules.ChainIndex
			origin = pm.dag.GetHeaderByNumber(query.Origin.Number)
		}
		if origin == nil {
			break
		}

		number := origin.Number.Index
		headers = append(headers, origin)
		bytes += estHeaderRlpSize

		// Advance to the next header of the query
		switch {
		case hashMode && query.Reverse:
			// Hash based traversal towards the genesis block
			for i := 0; i < int(query.Skip)+1; i++ {
				if header, err := pm.dag.GetHeader(query.Origin.Hash, number); err == nil && header != nil {
					if number != 0 {
						query.Origin.Hash = header.ParentsHash[0]
					}
					number--
				} else {
					log.Info("========GetBlockHeadersMsg========", "number", number, "err:", err)
					unknown = true
					break
				}
			}
		case hashMode && !query.Reverse:
			// Hash based traversal towards the leaf block
			var (
				current = origin.Number.Index
				next    = current + query.Skip + 1
				index   = origin.Number
			)
			if next <= current {
				infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
				log.Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
				unknown = true
			} else {
				index.Index = next
				if header := pm.dag.GetHeaderByNumber(index); header != nil {
					if pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip] == query.Origin.Hash {
						query.Origin.Hash = header.Hash()
					} else {
						unknown = true
					}
				} else {
					unknown = true
				}
			}
		case query.Reverse:
			// Number based traversal towards the genesis block
			if query.Origin.Number.Index >= query.Skip+1 {
				query.Origin.Number.Index -= query.Skip + 1
			} else {
				log.Info("========GetBlockHeadersMsg========", "query.Reverse", "unknown is true")
				unknown = true
			}

		case !query.Reverse:
			// Number based traversal towards the leaf block
			query.Origin.Number.Index += query.Skip + 1
		}
	}
	log.Debug("========GetBlockHeadersMsg========", "query.Amount", query.Amount, "send number:", len(headers))
	return p.SendUnitHeaders(headers)
}

func (pm *ProtocolManager) BlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// A batch of headers arrived to one of our previous requests
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// If no headers were received, but we're expending a DAO fork check, maybe it's that
	if len(headers) == 0 {
		log.Debug("===handler->msg.Code == BlockHeadersMsg len(headers)is 0===")
		return nil
	}
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
	//log.Debug("===GetBlockBodiesMsg===")
	msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if _, err := msgStream.List(); err != nil {
		return err
	}
	// Gather blocks until the fetch or network limits is reached
	var (
		hash  common.Hash
		bytes int
		//bodies []rlp.RawValue
		bodies blockBody
	)

	for bytes < softResponseLimit && len(bodies.Transactions) < downloader.MaxBlockFetch {
		// Retrieve the hash of the next block
		if err := msgStream.Decode(&hash); err == rlp.EOL {
			break
		} else if err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		//TODO must recover
		// Retrieve the requested block body, stopping if enough was found
		//GetTransactionsByUnitHash(hash)
		//TODO must modify
		//txs, err := pm.dag.GetTransactionByHash(hash)
		txs := []*modules.Transaction{}
		var err error
		if err != nil {
			log.Debug("===GetBlockBodiesMsg===", "GetTransactionByHash err:", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		log.Debug("===GetBlockBodiesMsg===", "GetTransactionByHash txs:", txs)
		data, err := rlp.EncodeToBytes(txs)
		if err != nil {
			log.Debug("Get body rlp when rlp encode", "unit hash", hash.String(), "error", err.Error())
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		bytes += len(data)

		for _, tx := range txs {
			bodies.Transactions = append(bodies.Transactions, tx)
		}
	}
	//log.Debug("===GetBlockBodiesMsg===", "tempGetBlockBodiesMsgSum:", tempGetBlockBodiesMsgSum, "sum:", sum)
	log.Debug("===GetBlockBodiesMsg===", "len(bodies):", len(bodies.Transactions), "bytes:", bytes)
	return p.SendBlockBodies([]*blockBody{&bodies})
}

func (pm *ProtocolManager) BlockBodiesMsg(msg p2p.Msg, p *peer) error {
	//log.Debug("===BlockBodiesMsg===")
	// A batch of block bodies arrived to one of our previous requests
	var request blockBodiesData
	if err := msg.Decode(&request); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Deliver them all to the downloader for queuing
	transactions := make([][]*modules.Transaction, len(request))
	sum := 0
	for i, body := range request {
		transactions[i] = body.Transactions
		sum++
	}

	log.Debug("===BlockBodiesMsg===", "len(transactions:)", len(transactions), "transactions[0]:", transactions[0])
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

	var unit modules.Unit
	if err := msg.Decode(&unit); err != nil {
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	unit.ReceivedAt = msg.ReceivedAt
	unit.ReceivedFrom = p
	log.Info("===NewBlockMsg===", "index:", unit.Number().Index)

	// Mark the peer as owning the block and schedule it for import
	p.MarkUnit(unit.UnitHash)
	pm.fetcher.Enqueue(p.id, &unit)

	hash, number := p.Head(unit.Number().AssetID)

	if common.EmptyHash(hash) || (!common.EmptyHash(hash) && unit.UnitHeader.ChainIndex().Index > number.Index) {
		trueHead := unit.Hash()
		log.Info("=================handler p.SetHead===============")
		p.SetHead(trueHead, unit.UnitHeader.ChainIndex())
		// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
		// a singe block (as the true TD is below the propagated block), however this
		// scenario should easily be covered by the fetcher.
		//如果在我们上面安排一个同步。注意，这将不会为单个块的间隙触发同步(因为真正的TD位于传播的块之下)，
		//但是这个场景应该很容易被fetcher所覆盖。
		currentUnit := pm.dag.CurrentUnit()
		if currentUnit != nil && unit.UnitHeader.ChainIndex().Index > currentUnit.UnitHeader.ChainIndex().Index {
			go pm.synchronise(p, unit.Number().AssetID)
		}
	}
	return nil
}

func (pm *ProtocolManager) TxMsg(msg p2p.Msg, p *peer) error {
	log.Info("===============ProtocolManager TxMsg====================")
	// Transactions arrived, make sure we have a valid and fresh chain to handle them
	if atomic.LoadUint32(&pm.acceptTxs) == 0 {
		log.Debug("ProtocolManager handlmsg TxMsg pm.acceptTxs==0")
		return nil
	}
	// Transactions can be processed, parse all of them and deliver to the pool
	var txs []*modules.Transaction
	if err := msg.Decode(&txs); err != nil {
		log.Info("ProtocolManager handlmsg TxMsg", "Decode err:", err, "msg:", msg)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	//TODO VerifyTX
	log.Info("===============ProtocolManager", "TxMsg txs:", txs)
	for i, tx := range txs {
		// Validate and mark the remote transaction
		if tx == nil {
			return errResp(ErrDecode, "transaction %d is nil", i)
		}
		p.MarkTransaction(tx.Hash())
	}
	log.Info("===============ProtocolManager TxMsg AddRemotes====================")
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

func (pm *ProtocolManager) NewUnitMsg(msg p2p.Msg, p *peer) error {
	// Retrieve and decode the propagated new produced unit
	var unit modules.Unit
	if err := msg.Decode(&unit); err != nil {
		log.Info("===NewUnitMsg===", "err:", err)
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
