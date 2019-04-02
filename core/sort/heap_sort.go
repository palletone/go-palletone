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

// siftDown implements the heap property on data[lo, hi).
// first is an offset into the array where the root of the heap lies.
// (Reference from Go SDK 1.10.2 package sort.)
func siftDown(data sort.Interface, lo, hi, first int) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}
		if child+1 < hi && data.Less(first+child, first+child+1) {
			child++
		}
		if !data.Less(first+root, first+child) {
			return
		}
		data.Swap(first+root, first+child)
		root = child
	}
}

// makeHeap implement Build heap with greatest element at top.
func makeHeap(data sort.Interface, a, b int) {
	first := a
	hi := b - a

	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown(data, i, hi, first)
	}
}

// heapsort implement Heap sorting a max-heap
func heapsort(data sort.Interface, a, b int) {
	first := a
	hi := b - a

	// Pop elements, largest first, into end of data.
	for i := hi - 1; i >= 0; i-- {
		data.Swap(first, first+i)
		siftDown(data, first, i, first)
	}
}

// heapSort implement heap sorting of all elements in data
func HeapSort(data sort.Interface) {
	heapSort(data, 0, data.Len())
}

// heapSort implement heap sorting of all elements in data[a, b)
func heapSort(data sort.Interface, a, b int) {
	makeHeap(data, a, b)
	heapsort(data, a, b)
}
