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
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag/util"
	"github.com/palletone/go-palletone/dag/common"
)

func NewAsset() common.IDType {
	var assetID common.IDType

	out := util.TimeUUID()
	byteOut := out.ToUUid()

	length := 0
	if len(byteOut) < len(assetID) {
		length = len(byteOut)
	} else {
		length = len(assetID)
	}
	for i := 0; i < length; i++ {
		assetID[i] = byteOut[i]
	}

	_, err := storage.Get(assetID.Bytes())
	if err != nil { // assetID already exists
		return common.IDType{}
	}
	return assetID
}
