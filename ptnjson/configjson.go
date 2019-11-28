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

package ptnjson

import (
	"fmt"
	"github.com/palletone/go-palletone/core"
	"reflect"
	"strconv"
)

// 配置参数的键值对，区块链浏览器专用
type ConfigJson struct {
	Key   string `json:"key"`   // 配置参数的key
	Value string `json:"value"` // 配置参数的value
}

func ConvertAllSysConfigToJson(configs *core.ChainParameters) []*ConfigJson {
	result := make([]*ConfigJson, 0)
	tt := reflect.TypeOf(*configs)
	vv := reflect.ValueOf(*configs)

	for i := 0; i < vv.NumField(); i++ {
		sf := tt.Field(i)
		sv := vv.Field(i)
		scjs := getConfigJson(sf, sv)

		result =append(result, scjs...)
	}

	return result
}

func getConfigJson(sf reflect.StructField, vv reflect.Value) []*ConfigJson {
	var res []*ConfigJson

	sft := sf.Type
	if sft.Kind() == reflect.Struct {
		for i := 0; i < vv.NumField(); i++  {
			ssf := sft.Field(i)
			svv := vv.Field(i)
			cjs := getConfigJson(ssf, svv)

			res = append(res, cjs...)
		}
	} else {
		cj := &ConfigJson{Key: sf.Name, Value: toString(vv)}
		res = append(res, cj)
	}

	return res
}

func toString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Invalid:
		return "invalid field"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.String:
		return v.String()
	case reflect.Float64, reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Sprintf("unexpected type: %v", v.Type().String())
	}
}
