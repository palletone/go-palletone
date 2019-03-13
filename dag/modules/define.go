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

const (
	UNIT_CREATION_DATE_INITIAL_UINT64  = 1536451200
	UNIT_CREATION_DATE_INITIAL_FORMATE = "2018-09-09 00:00:00"
)

type DbRow struct {
	Key   []byte
	Value []byte
}
type KeyValue struct {
	Key   string
	Value []byte
}
