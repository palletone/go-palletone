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
	"encoding/hex"
	"errors"
	"github.com/martinlindhe/base36"
	"strconv"
	"strings"
)

var (
	TimeFormatString = "2006/01/02 15:04:05"
	PTNCOIN          = IDType16{0x40, 0x00, 0x82, 0xBB, 0x08, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00}
	BTCCOIN          = IDType16{'b', 't', 'c', 'c', 'o', 'i', 'n'}
)

// type 	Hash 		[]byte
const (
	ID_LENGTH = 16
)

type IDType16 [ID_LENGTH]byte

func ZeroIdType16() IDType16 {
	return IDType16{}
}

func (it *IDType16) String() string {
	return it.Str()
}

func (it *IDType16) ToAssetId() string {
	//if *it == PTNCOIN {
	//	return "PTN"
	//}
	symbol, assetType, decimal, txHash := it.ParseAssetId()
	//b12 := make([]byte, 11)
	//b12[0] = decimal
	//copy(b12[1:], txHash)
	return symbol + "+" + base36.EncodeBytes([]byte{decimal}) + strconv.Itoa(int(assetType)) + base36.EncodeBytes(txHash)
}
func String2AssetId(str string) (IDType16, error) {
	if str == "PTN" {
		return PTNCOIN, nil
	}
	strArray := strings.Split(str, "+")
	if len(strArray) < 2 {
		return IDType16{}, errors.New("Asset string invalid")
	}
	symbol := strArray[0]
	ty := strArray[1][1] - 48
	decimal := base36.DecodeToBytes(strArray[1][0:1])
	tx12 := base36.DecodeToBytes(strArray[1][2:])
	return NewAssetId(symbol, AssetType(ty), decimal[0], tx12)

}
func NewAssetId(symbol string, assetType AssetType, decimal byte, requestId []byte) (IDType16, error) {
	if len(symbol) > 5 {
		return IDType16{}, errors.New("Symbol must less than 5 characters")
	}
	assetId := IDType16{}
	assetSymbol := base36.DecodeToBytes(symbol)
	//fmt.Printf(base36.EncodeBytes(assetSymbol))
	copy(assetId[4-len(assetSymbol):4], assetSymbol)
	firstByte := assetId[0] | (byte(len(assetSymbol) << 5))
	firstByte = firstByte | byte(assetType)<<2
	assetId[0] = firstByte
	assetId[4] = decimal
	copy(assetId[5:], requestId[0:11])
	return assetId, nil
}

func (id *IDType16) ParseAssetId() (string, AssetType, byte, []byte) {
	var assetId [16]byte
	copy(assetId[:], id[:])
	assetId0 := id[0]
	len := assetId0 >> 5
	t := (assetId0 & 0xc) >> 2
	assetId[0] = assetId0 & 3
	symbol := base36.EncodeBytes(assetId[4-len : 4])
	return symbol, AssetType(t), assetId[4], assetId[5:]
}
func (it *IDType16) Str() string {
	return hex.EncodeToString(it.Bytes())
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
	bytes, err := hex.DecodeString(id)
	if err != nil {
		return IDType16{}, err
	}
	var id_type IDType16
	copy(id_type[0:], bytes)
	return id_type, nil
}
