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
	"github.com/palletone/go-palletone/dag/modules"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
)

// NewUnitEvent is posted when a unit has been produced.
type NewProducedUnitEvent struct {
	Unit *modules.Unit
}

type SigShareEvent struct {
	UnitHash common.Hash
	SigShare []byte
}

type VSSDealEvent struct {
	DstIndex uint32
	Deal     *dkg.Deal
}

type VSSResponseEvent struct {
	Resp *dkg.Response
}

type GroupSigEvent struct {
	UnitHash common.Hash
	GroupSig []byte
}
