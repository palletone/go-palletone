/* This file is part of go-palletone.
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

   @author PalletOne core developers <dev@pallet.one>
   @date 2018
*/
package modules

import "github.com/palletone/go-palletone/common"

type headerTemp struct {
	ParentsHash []common.Hash`json:"parents_hash"`
	//AssetIDs     []IDType16    `json:"assets"`
	Authors      Authentifier `json:"mediator"`    // the unit creation authors
	GroupSign    []byte       `json:"groupSign"`   // 群签名, 用于加快单元确认速度
	GroupPubKey  []byte       `json:"groupPubKey"` // 群公钥, 用于验证群签名
	TxRoot       common.Hash  `json:"root"`
	Number       *ChainIndex  `json:"index"`
	Extra        []byte       `json:"extra"`
	Creationdate uint32        `json:"creation_time"` // unit create time
}