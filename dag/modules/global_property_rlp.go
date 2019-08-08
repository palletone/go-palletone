/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package modules

import (
	"io"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
)

func (gp *GlobalProperty) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}

	gpt := &GlobalPropertyTemp{}
	err = rlp.DecodeBytes(raw, gpt)
	if err != nil {
		return err
	}

	gpt.getGP(gp)
	return nil
}

func (gp *GlobalProperty) EncodeRLP(w io.Writer) error {
	temp := gp.getGPT()
	return rlp.Encode(w, temp)
}

// only for serialization(storage/p2p)
type GlobalPropertyTemp struct {
	GlobalPropBase

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

func (gp *GlobalProperty) getGPT() *GlobalPropertyTemp {
	ajs := make([]common.Address, 0)
	ams := make([]common.Address, 0)
	pms := make([]common.Address, 0)

	for juryAdd := range gp.ActiveJuries {
		ajs = append(ajs, juryAdd)
	}

	for medAdd := range gp.ActiveMediators {
		ams = append(ams, medAdd)
	}

	for medAdd := range gp.PrecedingMediators {
		pms = append(pms, medAdd)
	}

	gpt := &GlobalPropertyTemp{
		GlobalPropBase:     gp.GlobalPropBase,
		ActiveJuries:       ajs,
		ActiveMediators:    ams,
		PrecedingMediators: pms,
	}

	return gpt
}

func (gpt *GlobalPropertyTemp) getGP(gp *GlobalProperty) {
	ajs := make(map[common.Address]bool)
	ams := make(map[common.Address]bool)
	pms := make(map[common.Address]bool)

	for _, addStr := range gpt.ActiveJuries {
		ajs[addStr] = true
	}

	for _, addStr := range gpt.ActiveMediators {
		ams[addStr] = true
	}

	for _, addStr := range gpt.PrecedingMediators {
		pms[addStr] = true
	}

	gp.GlobalPropBase = gpt.GlobalPropBase
	gp.ActiveJuries = ajs
	gp.ActiveMediators = ams
	gp.PrecedingMediators = pms
}
