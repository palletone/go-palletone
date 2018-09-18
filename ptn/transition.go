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
	"errors"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
)

//go pm.mediatorMonitor(maxPeers)
//Start MediatorNetwork
func (pm *ProtocolManager) startMediatorNetwork(srvr *p2p.Server, maxPeers int) error {
	peers := pm.dag.GetActiveMediatorNodes()
	if maxPeers < len(peers)+3 {
		log.Error("PalletOne start", "maxpeers", maxPeers, "mediator size", len(peers)+3) //3:nomediator
		return errors.New("maxpeers < mediator size")
	}

	for _, peer := range peers {
		srvr.AddPeer(peer)
	}

	log.Debug("PalletOne", "startMediatorNetwork mediators:", len(peers))
	return nil
}

/*
	1.add flag.This node whether or not mediator
	2.check connections.
		2.1 mediator node: The no mediator connections is maxpeers sub mediators
		2.2 no mediator node:unlimited
	3.all the mediators node is connectin.Notice the mediator plugin
*/
func (pm *ProtocolManager) StartMediatorMonitor(srvr *p2p.Server, maxPeers int) {
	if !pm.producer.LocalHaveActiveMediator() {
		log.Info("This node is not Mediator")
		return
	}
	log.Info("mediator transition")
	for {
		if err := pm.startMediatorNetwork(srvr, maxPeers); err != nil {
			return
		}
		//TODO must modify
		go pm.monitor()
		break
	}
}

func (pm *ProtocolManager) monitor() {

}
