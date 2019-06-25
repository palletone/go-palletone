/*
	This file is part of go-palletone.
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
*/

package deposit

import (
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

type DepositInterface interface {
	//  申请加入超级节点
	applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  交付保证金
	mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  申请退还部分保证金
	mediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  申请退出超级节点候选列表
	mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  交付保证金
	juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  申请退还部分保证金
	juryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  交付保证金
	developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  申请退还部分保证金
	developerApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  处理超级节点的申请
	handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理退还部分保证金的申请
	handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理退还部分保证金的申请
	handleForJuryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理退还部分保证金的申请
	handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理退出超级节点列表的申请
	handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理没收申请
	handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  申请没收保证金
	applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  普通节点质押PTN投票某个mediator
	normalNodePledgeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  普通节点修改所质押的mediator
	normalNodeChangeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  普通节点提取质押PTN
	normalNodeExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//
	handleExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//
	handleEachDayAward(stub shim.ChaincodeStubInterface, args []string) pb.Response
}
