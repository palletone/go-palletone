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

//记录了所有用户的质押充币、提币、分红等过程
//最新状态集
//Advance：形成流水日志，
package deposit

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	//pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	//"github.com/palletone/go-palletone/dag/constants"
	//"github.com/palletone/go-palletone/dag/modules"
	//"github.com/shopspring/decimal"
)

//充币
func depositToken(stub shim.ChaincodeStubInterface, addr common.Address, amount uint64) {

}

//提币申请
func withdrawTokenApply(stub shim.ChaincodeStubInterface, addr common.Address, amount uint64) {

}

//质押分红
func rewardDeposit(stub shim.ChaincodeStubInterface) {

}
func tokenChangeLog(stub shim.ChaincodeStubInterface, addr common.Address, log string) {
	//TODO
}
