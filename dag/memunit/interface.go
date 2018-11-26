/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package memunit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type IMemDag interface {
	Save(unit *modules.Unit) error
	GetUnit(hash common.Hash) (*modules.Unit, error)
	UpdateMemDag(hash common.Hash, sign []byte) error
	Exists(uHash common.Hash) bool
	Prune(assetId string, maturedUnitHash common.Hash) error
	SwitchMainChain() error
	QueryIndex(assetId string, maturedUnitHash common.Hash) (uint64, int)
	GetCurrentUnit(assetid modules.IDType16, index uint64) (*modules.Unit, error)
	GetDelhashs() chan common.Hash
	PushDelHashs(hashs []common.Hash)
}
