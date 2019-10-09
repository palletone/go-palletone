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

package deposit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpperFirstChar(t *testing.T) {
	a := ""
	a1 := UpperFirstChar(a)
	assert.Equal(t, a, a1)
	b := "a"
	b1 := UpperFirstChar(b)
	assert.Equal(t, "A", b1)
	c := "aB"
	c1 := UpperFirstChar(c)
	assert.Equal(t, "AB", c1)
	d := "AB"
	d1 := UpperFirstChar(d)
	assert.Equal(t, "AB", d1)
}
