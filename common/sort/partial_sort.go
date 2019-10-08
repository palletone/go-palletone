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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package sort

import "sort"

// Rearranges elements such that the range [0, m)
// contains the sorted m smallest elements in the range [first, data.Len).
// The order of equal elements is not guaranteed to be preserved.
// The order of the remaining elements in the range [m, data.Len) is unspecified.
func PartialSort(data sort.Interface, m int) {
	len := data.Len()

	// Build max-heap
	makeHeap(data, 0, m)
	minElemIdx := 0

	// Traverse the subsequent elements
	for i := m; i < len; i++ {
		if data.Less(i, minElemIdx) {
			// swap when this element is smaller than the largest element
			data.Swap(i, minElemIdx)
			// Rearrange the heap to max-heap
			siftDown(data, minElemIdx, m, minElemIdx)
		}
	}

	// Sort this heap
	heapsort(data, 0, m)
}

func IsPartialSorted(data sort.Interface, m int) bool {
	mid := m - 1
	for i := mid; i > 0; i-- {
		if data.Less(i, i-1) {
			return false
		}
	}

	n := data.Len()
	for i := m; i < n; i++ {
		if data.Less(i, mid) {
			return false
		}
	}

	return true
}
