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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"golang.org/x/crypto/sha3"
)

func Sha1(data []byte) string {
	s1 := sha1.New()
	s1.Write(data)
	return fmt.Sprintf("%x", s1.Sum(nil))
}
func Sha256(data []byte) string {
	s256 := sha256.New()
	s256.Write(data)
	return fmt.Sprintf("%x", s256.Sum(nil))
}
func Sha512(data []byte) string {
	s512 := sha512.New()
	s512.Write(data)
	return fmt.Sprintf("%x", s512.Sum(nil))
}
func Md5(data []byte) string {
	m5 := md5.New()
	m5.Write(data)
	return fmt.Sprintf("%x", m5.Sum(nil))
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.New256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
