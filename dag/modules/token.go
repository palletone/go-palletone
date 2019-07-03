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

package modules

import (
	"time"

	"github.com/palletone/go-palletone/common"
)

const GlobalPrefix = "Tokens_"

//定义所有Token的基本信息
type GlobalTokenInfo struct {
	Symbol      string
	TokenType   uint8 //1:prc20 2:prc721 3:vote 4:SysVote
	Status      uint8
	CreateAddr  string
	TotalSupply uint64
	SupplyAddr  string
	AssetID     AssetId
}

//定义一种全新的Token
type TokenDefine struct {
	TokenDefineJson []byte         `json:"token_define_json"`
	TokenType       int            `json:"token_type"` //0 ERC20  1 ERC721   2 VoteToken
	Creator         common.Address `json:"creator"`
}

//增发一种已经定义好的Token
type TokenSupply struct {
	UniqueId []byte
	AssetId  []byte
	Amount   uint64
	Creator  common.Address
}

//同质化通证，比如ERC20
type FungibleToken struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals byte   `json:"decimals"`
	//总发行量
	TotalSupply uint64 `json:"total_supply"`
	//如果允许增发，那么允许哪个地址进行增发，如果为空则不允许增发
	SupplyAddress string `json:"supply_address"`
}

type NonFungibleMetaData struct {
	UniqueBytes []byte `json:"UniqueBytes"`
}

//非同质化通证，比如ERC721
type NonFungibleToken struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Type   byte   `json:"type"`
	//总发行量
	TotalSupply     uint64                `json:"total_supply"`
	NonFungibleData []NonFungibleMetaData `json:"NonFungibleData"`
	SupplyAddress   string                `json:"supply_address"`
}

//为投票而创建的Token
type VoteToken struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	//该投票是否允许改投
	VoteType byte `json:"vote_type"`
	//投票结束时间
	VoteEndTime time.Time `json:"vote_end_time"`
	//投票内容，JSON格式的表单
	VoteContent []byte `json:"vote_content"`
	//总发行量
	TotalSupply   uint64 `json:"total_supply"`
	SupplyAddress string `json:"supply_address"`
}
