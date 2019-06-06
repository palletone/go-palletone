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

import (
	"github.com/palletone/go-palletone/contracts/shim"
)

type SysParamsConfInterface interface {
	//获取全部系统参数配置信息
	//getAllSysParamsConf(stub shim.ChaincodeStubInterface) (map[string]*modules.ContractStateValue, error)
	//基金会发起更新某个系统参数（不需要投票）
	updateSysParamWithoutVote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)
	//通过键获取值
	//getSysParamValByKey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)

	//通过投票方式来修改系统参数
	//首先，提供查询投票当前结果
	getVotesResult(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)
	//第一步：基金会创建投票tokens
	createVotesTokens(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)
	//第二步：基金会将投票tokens发给其他参与投票的节点：这一步需要基金会转账相应tokens给投票节点
	//第三步：参与投票的节点在截止日期之前投票即可
	nodesVote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)
	//第四步：到了mediators的时候更新当前得票数最高的结果
}
