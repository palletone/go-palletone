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
	"github.com/palletone/go-palletone/common/hexutil"
)

var (
	TimeFormatString = "2006/01/02 15:04:05"

	PTNCOIN = IDType16{'p', 't', 'n', 'c', 'o', 'i', 'n'}
	BTCCOIN = IDType16{'b', 't', 'c', 'c', 'o', 'i', 'n'}
)

// type 	Hash 		[]byte
const (
	ID_LENGTH = 16
)

type IDType16 [ID_LENGTH]byte
type TokenTypes struct {
	Tokens []IDType16
}

func ZeroIdType16() IDType16 {
	return IDType16{}
}

func (it *IDType16) String() string {
	return hexutil.Encode(it.Bytes()[:])
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
