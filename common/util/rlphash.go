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

package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"golang.org/x/crypto/sha3"
)

func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func RHashStr(x interface{}) string {
	x_byte, err := json.Marshal(x)
	if err != nil {
		return ""
	}
	s256 := sha256.New()
	s256.Write(x_byte)
	return fmt.Sprintf("%x", s256.Sum(nil))
}

func RHashBytes(x interface{}) []byte {
	x_byte, err := json.Marshal(x)
	if err != nil {
		return []byte("")
	}
	s256 := sha256.New()
	s256.Write(x_byte)
	return s256.Sum(nil)
}
