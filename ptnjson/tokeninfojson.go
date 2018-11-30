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
	"github.com/palletone/go-palletone/dag/modules"
)

// TokenInfoJson Yiran@ sturct made for display
type TokenInfoJson struct {
	Name         string `json:"name"`
	TokenStr     string `json:"token_str"`
	TokenHex     string `json:"token_hex"` // idtype16's hex
	Token        string `json:"token_id"`
	Creator      string `json:"creator"`
	CreationDate string `json:"creation_date"`
}

// ConvertTokenInfo2Json Yiran@convert token info to json format
func ConvertTokenInfo2Json(tokenInfo *modules.TokenInfo) *TokenInfoJson {
	return &TokenInfoJson{
		Name:         tokenInfo.Name,
		TokenStr:     tokenInfo.TokenStr,
		TokenHex:     tokenInfo.TokenHex,
		Token:        tokenInfo.Token.TokenType(),
		Creator:      tokenInfo.Creator,
		CreationDate: tokenInfo.CreationDate,
	}
}
