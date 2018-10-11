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

package asset

import (
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn"
	"github.com/pborman/uuid"
)

func NewAsset() modules.IDType16 {
	var assetId modules.IDType16

	// use version 1: timestamp and mac
	uuid := uuid.NewUUID()
	lenth := len(uuid)
	if lenth > cap(assetId) {
		lenth = cap(assetId)
	}
	for i := 0; i < lenth; i++ {
		assetId[i] = uuid[i]
	}
	return assetId
}
func PTN() *modules.Asset {

	return &modules.Asset{AssetId: modules.PTNCOIN,
		UniqueId: modules.ZeroIdType16(),
		ChainId:  ptn.DefaultConfig.NetworkId,
	}
}
