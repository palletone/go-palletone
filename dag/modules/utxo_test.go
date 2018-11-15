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
 *  * @date 2018
 *
 */

package modules

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAsset_String(t *testing.T) {
	asset := Asset{AssetId: PTNCOIN}
	t.Log(asset.String())
	asset2 := Asset{AssetId: IDType16{'t', 'e', 's', 't'}, UniqueId: IDType16{'1', '2', '3', '4'}}
	t.Logf("ERC721 asset:%s", asset2.String())

	asset11 := Asset{}
	asset11.SetString(asset.String())
	assert.Equal(t, asset, asset11)
	asset22 := Asset{}
	err := asset22.SetString(asset2.String())
	assert.Nil(t, err, "SetString error")
	assert.Equal(t, asset22, asset2)

}
