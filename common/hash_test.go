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

package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHashFromStr(t *testing.T) {
	str := "e01c4bae7b396bc3c9bcb9275cef479560141c2010b6537abd78795bc935a2dd"
	hash := HexToHash(str)
	t.Logf("Hash:%s", hash.String())
	hash2 := &Hash{}
	err := hash2.SetHexString(str)
	t.Log(err)
	t.Logf("Hash:%s", hash2.String())

}
func TestHash_IsZero(t *testing.T) {
	h1 := Hash{}
	assert.True(t, h1.IsZero())
}
func TestHash_IsSelfHash(t *testing.T) {
	h2 := NewSelfHash()
	assert.True(t, h2.IsSelfHash())
}
