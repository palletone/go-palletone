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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package mediatorplugin

import (
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

func (mp *MediatorPlugin) LocalMediators() []common.Address {
	addrs := make([]common.Address, 0)
	for add, _ := range mp.mediators {
		addrs = append(addrs, mp.mediators[add].Address)
	}
	return addrs
}

func (mp *MediatorPlugin) IsEnabledGroupSign() bool {
	return mp.groupSigningEnabled
}

func (mp *MediatorPlugin) GetLocalActiveMediators() []common.Address {
	lams := make([]common.Address, 0)

	dag := mp.dag
	for add := range mp.mediators {
		if dag.IsActiveMediator(add) {
			lams = append(lams, add)
		}
	}

	return lams
}

func (mp *MediatorPlugin) LocalHaveActiveMediator() bool {
	dag := mp.dag
	for add := range mp.mediators {
		if dag.IsActiveMediator(add) {
			return true
		}
	}

	return false
}

func (mp *MediatorPlugin) IsLocalActiveMediator(add common.Address) bool {
	if mp.isLocalMediator(add) {
		return mp.dag.IsActiveMediator(add)
	}

	return false
}

func (mp *MediatorPlugin) LocalHavePrecedingMediator() bool {
	dag := mp.dag
	for add := range mp.mediators {
		if dag.IsPrecedingMediator(add) {
			return true
		}
	}

	return false
}

func (mp *MediatorPlugin) LocalMediatorPubKey(add common.Address) []byte {
	var pubKey []byte = nil
	dkgr, err := mp.getLocalActiveDKG(add)
	if err != nil {
		log.Debugf(err.Error())
		return pubKey
	}

	dks, err := dkgr.DistKeyShare()
	if err == nil {
		pubKey, err = dks.Public().MarshalBinary()
		if err != nil {
			pubKey = nil
		}
	}

	return pubKey
}

type PublicMediatorAPI struct {
	*MediatorPlugin
}

func NewPublicMediatorAPI(mp *MediatorPlugin) *PublicMediatorAPI {
	return &PublicMediatorAPI{mp}
}

func (a *PublicMediatorAPI) GetList() []string {
	addStrs := make([]string, 0)
	mas := a.dag.GetMediators()

	for address, _ := range mas {
		addStrs = append(addStrs, address.Str())
	}

	return addStrs
}

func (a *PublicMediatorAPI) ListVoteResult() map[string]uint64 {
	mediatorVoteCount := make(map[string]uint64)
	mas := a.dag.GetMediators()

	for address, _ := range mas {
		mediatorVoteCount[address.String()] = 0
	}
	accounts := a.dag.LookupAccount()
	for _, info := range accounts {
		ma := info.VotedMediator.String()
		count, ok := mediatorVoteCount[ma]
		if ok {
			mediatorVoteCount[ma] = count + info.Balance
		}
	}
	return mediatorVoteCount
}

func (a *PublicMediatorAPI) GetActives() []string {
	addStrs := make([]string, 0)
	ms := a.dag.ActiveMediators()

	for medAdd, _ := range ms {
		addStrs = append(addStrs, medAdd.Str())
	}

	return addStrs
}

type InitDKSResult struct {
	PrivateKey string
	PublicKey  string
}

func (a *PublicMediatorAPI) DumpInitDKS() (res InitDKSResult) {
	sec, pub := core.GenInitPair()

	res.PrivateKey = core.ScalarToStr(sec)
	res.PublicKey = core.PointToStr(pub)

	return
}

func (a *PublicMediatorAPI) GetVoted(addStr string) ([]string, error) {
	addr, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, err
	}

	voted := a.dag.GetAccountVotedMediator(addr)
	return []string{voted.String()}, nil

	//medMap := a.dag.GetAccountInfo(addr).VotedMediator
	//return []string{medMap.String()}, nil
	//mediators := make([]string, 0, len(medMap))
	//
	//for med, _ := range medMap {
	//	mediators = append(mediators, med.Str())
	//}
	//
	//return mediators, nil
}

//func (a *PublicMediatorAPI) GetDesiredCount(addStr string) (uint8, error) {
//	addr, err := common.StringToAddress(addStr)
//	if err != nil {
//		return 0, err
//	}
//
//	desiredCount := a.dag.GetAccountInfo(addr).DesiredMediatorCount
//	return desiredCount, nil
//}

func (a *PublicMediatorAPI) GetNextUpdateTime() string {
	dgp := a.dag.GetDynGlobalProp()
	time := time.Unix(int64(dgp.NextMaintenanceTime), 0)

	return time.Format("2006-01-02 15:04:05")
}

func (a *PublicMediatorAPI) GetInfo(addStr string) (*modules.MediatorInfo, error) {
	mediator, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, err
	}

	if !a.dag.IsMediator(mediator) {
		return nil, fmt.Errorf("%v is not mediator", mediator.Str())
	}

	return a.dag.GetMediatorInfo(mediator), nil
}
