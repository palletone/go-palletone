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

package storage

import (
	"github.com/palletone/go-palletone/common"
	"log"
	"testing"
)

func TestGet(t *testing.T) {
	if m := GetPrefix([]byte("array")); m != nil {
		for k, v := range m {
			log.Println("key: ", k, "value: ", string(v))
		}
	}

	if m := GetPrefix([]byte("20")); m != nil {
		for k, v := range m {
			log.Println("key: ", k, "value: ", string(v))
		}
	}

	if m := GetPrefix([]byte("unit")); m != nil {
		for k, _ := range m {
			log.Println("key: ", k, "value: ", string(m[k]))
		}
	}
}

func TestGetUnit(t *testing.T) {
	GetUnit(common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"), 0)
}
