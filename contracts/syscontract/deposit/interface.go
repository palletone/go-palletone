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
	//  超级节点交付规定保证金
	mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  超级节点申请退出候选列表
	mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  Jury节点交付规定保证金
	juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  Jury节点申请退出候选列表
	juryApplyQuit(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  dev节点交付规定保证金
	developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  dev节点申请退出候选列表
	devApplyQuit(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  处理超级节点的申请
	handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理超级节点退出
	handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理Jury节点退出
	handleForApplyQuitJury(stub shim.ChaincodeStubInterface, args []string) pb.Response
	//  处理dev节点退出
	handleForApplyQuitDev(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  处理没收申请
	handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  申请没收保证金
	applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  普通节点质押PTN投票某个mediator
	normalNodePledgeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//  普通节点提取质押PTN
	normalNodeExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response

	//
	handleEachDayAward(stub shim.ChaincodeStubInterface, args []string) pb.Response
}
