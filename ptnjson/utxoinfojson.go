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
package ptnjson

import (
	"fmt"
	"github.com/palletone/go-palletone/dag/modules"
)

type TokenInfoJson struct {
	Name         string `json:"name"`
	TokenHex     string `json:"token_hex"` // idtype16's hex
	Token        string `json:"token_id"`
	Creator      string `json:"creator"`
	CreationDate string `json:"creation_date"`
}

func ConvertTokenInfo2Json(tokenInfo *modules.TokenInfo) *TokenInfoJson {
	return &TokenInfoJson{
		Name:         tokenInfo.Name,
		TokenHex:     tokenInfo.TokenHex,
		Token:        tokenInfo.Token.String(),
		Creator:      tokenInfo.Creator,
		CreationDate: tokenInfo.CreationDate,
	}
}

func (this *TokenInfoJson) String() string {
	return fmt.Sprintf("Tokeninfo:\n    Name:%s\n    TokenHex:%s\n    Token:%s\n    Creator:%s\n    CreationDate:%s", this.Name, this.TokenHex, this.Token, this.Creator, this.CreationDate)
}
