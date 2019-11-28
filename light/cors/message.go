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
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

func (pm *ProtocolManager) CorsHeaderMsg(msg p2p.Msg, p *peer) error {
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	if pm.fetcher != nil {
		for _, header := range headers {
			log.Trace("CorsHeaderMsg message content", "assetid:", header.Number.AssetID,
				"index:", header.Number.Index)
			pm.fetcher.Enqueue(p, header)
		}
	}
	return nil
}

func (pm *ProtocolManager) CorsHeadersMsg(msg p2p.Msg, p *peer) error {
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	log.Debug("CorsHeadersMsg message length", "len(headers)", len(headers))
	if pm.fetcher != nil {
		for _, header := range headers {
			pm.fetcher.Enqueue(p, header)
		}
		if len(headers) < MaxHeaderFetch {
			pm.bdlock.Lock()
			log.Info("CorsHeadersMsg message needboradcast", "assetid", headers[len(headers)-1].Number.AssetID,
				"index", headers[len(headers)-1].Number.Index)
			pm.needboradcast[p.id] = headers[len(headers)-1].Number.Index
			pm.bdlock.Unlock()

			ps := pm.peers.AllPeers()
			for _, p := range ps {
				p.SetHead(headers[len(headers)-1])
			}
		}
	}
	return nil
}
func (pm *ProtocolManager) GetCurrentHeaderMsg(msg p2p.Msg, p *peer) error {
	var number modules.ChainIndex
	if err := msg.Decode(&number); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	header := pm.dag.CurrentHeader(number.AssetID)
	log.Trace("GetCurrentHeaderMsg message content", "number", number.AssetID, "header", header)
	var headers []*modules.Header
	headers = append(headers, header)
	return p.SendCurrentHeader(headers)
}

func (pm *ProtocolManager) CurrentHeaderMsg(msg p2p.Msg, p *peer) error {
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	log.Trace("CurrentHeaderMsg message content", "len(headers)", len(headers))
	if len(headers) != 1 {
		log.Info("CurrentHeaderMsg len err", "len(headers)", len(headers))
		return errResp(ErrDecode, "msg %v: %v", msg, "len is err")
	}
	if headers[0].Number.AssetID.String() != pm.assetId.String() {
		log.Info("CurrentHeaderMsg", "assetid not equal response", headers[0].Number.AssetID.String(),
			"local", pm.assetId.String())
		return errBadPeer
	}
	pm.headerCh <- &headerPack{p.id, headers}
	return nil
}

func (pm *ProtocolManager) GetBlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// Decode the complex header query
	log.Debug("===Enter Light GetBlockHeadersMsg===")
	defer log.Debug("===End Ligth GetBlockHeadersMsg===")

	access := false
	if pcs, err := pm.dag.GetPartitionChains(); err == nil {
		for _, pc := range pcs {
			for _, pr := range pc.Peers {
				if node, err := discover.ParseNode(pr); err == nil {
					log.Debug("Cors ProtocolManager GetBlockHeadersMsg", "node.ID.String()", node.ID.String())
					if node.ID.TerminalString() == p.id {
						access = true
					}
				}
			}
		}
	}
	if !access {
		log.Error("Cors ProtocolManager GetBlockHeadersMsg do not access", "p.id", p.id)
		return errResp(ErrRequestRejected, "%v: %v", msg, "forbidden access")
	}

	var query getBlockHeadersData
	if err := msg.Decode(&query); err != nil {
		log.Info("GetBlockHeadersMsg Decode", "err:", err, "msg:", msg)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg getBlockHeadersData:", query)

	hashMode := query.Origin.Hash != (common.Hash{})
	log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg hashMode:", hashMode)
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
			log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg query.Origin.Number:", query.Origin.Number.Index)
			origin, _ = pm.dag.GetHeaderByNumber(&query.Origin.Number)
		}

		if origin == nil {
			break
		}
		log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg origin index:", origin.Number.Index)

		number := origin.Number.Index
		headers = append(headers, origin)
		bytes += estHeaderRlpSize

		// Advance to the next header of the query
		switch {
		case hashMode && query.Reverse:
			// Hash based traversal towards the genesis block
			log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the genesis block")
			for i := 0; i < int(query.Skip)+1; i++ {
				if header, err := pm.dag.GetHeaderByHash(query.Origin.Hash); err == nil && header != nil {
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
			log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the leaf block")
			var (
				currentIndex = origin.Number.Index
				nextIndex    = currentIndex + query.Skip + 1
				number       = origin.Number
			)

			log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg next", nextIndex, "current:", currentIndex)

			if nextIndex <= currentIndex {
				unknown = true
			} else {
				number.Index = nextIndex
				if header, _ := pm.dag.GetHeaderByNumber(number); header != nil {
					hashs := pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)
					if len(hashs) > int(query.Skip) && (hashs[query.Skip] == query.Origin.Hash) {
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
			log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the genesis block")
			if query.Origin.Number.Index >= query.Skip+1 {
				query.Origin.Number.Index -= query.Skip + 1
			} else {
				log.Info("Cors ProtocolManager", "GetBlockHeadersMsg query.Reverse", "unknown is true")
				unknown = true
			}

		case !query.Reverse:
			// Number based traversal towards the leaf block
			log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the leaf block")
			query.Origin.Number.Index += query.Skip + 1
		}
	}

	number := len(headers)
	if number > 0 {
		log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg query.Amount", query.Amount, "send number:", number,
			"start:", headers[0].Number.Index, "end:", headers[number-1].Number.Index,
			" getBlockHeadersData:", query)
	} else {
		log.Debug("Cors ProtocolManager", "GetBlockHeadersMsg query.Amount", query.Amount, "send number:", 0,
			" getBlockHeadersData:", query)
	}

	return p.SendUnitHeaders(headers)
}

func (pm *ProtocolManager) BlockHeadersMsg(msg p2p.Msg, p *peer) error {
	if pm.downloader == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received block header response message")
	// A batch of headers arrived to one of our previous requests
	var headers []*modules.Header

	if err := msg.Decode(&headers); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	err := pm.downloader.DeliverHeaders(p.id, headers)
	if err != nil {
		log.Debug(fmt.Sprint(err))
	}
	//p.fcServer.GotReply(resp.ReqID, resp.BV)
	//if pm.fetcher != nil && pm.fetcher.requestedID(resp.ReqID) {
	//	pm.fetcher.deliverHeaders(p, resp.ReqID, resp.Headers)
	//} else {
	//	err := pm.downloader.DeliverHeaders(p.id, resp.Headers)
	//	if err != nil {
	//		log.Debug(fmt.Sprint(err))
	//	}
	//}
	return nil
}
