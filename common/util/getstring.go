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
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

const (
	STRING_JOIN_CHAR = "\x00"
)

func GetSourceString(obj interface{}) string {
	var arrResult []string
	arrResult = extractComp(obj, arrResult)
	log.Println("result:", arrResult)
	return strings.Join(arrResult, STRING_JOIN_CHAR)
}

func extractComp(value interface{}, arrResult []string) []string {
	if value == nil {
		log.Println("value is nil.")
		return arrResult
	}
	switch reflect.TypeOf(value).String() {
	case "int", "int64", "float64", "float32":
		t := reflect.TypeOf(value).String()
		if t == "int" {
			arrResult = append(arrResult, "n", strconv.Itoa(value.(int)))
		} else if t == "int64" {
			arrResult = append(arrResult, "n", strconv.FormatInt(value.(int64), 10))
		} else if t == "float32" {
			arrResult = append(arrResult, "n", strconv.FormatFloat(value.(float64), 'E', -1, 32))
		} else if t == "float64" {
			arrResult = append(arrResult, "n", strconv.FormatFloat(value.(float64), 'E', -1, 64))
		}
	case "string":
		arrResult = append(arrResult, "s", value.(string))
		//arrResult = InsertArray("s", arrResult)
	case "bool":
		arrResult = append(arrResult, "b", strconv.FormatBool(value.(bool)))

	default:
		log.Println("type:", reflect.TypeOf(value).String())
		reftyp := reflect.TypeOf(value)
		refvalue := reflect.ValueOf(value)

		// typeOfType := refvalue.Type()
		// if typeOfType.Kind() == reflect.Ptr {
		// 	log.Println("=====")
		// 	typeOfType = typeOfType.Elem()

		// }
		var keys []string
		if strings.Contains(strings.ToLower(reftyp.String()), ".obj") {
			for i := 0; i < reftyp.NumField(); i++ {
				log.Println("key:", reftyp.Field(i).Name)
				keys = append(keys, reftyp.Field(i).Name)
			}
			sort.Strings(keys)
			for _, v := range keys {
				arrResult = append(arrResult, v)
				log.Println("value:=", refvalue.FieldByName(v).Interface())
				arrResult = extractComp(refvalue.FieldByName(v).Interface(), arrResult)
			}
		} else if strings.Contains(strings.ToLower(reftyp.String()), "map") {
			val_byte, _ := json.Marshal(value)
			// var val map[string]interface{}
			val := make(map[string]interface{})
			if err := json.Unmarshal(val_byte, &val); err != nil {
				log.Println("unmarshal err:", err, string(val_byte))
			}

			for k := range val {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, v := range keys {
				arrResult = append(arrResult, v)
				arrResult = extractComp(val[v], arrResult)
			}
			log.Println("map:", val)
		} else if strings.Contains(strings.ToLower(reftyp.String()), "[]") {
			arrvalue := value.([]string)
			arrResult = append(arrResult, "[")
			for i := 0; i < len(arrvalue); i++ {
				log.Println("i=", i, "v:=", arrvalue[i])
				arrResult = extractComp(arrvalue[i], arrResult)
			}
			arrResult = append(arrResult, "]")
		} else {
			log.Println("unknown type=", reftyp.String())
		}
	}
	return arrResult
}

func InsertArray(a string, arr []string) []string {
	lengh := len(arr)
	if lengh == 0 {
		arr = append(arr, a)
		return arr
	}
	arr = append(arr, a)
	for i := 0; i < lengh; i++ {
		arr[i], arr[lengh-1] = arr[lengh-1], arr[i]
	}
	return arr
}

func ToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func ToByte(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(sh))
}
func Bytes(value interface{}) ([]byte, error) {
	re := make([]byte, 0)
	if value == nil {
		return re, errors.New("value is nil.")
	}
	return json.Marshal(value)
}
