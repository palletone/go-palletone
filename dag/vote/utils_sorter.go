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
 * @author PalletOne core developer YiRan <dev@pallet.one>
 * @date 2018
 */
package vote

//Card @YiRan : struct for sorting map.
type ScoreCard struct {
	object interface{}
	score  uint64
}

type MapSorter []ScoreCard

func (ms MapSorter) Len() int {
	return len(ms)
}

//Less @YiRan : Descending order
func (ms MapSorter) Less(i, j int) bool {
	return ms[i].score > ms[j].score //
}
func (ms MapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

//NewMapSorter @YiRan : TODO:change score type to interface{}
func NewMapSorter(m map[interface{}]uint64) MapSorter {
	MapSorter := MapSorter{}
	for o, s := range m {
		MapSorter = append(MapSorter, ScoreCard{object: o, score: s})
	}
	return MapSorter
}
