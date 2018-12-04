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
	"github.com/martinlindhe/base36"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/dag/constants"
	"strconv"
	"time"
	"unsafe"
)

var (
	TimeFormatString = "2006/01/02 15:04:05"
	PTNCOIN   = IDType16{0x40, 0x00, 0x82, 0xBB, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	BTCCOIN = IDType16{'b', 't', 'c', 'c', 'o', 'i', 'n'}
)

// type 	Hash 		[]byte
const (
	ID_LENGTH = 16
	
)

type IDType16 [ID_LENGTH]byte
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

func ZeroIdType16() IDType16 {
	return IDType16{}
}

func (it *IDType16) String() string {
	return hexutil.Encode([]byte(it.Str()))
}
func (it *IDType16) ToAssetId() string {
	if *it == PTNCOIN {
		return "PTN"
	}
	var assetId [16]byte
	copy(assetId[:], it[:])
	assetId0 := it[0]
	len := assetId0 >> 5
	t := (assetId0 & 0xc) >> 2
	assetId[0] = assetId0 & 3
	symbol := base36.EncodeBytes(assetId[4-len : 4])
	return symbol + "+" + strconv.Itoa(int(t)) + base36.EncodeBytes(assetId[4:])
}
func (it *IDType16) Str() string {
	var b []byte

	b = append(b, it.Bytes()...)

	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == ' ' || b[i] == 0 {
			b = b[:i]
		}
	}
	return string(b)
}

func (it *IDType16) TokenType() string {
	return string(it.Bytes()[:])
}

func (it *IDType16) Bytes() []byte {
	idBytes := make([]byte, len(it))
	for i := 0; i < len(it); i++ {
		idBytes[i] = it[i]
	}
	return idBytes
}

func (it *IDType16) SetBytes(b []byte) {
	if len(b) > len(it) {
		b = b[len(b)-ID_LENGTH:]
	}

	copy(it[ID_LENGTH-len(b):], b)
}

func SetIdTypeByHex(id string) (IDType16, error) {
	bytes, err := hexutil.Decode(id)
	if err != nil {
		return IDType16{}, err
	}
	var id_type IDType16
	copy(id_type[0:], bytes)
	return id_type, nil
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
