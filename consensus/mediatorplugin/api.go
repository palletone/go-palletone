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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core"
)

func (mp *MediatorPlugin) LocalMediators() []common.Address {
	addrs := make([]common.Address, 0)

	for add := range mp.mediators {
		if mp.dag.IsMediator(add) {
			addrs = append(addrs, mp.mediators[add].Address)
		}
	}

	return addrs
}

func (mp *MediatorPlugin) GetLocalActiveMediators() []common.Address {
	lams := make([]common.Address, 0)

	for add := range mp.mediators {
		if mp.dag.IsActiveMediator(add) {
			lams = append(lams, add)
		}
	}

	return lams
}

func (mp *MediatorPlugin) GetLocalPrecedingMediators() []common.Address {
	lams := make([]common.Address, 0)

	for add := range mp.mediators {
		if mp.dag.IsPrecedingMediator(add) {
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
	if _, ok := mp.mediators[add]; ok {
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

func (mp *MediatorPlugin) localMediatorPubKey(add common.Address) []byte {
	var pubKey []byte = nil
	mp.dkgLock.Lock()
	defer mp.dkgLock.Unlock()

	dkgr, ok := mp.activeDKGs[add]
	if !ok || dkgr == nil {
		log.Debugf("the mediator(%v)'s dkg is not existed, or it is not active", add.String())
		return pubKey
	}

	dks, err := dkgr.DistKeyShare()
	if err == nil {
		pubKey, err = dks.Public().MarshalBinary()
		if err != nil {
			log.Debugf(err.Error())
			pubKey = nil
		}
	} else {
		log.Debugf(err.Error())
	}

	return pubKey
}

type PublicMediatorAPI struct {
	*MediatorPlugin
}

func NewPublicMediatorAPI(mp *MediatorPlugin) *PublicMediatorAPI {
	return &PublicMediatorAPI{mp}
}

// 初始群签名公私钥
type InitDKSResult struct {
	PrivateKey string `json:"private_key"` // 初始群签名私钥
	PublicKey  string `json:"public_key"`  // 初始群签名公钥
}

func (a *PublicMediatorAPI) DumpInitDKS() (res InitDKSResult) {
	sec, pub := core.GenInitPair()

	res.PrivateKey = core.ScalarToStr(sec)
	res.PublicKey = core.PointToStr(pub)

	return
}

type PrivateMediatorAPI struct {
	*MediatorPlugin
}

func NewPrivateMediatorAPI(mp *MediatorPlugin) *PrivateMediatorAPI {
	return &PrivateMediatorAPI{mp}
}

func (a *PrivateMediatorAPI) StartProduce() bool {
	if !a.producingEnabled {
		a.producingEnabled = true
		go a.launchProduction()

		return true
	}

	return false
}

func (a *PrivateMediatorAPI) StopProduce() bool {
	if a.producingEnabled {
		a.producingEnabled = false
		go func() {
			a.stopProduce <- struct{}{}
		}()

		return true
	}

	return false
}
