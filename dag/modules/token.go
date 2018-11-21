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

import "time"

type TokenSupply struct {
	TokenDefineJson string
	AssetId         string
	TokenType       int
}

//同质化通证，比如ERC20
type FungibleToken struct {
	Name     string
	Symbol   string
	Decimals byte
	//总发行量
	TotalSupply uint64
	//如果允许增发，那么允许哪个地址进行增发，如果为空则不允许增发
	SupplyAddress string
}

//非同质化通证，比如ERC721
type NonFungibleToken struct {
	Name   string
	Symbol string
	//总发行量
	TotalSupply   uint64
	SupplyAddress string
}

//为投票而创建的Token
type VoteToken struct {
	Name   string
	Symbol string
	//该投票是否允许改投
	VoteType byte
	//投票结束时间
	VoteEndTime time.Time
	//投票内容，JSON格式的表单
	VoteContent []byte
	//总发行量
	TotalSupply   uint64
	SupplyAddress string
}
