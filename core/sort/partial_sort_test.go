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

// (Reference from Go SDK 1.10.2 package sort.)
*/
/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package sort_test

import (
	"math"
	"sort"
	"testing"

	csort "github.com/palletone/go-palletone/core/sort"
)

var ints = [...]int{74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586}
var float64s = [...]float64{74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, math.NaN(), math.NaN(), math.Inf(-1), 9845.768, -959.7485, 905, 7.8, 7.8}
var strings = [...]string{"", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&", "***"}

func TestPartialSortIntSlice(t *testing.T) {
	data := ints
	a := sort.IntSlice(data[0:])

	m := 4
	csort.PartialSort(a, m)

	if !csort.IsPartialSorted(a, m) {
		t.Errorf("sorted %v", ints)
		t.Errorf("   got %v", data)
	}
}

func TestPartialSortFloat64Slice(t *testing.T) {
	data := float64s
	a := sort.Float64Slice(data[0:])

	m := 5
	csort.PartialSort(a, m)

	if !csort.IsPartialSorted(a, m) {
		t.Errorf("sorted %v", ints)
		t.Errorf("   got %v", data)
	}
}

func TestPartialSortStringSlice(t *testing.T) {
	data := strings
	a := sort.StringSlice(data[0:])

	m := 3
	csort.PartialSort(a, m)

	if !csort.IsPartialSorted(a, m) {
		t.Errorf("sorted %v", ints)
		t.Errorf("   got %v", data)
	}
}
