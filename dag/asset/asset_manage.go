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
	"github.com/pborman/uuid"
	"fmt"
)

func NewAsset() modules.IDType36 {
	var assetId modules.IDType36

	// use version 1: timestamp and mac
	sId := uuid.NewUUID().String()

	if len(sId) > cap(assetId) {
		return assetId
	}

	bID := []byte(sId)
	for i:=0; i<len(bID); i++ {
		assetId[i] = bID[i]
	}

	for j:=len(bID); j<cap(assetId); j++ {
		fmt.Println(j)
		assetId[j] = '_'
	}
	return assetId
}