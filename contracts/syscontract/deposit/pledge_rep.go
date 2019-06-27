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
	"github.com/palletone/go-palletone/dag/modules"
)

//质押充币
func pledgeDepositRep(stub shim.ChaincodeStubInterface, addr common.Address, amount uint64) error {
	addrStr := addr.String()
	node, err := getPledgeDepositRecord(stub, addrStr)
	if err != nil {
		return err
	}
	if node == nil {
		node = &modules.AddressAmount{}
	}
	node.Amount += amount
	node.Address = addrStr
	return savePledgeDepositRecord(stub, node)
}
//质押提币，以当天最后一次为准
func
pledgeWithdrawRep(stub shim.ChaincodeStubInterface, addr common.Address, amount uint64) error {
	err := savePledgeWithdrawRecord(stub, modules.NewAddressAmount(addr.String(), amount))
	return err
}
//撤销当天的提币请求
func pledgeWithdrawCancelRep(stub shim.ChaincodeStubInterface, addr common.Address) error {
	err := delPledgeWithdrawRecord(stub, addr.String())
	return err
}
//质押分红,按持仓比例分固定金额
func pledgeRewardAllocation(pledgeList *modules.PledgeList, rewardAmount uint64) *modules.PledgeList {
	newPledgeList := &modules.PledgeList{TotalAmount: 0, Members: []*modules.AddressAmount{}}
	rewardPerDao := float64(rewardAmount) / float64(pledgeList.TotalAmount)
	for _, pledge := range pledgeList.Members {
		newAmount := pledge.Amount + uint64(rewardPerDao*float64(pledge.Amount))
		newPledgeList.Members = append(newPledgeList.Members, &modules.AddressAmount{Address: pledge.Address, Amount: newAmount})
		newPledgeList.TotalAmount += newAmount
	}
	return newPledgeList
}
//查询一个账户的质押状态
func getPledgeStatus(stub shim.ChaincodeStubInterface, addr string) (*modules.PledgeStatus,error){
	d,err:= getPledgeDepositRecord(stub,addr)
	if err!=nil{
		return nil,err
	}
	w,err:=getPledgeWithdrawRecord(stub,addr)
	if err!=nil{
		return nil,err
	}
	list,err:=getLastPledgeList(stub)
	if err!=nil{
		return nil,err
	}
	status:=&modules.PledgeStatus{}
	if d!=nil{
		status.NewDepositAmount=d.Amount
	}
	if w!=nil{
		status.WithdrawApplyAmount=w.Amount
	}
	if list!=nil{
		status.PledgeAmount=list.GetAmount(addr)
	}
	//TODO 保证金部分？
	return status,nil
}