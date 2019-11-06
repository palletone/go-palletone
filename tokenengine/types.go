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

package tokenengine

import "github.com/palletone/go-palletone/common"

//用指定地址对应的私钥对消息进行签名，并返回签名结果
type AddressGetSign func(common.Address, []byte) ([]byte, error)

//根据地址获得对应的公钥
type AddressGetPubKey func(common.Address) ([]byte, error)

//根据合约地址，获得该合约对应的陪审团赎回脚本
type PickupJuryRedeemScript func(common.Address) ([]byte, error)
