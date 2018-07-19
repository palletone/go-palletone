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
	"github.com/palletone/go-palletone/common/util"
)

var TimeFormatString = "2006/01/02 15:04:05"

// type 	Hash 		[]byte
const (
	ID_LENGTH = 32
)

type IDType16 [ID_LENGTH]byte

var (
	PTNCOIN = IDType16{'p', 't', 'n', ' ', '0', 'f', '5', ' ', ' '}
	BTCCOIN = IDType16{'b', 't', 'c', '0', ' '}
)

func (it *IDType16) String() string {
	var b []byte
	length := len(it)
	for _, v := range it {
		b = append(b, v)
	}
	count := 0
	for i := length - 1; i >= 0; i-- {
		if b[i] == ' ' || b[i] == 0 {
			count++
		} else {
			break
		}
	}
	return util.ToString(b[:length-count])
}

func (it *IDType16) Bytes() []byte {
	idBytes := make([]byte, len(it))
	for i := 0; i < len(it); i++ {
		idBytes[i] = it[i]
	}
	//return idBytes
	return idBytes
}

func (it *IDType16) SetBytes(b []byte) {
	if len(b) > len(it) {
		b = b[len(b)-ID_LENGTH:]
	}

	copy(it[ID_LENGTH-len(b):], b)
}
