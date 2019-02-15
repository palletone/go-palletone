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
	alg "github.com/palletone/go-palletone/consensus/jury/algorithm"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common"
	"crypto/rand"
	"github.com/palletone/go-palletone/common/log"
)

type elector struct {
	num    int
	weight uint64
	total  uint64

	privkey *alg.PrivateKey
	pubkey  *alg.PublicKey
}

func (p *Processor) ElectionRequest(reqId common.Hash) error {
	return nil //todo

	if reqId == (common.Hash{}) {
		return errors.New("ElectionRequest param is nil")
	}
	//seedData= reqId + rand
	rd := make([]byte, 20)
	_, err := rand.Read(rd)
	if err != nil {
		return errors.New("ElectionRequest rand fail")
	}
	seedData := make([]byte, len(reqId)+len(rd))
	copy(seedData, reqId[:])
	copy(seedData[len(reqId):], rd)

	reqEvent := ElectionRequestEvent{
		reqHash: reqId,
		num:     4, //todo
		data:    seedData,
	}
	log.Debug("ElectionRequest", "reqId", reqId.String(), "seedData", seedData)

	go p.ptn.ElectionBroadcast(ElectionEvent{EType: ELECTION_EVENT_REQUEST, Event: reqEvent})
	return nil
}

func (e *elector) checkElected(data []byte) (proof []byte, err error) {
	if e.num < 0 || e.weight < 10 || data == nil {
		return nil, errors.New("CheckElected param error")
	}
	proofHash, proof, err := e.privkey.Evaluate(data)
	if err != nil {
		return nil, err
	}
	selected := alg.Selected(e.num, e.weight, uint64(e.total), proofHash)
	if selected > 0 {
		return proof, nil
	} else {
		return nil, nil
	}
}

func (e *elector) verifyVRF(proofHash, proof, data []byte) (bool, error) {
	ok, err := e.pubkey.VerifyVRF(proof, data)
	if err != nil {
		return false, err
	}
	if ok {
		selected := alg.Selected(e.num, e.weight, e.total, proofHash)
		if selected > 0 {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}

func (p *Processor) ProcessElectionRequestEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回

	//pi, err := vrf.ECVRF_prove(pk, sk, msg[:])
	//if err != nil {
	//
	//}
	//
	//algorithm.Selected()

	return nil, nil
}

func (p *Processor) ProcessElectionResultEvent(event *ElectionEvent) error {

	return nil
}
