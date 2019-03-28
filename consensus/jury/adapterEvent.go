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

package jury

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/dag/errors"
	"time"
)

func checkValid(reqEvt *AdapterRequestEvent) bool {
	hash := crypto.Keccak256(reqEvt.data)
	sig := reqEvt.sig[:len(reqEvt.sig)-1] // remove recovery id
	return crypto.VerifySignature(reqEvt.pubkey, hash, sig)
}
func (p *Processor) processAdapterRequestEvent(reqEvt *AdapterRequestEvent) (result *AdapterEvent, err error) {
	log.Info("processAdapterRequestEvent")

	//todo
	isValid := checkValid(reqEvt)
	if isValid {

	}

	//todo

	//todo

	return nil, nil
}

func (p *Processor) processAdapterResultEvent(rstEvt *AdapterResultEvent) error {
	log.Info("processAdapterResultEvent")
	//todo

	return nil
}

func (p *Processor) AdapterFunRequest(reqId common.Hash, contractId common.Address, msgType uint32, content []byte) ([]byte, error) {
	if reqId == (common.Hash{}) {
		return nil, errors.New("AdapterFunRequest param is nil")
	}

	//
	account := p.getLocalAccount()
	if account == nil {
		return nil, errors.New("AdapterFunRequest no local account")
	}

	//
	hash := crypto.Keccak256(content)
	sig, err := p.ptn.GetKeyStore().SignHashWithPassphrase(accounts.Account{Address: account.Address}, account.Password, hash)
	if err != nil {
		return nil, errors.New("AdapterFunRequest SignHashWithPassphrase failed")
	}
	//
	pubKey, err := p.ptn.GetKeyStore().GetPublicKey(account.Address)
	if err != nil {
		return nil, errors.New("AdapterFunRequest GetPublicKey failed")
	}
	//
	reqEvt := &AdapterRequestEvent{
		contractId: contractId,
		data:       content,
		sig:        sig,
		pubkey:     pubKey,
	}
	//
	isValid := checkValid(reqEvt)
	if isValid {

	}

	go p.ptn.AdapterBroadcast(AdapterEvent{AType: AdapterEventType(msgType), Event: reqEvt})

	p.locker.Lock()
	p.mtx[reqId].adaChan = make(chan bool, 1)
	p.locker.Unlock()

	return nil, errors.New("AdapterFunRequest, fail")
}

func (p *Processor) AdapterFunResult(reqId common.Hash, contractId common.Address, msgType uint32, timeOut time.Duration) ([]byte, error) {
	if reqId == (common.Hash{}) {
		return nil, errors.New("AdapterFunRequest param is nil")
	}

	p.locker.Lock()
	p.mtx[reqId].adaChan = make(chan bool, 1)
	p.locker.Unlock()

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(timeOut)
		timeout <- true
	}()

	select {
	case <-p.mtx[reqId].adaChan:
		log.Debug("AdapterFunRequest,  Ok")
		return nil, nil
	case <-timeout:
		log.Debug("AdapterFunRequest, time out")
		return nil, errors.New("AdapterFunRequest, time out")
	}
	return nil, errors.New("AdapterFunRequest, fail")
}

func (p *Processor) ProcessAdapterEvent(event *AdapterEvent) (result *AdapterEvent, err error) {
	if event == nil {
		return nil, errors.New("ProcessAdapterEvent, event is nil")
	}
	log.Info("ProcessAdapterEvent", "event", event.AType)

	if event.AType == ADAPTER_EVENT_REQUEST {
		return p.processAdapterRequestEvent(event.Event.(*AdapterRequestEvent))
	} else if event.AType == ADAPTER_EVENT_RESULT {
		return nil, p.processAdapterResultEvent(event.Event.(*AdapterResultEvent))
	}

	return nil, errors.New("ProcessAdapterEvent, fail")
}
