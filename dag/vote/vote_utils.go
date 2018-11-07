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

import (
	"fmt"
	"reflect"
)

//ToSliceInterface : trans map[anything]bool or []anything to []interface{}
func ToInterfaceSlice(s interface{}) []interface{} {
	ret := make([]interface{}, 0)

	v := reflect.ValueOf(s)
	switch v.Kind() {
	case reflect.Slice:
		l := v.Len()
		for i := 0; i < l; i++ {
			// ret[i] = v.Index(i).Interface()
			ret = append(ret, v.Index(i).Interface())
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			ret = append(ret, key.Interface())
		}
	default:
		fmt.Println("invalid parameter type, must be slice or map")
	}
	return ret
}

func resultNumber(inputLen uint8, resLenth uint8) uint8 {
	var resultNumber uint8
	if inputLen == 0 || inputLen > resLenth {
		resultNumber = resLenth
	} else {
		resultNumber = inputLen
	}
	return resultNumber
}

//TODO
func MapExist(m interface{}, k interface{}) bool {
	if reflect.ValueOf(m).Kind() == reflect.Map {
		//TODO
		return true
	}
	return false
}
