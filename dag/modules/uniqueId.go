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
 *  * @date 2018-2019
 *
 */

package modules

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

type UniqueId [ID_LENGTH]byte
type UniqueIdType byte
func ZeroUniqueId() UniqueId {
	return UniqueId{}
}
const (
	UniqueIdType_Null       UniqueIdType = iota
	UniqueIdType_Sequence   UniqueIdType = 1
	UniqueIdType_Uuid       UniqueIdType = 2
	UniqueIdType_UserDefine UniqueIdType = 3
	UniqueIdType_Ascii      UniqueIdType = 4
)

func FormatUUID(buf []byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}
func ParseUUID(uuid string) ([]byte, error) {
	uuidLen := 16
	if len(uuid) != 2*uuidLen+4 {
		return nil, fmt.Errorf("uuid string is wrong length")
	}

	if uuid[8] != '-' ||
		uuid[13] != '-' ||
		uuid[18] != '-' ||
		uuid[23] != '-' {
		return nil, fmt.Errorf("uuid is improperly formatted")
	}

	hexStr := uuid[0:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:36]

	ret, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	if len(ret) != uuidLen {
		return nil, fmt.Errorf("decoded hex is the wrong length")
	}

	return ret, nil
}
func (it *UniqueId) String() string {

	return hex.EncodeToString(it.Bytes())

}
func (it *UniqueId) StringFriendly(t UniqueIdType) string {
	switch t {
	case UniqueIdType_Sequence:
		{
			i := big.Int{}
			i.SetBytes(it.Bytes())
			return i.String()
		}
	case UniqueIdType_Null:
		return ""
	case UniqueIdType_Uuid:
		return FormatUUID(it.Bytes())
	case UniqueIdType_UserDefine:
		return hex.EncodeToString(it.Bytes())
	case UniqueIdType_Ascii:
		return string(it.Bytes())
	}
	return ""

}
func String2UniqueId(str string, t UniqueIdType) (UniqueId, error) {
	uid := UniqueId{}
	switch t {
	case UniqueIdType_Sequence:
		{
			i := big.Int{}
			i.SetString(str, 0)
			uid.SetBytes(i.Bytes())
			return uid, nil
		}
	case UniqueIdType_Null:
		return uid, nil
	case UniqueIdType_Uuid:
		{
			b, err := ParseUUID(str)
			uid.SetBytes(b)
			return uid, err
		}

	case UniqueIdType_UserDefine:
		{
			b, err := hex.DecodeString(str)
			uid.SetBytes(b)
			return uid, err
		}
	case UniqueIdType_Ascii:
		{
			b := []byte(str)
			uid.SetBytes(b)
			return uid, nil
		}
	}

	return uid, errors.New("Unknown UniequeIdType")
}
func (it *UniqueId) Bytes() []byte {
	return it[:]
}

func (it *UniqueId) SetBytes(b []byte) {
	if len(b) > len(it) {
		b = b[len(b)-ID_LENGTH:]
	}

	copy(it[ID_LENGTH-len(b):], b)
}
