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
 *  * @date 2018
 *
 */

package walletjson

import (
	"encoding/json"
	"unsafe"
	"time"
)

type GetFileInfos struct {
	UnitHash    string `json:"unit_hash"`
	UintHeight  uint64 `json:"unit_index"`
	ParentsHash string `json:"parents_hash"`
	Txid        string `json:"txid"`
	Timestamp   time.Duration  `json:"timestamp"`
	FileData    string `json:"file_data"`
	ExtraData   string `json:"extra_data"`
}

func ConvertGetFileInfos2Json(gets []GetFileInfos) string {
	data, err := json.Marshal(gets)
	if err != nil {
		return ""
	}
	return *(*string)(unsafe.Pointer(&data))
}
