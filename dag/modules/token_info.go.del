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

package modules

import (
	"encoding/json"

	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/dag/constants"
	"time"
	"unsafe"
)

type AllTokenInfo struct {
	Items map[string]*TokenInfo //  token_info’json string
}
type TokenInfo struct {
	Name         string   `json:"name"`
	TokenStr     string   `json:"token_str"`
	TokenHex     string   `json:"token_hex"` // idtype16's hex
	Token        IDType16 `json:"token_id"`
	Creator      string   `json:"creator"`
	CreationDate string   `json:"creation_date"`
}

func NewTokenInfo(name, token, creator string) *TokenInfo {
	// 字符串转hex
	hex := hexutil.Encode([]byte(token))
	id, _ := SetIdTypeByHex(hex)
	return &TokenInfo{Name: name, TokenStr: token, TokenHex: hex, Token: id, Creator: creator, CreationDate: time.Now().Format(TimeFormatString)}
}

func (ti *TokenInfo) String() string {
	token_b, _ := json.Marshal(ti)
	return *(*string)(unsafe.Pointer(&token_b))
}

func Jsonbytes2AllTokenInfo(data []byte) (*AllTokenInfo, error) {
	info := new(AllTokenInfo)
	info.Items = make(map[string]*TokenInfo)
	err := json.Unmarshal(data, &info)

	return info, err
}

func (tf *AllTokenInfo) String() string {
	bytes, err := json.Marshal(tf)
	if err != nil {
		return ""
	}
	return *(*string)(unsafe.Pointer(&bytes))
}

func (tf *AllTokenInfo) Add(token *TokenInfo) {
	if tf == nil {
		tf = new(AllTokenInfo)
	}
	if tf.Items == nil {
		tf.Items = make(map[string]*TokenInfo)
	}
	tf.Items[string(constants.TOKENTYPE)+token.TokenHex] = token
}
