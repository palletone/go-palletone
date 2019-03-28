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

package sysconfigcc

import "github.com/palletone/go-palletone/contracts/shim"

type SysParamsConfInterface interface {
	//获取全部系统参数配置信息
	getAllSysParamsConf(stub shim.ChaincodeStubInterface) ([]byte,error)
	////通过键获取旧的相应的值
	//getOldSysParamByKey(stub shim.ChaincodeStubInterface,key string) ([]byte,error)
	////通过键获取当前申请修改的相应的值
	//getCurrSysParamByKey(stub shim.ChaincodeStubInterface,key string) ([]byte,error)
	//基金会发起更新某个系统参数（不需要投票）
	updateSysParamWithoutVote(stub shim.ChaincodeStubInterface,args []string) ([]byte, error)
	//通过键获取值
	getSysParamValByKey(stub shim.ChaincodeStubInterface,key string) ([]byte,error)
	//基金会发起更新某个系统参数（需要投票）
	//updateSysParamWithVote(stub shim.ChaincodeStubInterface,args []string) ([]byte,error)
}