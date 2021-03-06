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
	"github.com/palletone/go-palletone/ptn/downloader"
)

type Tag uint64

const errStr = "this node is not synced"

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
		log.Debug("ProtocolManager", "GetBlockHeadersMsg origin index:", origin.GetNumber().Index)

		number := origin.GetNumber().Index
		headers = append(headers, origin)
		bytes += estHeaderRlpSize

		// Advance to the next header of the query
		switch {
		case hashMode && query.Reverse:
			// Hash based traversal towards the genesis block
			for i := 0; i < int(query.Skip)+1; i++ {
				if header, err := pm.dag.GetHeaderByHash(query.Origin.Hash); err == nil && header != nil {
					if number != 0 {
						query.Origin.Hash = header.ParentHash()[0]
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
				current = origin.GetNumber().Index
				next    = current + query.Skip + 1
				index   = origin.GetNumber()
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
						"header.index:", header.GetNumber().Index, "header.hash:", header.Hash().String(),
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
			"start:", headers[0].GetNumber().Index, "end:", headers[number-1].GetNumber().Index, " getBlockHeadersData:", query)
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

//func includeContractTx(txs []*modules.Transaction) bool { //todo  sort
//	for _, tx := range txs {
//		if tx.IsContractTx() && tx.IsOnlyContractRequest() {
//			return true
//		}
//	}
//	return false
//}
func (pm *ProtocolManager) TxMsg(msg p2p.Msg, p *peer) error {
	log.Debug("Enter TxMsg")
	defer log.Debug("End TxMsg")
	// Transactions arrived, make sure we have a valid and fresh chain to handle them
	if atomic.LoadUint32(&pm.acceptTxs) == 0 {
		log.Debug("TxMsg pm.acceptTxs==0")
		return nil
	}
	// Transactions can be processed, parse all of them and deliver to the pool
	var txs []*modules.Transaction
	if err := msg.Decode(&txs); err != nil {
		log.Debug("TxMsg", "Decode err:", err, "msg:", msg)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Debugf("TxMsg, tx num[%d]", len(txs))

	for i, tx := range txs {
		if tx == nil {
			return errResp(ErrDecode, "TxMsg transaction %d is nil", i)
		}
		txHash := tx.Hash()
		reqId := tx.RequestHash()
		p.MarkTransaction(txHash)
		if pm.IsExistInCache(txHash.Bytes()) {
			return nil
		}

		log.Debugf("[%s]TxMsg, index[%d] txHash[%s] tx:%s", reqId.ShortStr(), i, txHash.ShortStr(), tx.String())
		if tx.IsContractTx() && tx.GetContractTxType() == modules.APP_CONTRACT_INVOKE_REQUEST {
			//系统合约的请求可以P2P广播，但是包含结果的系统合约请求，只能在打包时生成，不能广播
			if tx.IsSystemContract() {
				if !tx.IsOnlyContractRequest() {
					log.Debugf("[%s]TxMsg, Tx[%s] is a sys contract with result, don't need send by p2p",
						reqId.ShortStr(), txHash.String())
					continue
				}
			}
		}

		//添加到本地交易
		err := pm.contractProc.AddLocalTx(tx)
		if err != nil {
			log.Warnf("[%s]TxMsg, AddLocalTx[%s] err:%s", reqId.ShortStr(), txHash.String(), err.Error())
		}
		err = pm.txpool.AddRemote(tx)
		if err != nil {
			log.Infof("[%s]TxMsg,the transaction %s not accepteable, err:%s", reqId.ShortStr(), txHash.String(), err.Error())
		}
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

	unitHash := unit.Hash()
	p.MarkUnit(unitHash)
	p.SetHead(unitHash, unit.Number(), nil)

	if pm.IsExistInCache(unitHash.Bytes()) {
		//log.Debugf("Received unit(%v) again, ignore it", unitHash.TerminalString())
		return nil
	}

	timestamp := time.Unix(unit.Timestamp(), 0)
	log.Infof("Received unit(%v) #%v parent(%v) @%v signed by %v", unitHash.TerminalString(),
		unit.NumberU64(), unit.ParentHash()[0].TerminalString(), timestamp.Format("2006-01-02 15:04:05"),
		unit.Author().Str())

	// append by Albert·Gou
	latency := time.Since(timestamp)
	if latency < -5*time.Second {
		errStr := fmt.Sprintf("Rejecting unit #%v with timestamp(%v) in the future signed by %v",
			unit.NumberU64(), timestamp.Format("2006-01-02 15:04:05"), unit.Author().Str())
		log.Debugf(errStr)
		return fmt.Errorf(errStr)
	}

	log.DebugDynamic(func() string {
		txids := []common.Hash{}
		for _, tx := range unit.Txs {
			txids = append(txids, tx.Hash())
		}
		return fmt.Sprintf("NewBlockMsg, received unit hash %s, txs:[%x]", unit.Hash().String(), txids)
	})

	unit.ReceivedAt = msg.ReceivedAt
	unit.ReceivedFrom = p

	// Mark the peer as owning the block and schedule it for import
	//p.MarkUnit(unit.UnitHash)
	pm.fetcher.Enqueue(p.id, unit)

	requestNumber := unit.Number()
	hash, number := p.Head(requestNumber.AssetID)
	if common.EmptyHash(hash) || (!common.EmptyHash(hash) && requestNumber.Index > number.Index) {
		log.Debug("ProtocolManager", "NewBlockMsg SetHead request.Index:", requestNumber.Index,
			"local peer index:", number.Index)
		//p.SetHead(unit.Hash(), requestNumber, nil)

		//currentUnitIndex := pm.dag.GetCurrentUnit(unit.Number().AssetID).UnitHeader.Number.Index
		currentUnitIndex := pm.dag.HeadUnitNum()
		if requestNumber.Index > currentUnitIndex+1 {
			log.Debug("ProtocolManager", "NewBlockMsg synchronize request.Index:", requestNumber.Index,
				"current unit index+1:", currentUnitIndex+1)
			go func() {
				pm.synchronize(p, requestNumber.AssetID, nil)
			}()
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

	hash := sigShare.Hash()
	p.MarkSigShare(hash)

	if pm.IsExistInCache(hash.Bytes()) {
		return nil
	}

	// 判断是否同步, 如果没同步完成，接收到的 sigShare 对当前节点来说是超前的
	if !pm.dag.IsSynced(false) {
		log.Debugf(errStr)
		go pm.BroadcastSigShare(&sigShare)

		//return fmt.Errorf(errStr)
		return nil
	}

	unitHash := sigShare.UnitHash
	// 或者由于网络延迟，该单元在收到群签名之前，已经根据深度转为不可逆了
	isStable, _ := pm.dag.IsIrreversibleUnit(unitHash)
	if isStable {
		return nil
	}

	header, err := pm.dag.GetHeaderByHash(unitHash)
	if err != nil {
		log.Debugf("fail to get header of unit(%v), err: %v", unitHash.TerminalString(), err.Error())
		return nil
	}

	if pm.producer.IsLocalMediator(header.Author()) {
		go pm.producer.AddToTBLSRecoverBuf(&sigShare, header)
	} else {
		go pm.BroadcastSigShare(&sigShare)
	}

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

	hash := gSign.Hash()
	p.MarkGroupSig(hash)

	if pm.IsExistInCache(hash.Bytes()) {
		return nil
	}

	// 或者由于网络延迟，该单元在收到群签名之前，已经根据深度转为不可逆了
	isStable, _ := pm.dag.IsIrreversibleUnit(gSign.UnitHash)
	if !isStable {
		pm.BroadcastGroupSig(&gSign)
		go pm.dag.SetUnitGroupSign(gSign.UnitHash, gSign.GroupSig)
	}

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

	hash := deal.Hash()
	p.MarkVSSDeal(hash)

	if pm.IsExistInCache(hash.Bytes()) {
		return nil
	}

	// 判断是否同步, 如果没同步完成，接收到的 vss deal 对当前节点来说是超前的
	if !pm.dag.IsSynced(true) {
		log.Debugf(errStr)
		go pm.BroadcastVSSDeal(&deal)

		//return fmt.Errorf(errStr)
		return nil
	}

	ma := pm.dag.GetActiveMediatorAddr(int(deal.DstIndex))
	if pm.producer.IsLocalMediator(ma) {
		go pm.producer.AddToDealBuf(&deal)
	} else {
		go pm.BroadcastVSSDeal(&deal)
	}

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

	hash := resp.Hash()
	p.MarkVSSResponse(hash)

	if pm.IsExistInCache(hash.Bytes()) {
		return nil
	}

	go pm.BroadcastVSSResponse(&resp)

	// 判断是否同步, 如果没同步完成，接收到的 vss response 对当前节点来说是超前的
	if !pm.dag.IsSynced(true) {
		log.Debugf(errStr)
		//return fmt.Errorf(errStr)
		return nil
	}

	go pm.producer.AddToResponseBuf(&resp)
	return nil
}

func (pm *ProtocolManager) ContractMsg(msg p2p.Msg, p *peer) error {
	var event jury.ContractEvent
	if err := msg.Decode(&event); err != nil {
		log.Info("ContractMsg", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	if pm.IsExistInCache(event.Hash().Bytes()) {
		//log.Debugf("Received event(%v) again, ignore it", event.Hash().String())
		return nil
	}

	// 判断是否同步, 如果没同步完成，接收到的 ContractMsg 对当前节点来说是超前的
	//if !pm.dag.IsSynced(false) {
	//	log.Debugf(errStr)
	//	//return fmt.Errorf(errStr)
	//	return nil
	//}

	reqId := event.Tx.RequestHash()
	log.Debugf("[%s] ContractMsg, event type[%v]", reqId.ShortStr(), event.CType)
	brd, err := pm.contractProc.ProcessContractEvent(&event)
	if err != nil {
		log.Debugf("[%s]ContractMsg, error:%s", reqId.ShortStr(), err.Error())
	}
	if brd && pm.peers != nil {
		peers := pm.peers.GetPeers()
		log.Debugf("[%s]ContractMsg, event type[%d], peers num[%d]", reqId.ShortStr(), event.CType, len(peers))
		for _, peer := range peers {
			if err := peer.SendContractTransaction(event); err != nil {
				log.Error("ContractMsg", "SendContractTransaction err:", err.Error())
			}
		}
	}
	return nil
}

func (pm *ProtocolManager) ElectionMsg(msg p2p.Msg, p *peer) error {
	var evs jury.ElectionEventBytes
	if err := msg.Decode(&evs); err != nil {
		log.Info("ElectionMsg", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}
	if pm.IsExistInCache(evs.Hash().Bytes()) {
		return nil
	}

	// 判断是否同步, 如果没同步完成，接收到的 ElectionMsg 对当前节点来说是超前的
	if !pm.dag.IsSynced(false) {
		log.Debugf(errStr)
		return nil
	}
	event, err := evs.ToElectionEvent()
	if err != nil {
		log.Errorf("ElectionMsg, ToElectionEvent fail, err:%s", err.Error())
		return nil
	}
	ReqId := event.ReqId()
	log.Debugf("[%s]ElectionMsg, event type[%v]", ReqId.ShortStr(), event.EType)
	err = pm.contractProc.ProcessElectionEvent(event)
	if err != nil {
		log.Warnf("[%s]ElectionMsg, ProcessElectionEvent error:%s", ReqId.ShortStr(), err)
	}
	if pm.peers != nil {
		peers := pm.peers.GetPeers()
		log.Debugf("[%s]ElectionMsg, event type[%d], peers num[%d]", ReqId.ShortStr(), event.EType, len(peers))
		for idx, peer := range peers {
			if err := peer.SendElectionEvent(*event); err != nil {
				log.Errorf("[%s]ElectionMsg, SendContractTransaction err[%d]:%s", ReqId.ShortStr(), idx, err.Error())
			}
		}
	}
	return nil
}

func (pm *ProtocolManager) AdapterMsg(msg p2p.Msg, p *peer) error {
	var avs jury.AdapterEventBytes
	if err := msg.Decode(&avs); err != nil {
		log.Info("ProtocolManager AdapterMsg", "err:", err)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	if pm.IsExistInCache(avs.Hash().Bytes()) {
		return nil
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
