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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 *
 */

package modules

import "github.com/palletone/go-palletone/common"

type ProofOfExistence struct {
	MainData   []byte         `json:"main_data"`
	ExtraData  []byte         `json:"extra_data"`
	Reference  []byte         `json:"reference"`
	UintHeight uint64         `json:"unit_index"`
	TxId       common.Hash    `json:"tx_id"`
	UnitHash   common.Hash    `json:"unit_hash"`
	Timestamp  uint64         `json:"timestamp"`
	Creator    common.Address `json:"creator"`
}
