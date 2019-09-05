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
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/consensus/jury"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/ptn/downloader"
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

	hashMode := query.Origin.Hash != (common.Hash{})
	log.Debug("ProtocolManager", "GetBlockHeadersMsg getBlockHeadersData:", query,
		"GetBlockHeadersMsg hashMode:", hashMode)
	// Gather headers until the fetch or network limits is reached
	var (
		bytes   common.StorageSize
		headers []*modules.Header
		unknown bool
	)

	for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit &&
		len(headers) < downloader.MaxHeaderFetch {
		// Retrieve the next header satisfying the query
		var origin *modules.Header
		if hashMode {
			origin, _ = pm.dag.GetHeaderByHash(query.Origin.Hash)
		} else {
			log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Origin.Number:", query.Origin.Number.Index)
			origin, _ = pm.dag.GetHeaderByNumber(&query.Origin.Number)
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
			//log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the leaf block")
			var (
				current = origin.Number.Index
				next    = current + query.Skip + 1
				index   = origin.Number
			)
			//log.Debug("ProtocolManager", "GetBlockHeadersMsg next", next, "current:", current)
			if next <= current {
				infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
				log.Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip,
					"next", next, "attacker", infos)
				unknown = true
			} else {
				index.Index = next
				log.Debug("ProtocolManager", "GetBlockHeadersMsg index.Index:", index.Index)
				if header, _ := pm.dag.GetHeaderByNumber(index); header != nil {
					hashs := pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)
					log.Debug("ProtocolManager", "GetUnitHashesFromHash len(hashs):", len(hashs),
						"header.index:", header.Number.Index, "header.hash:", header.Hash().String(),
						"query.Skip+1", query.Skip+1)
					if len(hashs) > int(query.Skip) && (hashs[query.Skip] == query.Origin.Hash) {
						query.Origin.Hash = header.Hash()
					} else {
						log.Debug("ProtocolManager GetBlockHeadersMsg unknown = true;" +
							" pm.dag.GetUnitHashesFromHash not equal origin hash.")
						log.Debug("ProtocolManager", "GetBlockHeadersMsg header.Hash()", header.Hash(),
							"query.Skip+1:", query.Skip+1, "query.Origin.Hash:", query.Origin.Hash)
						unknown = true
					}
				} else {
					log.Debug("ProtocolManager", "GetBlockHeadersMsg unknown = true; pm.dag.GetHeaderByNumber"+
						" not found. Index:", index.Index)
					unknown = true
				}
			}
		case query.Reverse:
			// Number based traversal towards the genesis block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the genesis block")
			if query.Origin.Number.Index >= query.Skip+1 {
				query.Origin.Number.Index -= query.Skip + 1
			} else {
				log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Reverse", "unknown is true")
				unknown = true
			}

		case !query.Reverse:
			// Number based traversal towards the leaf block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the leaf block")
			query.Origin.Number.Index += query.Skip + 1
		}
	}

	number := len(headers)
	if number > 0 {
		log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Amount", query.Amount, "send number:", number,
			"start:", headers[0].Number.Index, "end:", headers[number-1].Number.Index, " getBlockHeadersData:", query)
	} else {
		log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Amount", query.Amount, "send number:", 0,
			" getBlockHeadersData:", query)
	}

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
		//log.Debug("===BlockHeadersMsg ===", "len(headers):", len(headers))
		err := pm.downloader.DeliverHeaders(p.id, headers)
		if err != nil {
			log.Debug("Failed to deliver headers", "err", err.Error())
		}
	}
	return nil
}

func (pm *ProtocolManager) GetBlockBodiesMsg(msg p2p.Msg, p *peer) error {
	// Decode the retrieval message
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
		//log.Debug("GetBlockBodiesMsg", "hash", hash)
		// Retrieve the requested block body, stopping if enough was found
		txs, err := pm.dag.GetUnitTransactions(hash)
		if err != nil || len(txs) == 0 {
			log.Debug("GetBlockBodiesMsg", "hash:", hash, "GetUnitTransactions err:", err)
			//return errResp(ErrDecode, "msg %v: %v", msg, err)
			continue
		}

		data, err := rlp.EncodeToBytes(txs)
		if err != nil {
			log.Debug("Get body Marshal encode", "error", err.Error(), "unit hash", hash.String())
			//return errResp(ErrDecode, "msg %v: %v", msg, err)
			continue
		}
		//log.Debug("Get body Marshal", "data:", string(data))
		bytes += len(data)
		bodies = append(bodies, data)
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

	transactions := make([][]*modules.Transaction, len(request))
	for i, body := range request {
		if len(body) == 0 {
			continue
		}
		var txs modules.Transactions
		if err := rlp.DecodeBytes(body, &txs); err != nil {
			log.Debug("have body rlp decode", "error", err.Error(), "body", string(body))
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}

		transactions[i] = txs
		log.Debug("BlockBodiesMsg", "i", i, "txs size:", len(txs))
	}
	log.Debug("Full ProtocolManager BlockBodiesMsg", "len(transactions:)", len(transactions))
	// Filter out any explicitly requested bodies, deliver the rest to the downloader
	filter := len(transactions) > 0
	if filter {
		transactions = pm.fetcher.FilterBodies(p.id, transactions, time.Now())
	}
	if len(transactions) > 0 || !filter {
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
	log.Debug("Enter Full ProtocolManager NewBlockHashesMsg")
	defer log.Debug("End Full ProtocolManager NewBlockHashesMsg")
	var announces newBlockHashesData
	if err := msg.Decode(&announces); err != nil {
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
	log.Debug("Full ProtocolManager NewBlockHashesMsg", "len(unknown):", len(unknown))
	for _, block := range unknown {
		pm.fetcher.Notify(p.id, block.Hash, &block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
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
	if err := rlp.DecodeBytes(data, &unit); err != nil {
		log.Info("ProtocolManager", "NewBlockMsg rlp decode err:", err, "data", string(data))
		return err
	}

	if unit.IsEmpty() {
		log.Errorf("unit is nil/empty")
		return nil
	}

	// append by Albert·Gou
	timestamp := time.Unix(unit.Timestamp(), 0)
	//latency := time.Now().Sub(timestamp)
	latency := time.Since(timestamp)
	if latency < -5*time.Second {
		errStr := fmt.Sprintf("Rejecting unit #%v with timestamp(%v) in the future signed by %v",
			unit.NumberU64(), timestamp.Format("2006-01-02 15:04:05"), unit.Author().Str())
		log.Debugf(errStr)
		return fmt.Errorf(errStr)
	}

	unitHash := unit.Hash()
	if pm.IsExistInCache(unitHash.Bytes()) {
		//log.Debugf("Received unit(%v) again, ignore it", unitHash.TerminalString())
		p.MarkUnit(unitHash)
		p.SetHead(unitHash, unit.Number(), nil)
		return nil
	}

	log.Infof("Received unit(%v) #%v parent(%v) @%v signed by %v", unitHash.TerminalString(),
		unit.NumberU64(), unit.ParentHash()[0].TerminalString(), timestamp.Format("2006-01-02 15:04:05"),
		unit.Author().Str())

	log.DebugDynamic(func() string {
		txids := []common.Hash{}
		for _, tx := range unit.Txs {
			txids = append(txids, tx.Hash())
		}
		return fmt.Sprintf("NewBlockMsg, received unit hash %s, txs:[%x]", unit.UnitHash.String(), txids)
	})

	rwset.Init()
	var temptxs modules.Transactions
	index := 0
	for i, tx := range unit.Txs {
		if i == 0 {
			temptxs = append(temptxs, tx)
			continue //coinbase
		}
		if tx.IsContractTx() {
			reqId := tx.RequestHash()
			log.Debugf("[%s]NewBlockMsg, index[%x],txHash[%s]", reqId.String()[0:8], index, tx.Hash().String())
			index++
			if !pm.contractProc.CheckContractTxValid(rwset.RwM, tx, true) {
				log.Debugf("[%s]NewBlockMsg, CheckContractTxValid is false.", reqId.String()[0:8])
				continue
			}
		}
		temptxs = append(temptxs, tx)
	}
	rwset.RwM.Close()
	unit.Txs = temptxs

	unit.ReceivedAt = msg.ReceivedAt
	unit.ReceivedFrom = p

	// Mark the peer as owning the block and schedule it for import
	p.MarkUnit(unit.UnitHash)
	pm.fetcher.Enqueue(p.id, unit)

	requestNumber := unit.Number()
	hash, number := p.Head(unit.Number().AssetID)
	if common.EmptyHash(hash) || (!common.EmptyHash(hash) && requestNumber.Index > number.Index) {
		log.Debug("ProtocolManager", "NewBlockMsg SetHead request.Index:", requestNumber.Index,
			"local peer index:", number.Index)
		p.SetHead(unit.Hash(), requestNumber, nil)

		//currentUnitIndex := pm.dag.GetCurrentUnit(unit.Number().AssetID).UnitHeader.Number.Index
		currentUnitIndex := pm.dag.HeadUnitNum()
		if requestNumber.Index > currentUnitIndex+1 {
			log.Debug("ProtocolManager", "NewBlockMsg synchronize request.Index:", requestNumber.Index,
				"current unit index+1:", currentUnitIndex+1)
			go func() {
				pm.synchronize(p, unit.Number().AssetID, nil)
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
	for i, tx := range txs {
		// Validate and mark the remote transaction
		if tx == nil {
			return errResp(ErrDecode, "transaction %d is nil", i)
		}
		txHash := tx.Hash()
		if pm.IsExistInCache(txHash.Bytes()) {
			return nil
		}
		if tx.IsContractTx() {
			if pm.contractProc.IsSystemContractTx(tx) {
				continue
			}
		}
		p.MarkTransaction(tx.Hash())
		_, err := pm.txpool.ProcessTransaction(tx, true, true, 0 /*pm.txpool.Tag(peer.ID())*/)
		if err != nil {
			log.Infof("the transaction %s not accepteable, err:%s", tx.Hash().String(), err.Error())
		}
	}

	return nil
}

func (pm *ProtocolManager) SigShareMsg(msg p2p.Msg, p *peer) error {
	var sigShare mp.SigShareEvent
	if err := msg.Decode(&sigShare); err != nil {
		errStr := fmt.Sprintf("SigShareMsg: %v, err: %v", msg, err)
		log.Debugf(errStr)

		//return fmt.Errorf(errStr)
		return nil
	}

	pm.producer.AddToTBLSRecoverBuf(sigShare.UnitHash, sigShare.SigShare)
	return nil
}

func (pm *ProtocolManager) VSSDealMsg(msg p2p.Msg, p *peer) error {
	var deal mp.VSSDealEvent
	if err := msg.Decode(&deal); err != nil {
		errStr := fmt.Sprintf("VSSDealMsg: %v, err: %v", msg, err)
		log.Debugf(errStr)

		//return fmt.Errorf(errStr)
		return nil
	}

	// 判断是否同步, 如果没同步完成，接收到的vss deal是超前的
	if !pm.dag.IsSynced() {
		errStr := "we are not synced"
		log.Debugf(errStr)
		//return fmt.Errorf(errStr)
		return nil
	}

	// todo albert 清除在限制时间范围之外的deal消息

	go pm.producer.AddToDealBuf(&deal)
	return nil
}

func (pm *ProtocolManager) VSSResponseMsg(msg p2p.Msg, p *peer) error {
	var resp mp.VSSResponseEvent
	if err := msg.Decode(&resp); err != nil {
		errStr := fmt.Sprintf("VSSResponseMsg: %v, err: %v", msg, err)
		log.Debugf(errStr)

		//return fmt.Errorf(errStr)
		return nil
	}

	// todo albert 清除在限制时间范围之外的response消息

	go pm.producer.AddToResponseBuf(&resp)
	return nil
}

//GroupSigMsg
func (pm *ProtocolManager) GroupSigMsg(msg p2p.Msg, p *peer) error {
	var gSign mp.GroupSigEvent
	if err := msg.Decode(&gSign); err != nil {
		errStr := fmt.Sprintf("GroupSigMsg: %v, err: %v", msg, err)
		log.Debugf(errStr)

		//return fmt.Errorf(errStr)
		return nil
	}

	go pm.dag.SetUnitGroupSign(gSign.UnitHash, gSign.GroupSig, pm.txpool)
	return nil
}

func (pm *ProtocolManager) ContractMsg(msg p2p.Msg, p *peer) error {
	var event jury.ContractEvent
	if err := msg.Decode(&event); err != nil {
		log.Info("ProtocolManager ContractMsg", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	if pm.IsExistInCache(event.Hash().Bytes()) {
		//log.Debugf("Received event(%v) again, ignore it", event.Hash().String())
		return nil
	}

	reqId := event.Tx.RequestHash()
	log.Debugf("[%s] ProtocolManager ContractMsg, event type[%v]", reqId.String()[0:8], event.CType)
	err := pm.contractProc.ProcessContractEvent(&event)
	if err != nil {
		log.Debugf("[%s]ProtocolManager ContractMsg, error:%s", reqId.String()[0:8], err.Error())
	}
	return nil
}

func (pm *ProtocolManager) ElectionMsg(msg p2p.Msg, p *peer) error {
	var evs jury.ElectionEventBytes
	if err := msg.Decode(&evs); err != nil {
		log.Info("===ElectionMsg===", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	if pm.IsExistInCache(evs.Hash().Bytes()) {
		return nil
	}
	event, err := evs.ToElectionEvent()
	if err != nil {
		log.Debug("ElectionMsg, ToElectionEvent fail")
		return nil
	}
	_, err = pm.contractProc.ProcessElectionEvent(event)
	if err != nil {
		log.Debug("ElectionMsg", "ProcessElectionEvent error:", err)
	}

	return nil
}

func (pm *ProtocolManager) AdapterMsg(msg p2p.Msg, p *peer) error {
	var avs jury.AdapterEventBytes
	if err := msg.Decode(&avs); err != nil {
		log.Info("ProtocolManager AdapterMsg", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	event, err := avs.ToAdapterEvent()
	if err != nil {
		log.Debug("ProtocolManager AdapterMsg, ToAdapterEvent fail")
		return nil
	}

	_, err = pm.contractProc.ProcessAdapterEvent(event)
	if err != nil {
		log.Debug("ProtocolManager AdapterMsg", "ProcessAdapterEvent error:", err)
	}
	return nil
}

func (pm *ProtocolManager) GetLeafNodesMsg(msg p2p.Msg, p *peer) error {
	headers, err := pm.dag.GetAllLeafNodes()
	if err != nil {
		log.Info("Full ProtocolManager GetLeafNodesMsg", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	return p.SendUnitHeaders(headers)
}
