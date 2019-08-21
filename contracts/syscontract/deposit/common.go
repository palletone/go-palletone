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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/syscontract"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

//  保存相关列表
func saveList(stub shim.ChaincodeStubInterface, key string, list map[string]bool) error {
	listByte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(key, listByte)
	if err != nil {
		return err
	}
	return nil
}

//  获取其他list
func getList(stub shim.ChaincodeStubInterface, typeList string) (map[string]bool, error) {
	byte, err := stub.GetState(typeList)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]bool)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//  判断 invokeTokens 是否包含保证金合约地址
func isContainDepositContractAddr(stub shim.ChaincodeStubInterface) (invokeToken *modules.InvokeTokens, err error) {
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return nil, err
	}

	for _, invokeTo := range invokeTokens {
		if strings.Compare(invokeTo.Address, syscontract.DepositContractAddress.String()) == 0 {
			if invokeTo.Asset.Equal(dagconfig.DagConfig.GetGasToken().ToAsset()) {
				return invokeTo, nil
			}
			return nil, fmt.Errorf("%s", "Deposit assets must be PTN")
		}
	}
	return nil, fmt.Errorf("it is not a depositContract invoke transaction")
}

//  处理部分保证金逻辑
func applyQuitList(role string, stub shim.ChaincodeStubInterface) error {
	//  获取请求调用地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return err
	}
	//  先获取申请列表
	listForQuit, err := GetListForQuit(stub)
	if err != nil {
		return err
	}
	// 判断列表是否为空
	if listForQuit == nil {
		listForQuit = make(map[string]*QuitNode)
	}
	quitNode := &QuitNode{
		Role: role,
		Time: getTiem(stub),
	}

	//  保存退还列表
	listForQuit[invokeAddr.String()] = quitNode
	err = SaveListForQuit(stub, listForQuit)
	if err != nil {
		return err
	}
	return nil

}

//  加入相应候选列表，mediator jury dev
func addCandaditeList(stub shim.ChaincodeStubInterface, invokeAddr common.Address, candidate string) error {
	//  获取列表
	list, err := getList(stub, candidate)
	if err != nil {
		return err
	}
	if list == nil {
		list = make(map[string]bool)
	}
	if list[invokeAddr.String()] {
		return fmt.Errorf("node was in the list")
	}
	list[invokeAddr.String()] = true
	listByte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(candidate, listByte)
	if err != nil {
		return err
	}
	return nil
}

//  从候选列表删除mediator jury dev
func moveCandidate(candidate string, invokeFromAddr string, stub shim.ChaincodeStubInterface) error {
	//
	list, err := getList(stub, candidate)
	if err != nil {
		log.Error("stub.GetCandidateList err:", "error", err)
		return err
	}
	//
	if list == nil {
		log.Error("stub.GetCandidateList err: list is nil")
		return fmt.Errorf("%s", "list is nil")
	}
	if !list[invokeFromAddr] {
		return fmt.Errorf("node was not in the list")
	}
	delete(list, invokeFromAddr)
	//
	err = saveList(stub, candidate, list)
	if err != nil {
		return err
	}
	return nil

}

//  保存没收列表
func SaveListForForfeiture(stub shim.ChaincodeStubInterface, list map[string]*Forfeiture) error {
	byte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(ListForForfeiture, byte)
	if err != nil {
		return err
	}
	return nil
}

//  获取没收列表
func GetListForForfeiture(stub shim.ChaincodeStubInterface) (map[string]*Forfeiture, error) {
	byte, err := stub.GetState(ListForForfeiture)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]*Forfeiture)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//  保存退款列表
func SaveListForQuit(stub shim.ChaincodeStubInterface, list map[string]*QuitNode) error {
	byte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(ListForQuit, byte)
	if err != nil {
		return err
	}
	return nil
}

//  获取退出列表
func GetListForQuit(stub shim.ChaincodeStubInterface) (map[string]*QuitNode, error) {
	byte, err := stub.GetState(ListForQuit)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]*QuitNode)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func mediatorDepositKey(medAddr string) string {
	return string(constants.MEDIATOR_INFO_PREFIX) + string(constants.DEPOSIT_BALANCE_PREFIX) + medAddr
}

//  获取mediator
func GetMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string) (*MediatorDeposit, error) {
	byte, err := stub.GetState(mediatorDepositKey(medAddr))
	if err != nil || byte == nil {
		return nil, err
	}
	balance := NewMediatorDeposit()
	err = json.Unmarshal(byte, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

//  保存mediator
func SaveMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string, balance *MediatorDeposit) error {
	byte, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	err = stub.PutState(mediatorDepositKey(medAddr), byte)
	if err != nil {
		return err
	}

	return nil
}

//  删除mediator
func DelMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string) error {
	err := stub.DelState(mediatorDepositKey(medAddr))
	if err != nil {
		return err
	}

	return nil
}

//  保存jury/dev
func SaveNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string, balance *DepositBalance) error {
	balanceByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	err = stub.PutState(string(constants.DEPOSIT_BALANCE_PREFIX)+balanceAddr, balanceByte)
	if err != nil {
		return err
	}
	return nil
}

//  获取jury/dev
func GetNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string) (*DepositBalance, error) {
	byte, err := stub.GetState(string(constants.DEPOSIT_BALANCE_PREFIX) + balanceAddr)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	balance := &DepositBalance{}
	err = json.Unmarshal(byte, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

//  删除jury/dev
func DelNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string) error {
	err := stub.DelState(string(constants.DEPOSIT_BALANCE_PREFIX) + balanceAddr)
	if err != nil {
		return err
	}
	return nil
}

//  判断是否基金会发起的
func isFoundationInvoke(stub shim.ChaincodeStubInterface) bool {
	//  判断是否基金会发起的
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return false
	}
	//  获取
	gp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := gp.ChainParameters.FoundationAddress
	// 判断当前请求的是否为基金会
	if invokeAddr.String() != foundationAddress {
		log.Error("please use foundation address")
		return false
	}
	return true
}

func getTiem(stub shim.ChaincodeStubInterface) string {
	t, _ := stub.GetTxTimestamp(10)
	ti := time.Unix(t.Seconds, 0)
	return ti.Format(Layout2)
}

func getToday(stub shim.ChaincodeStubInterface) string {
	t, _ := stub.GetTxTimestamp(10)

	ti := time.Unix(t.Seconds, 0)
	str := ti.Format("20060102")
	log.Debugf("getToday GetTxTimestamp 10 result:%d, format string:%s", t.Seconds, str)
	return str
}

//  社区申请没收某节点的保证金数量
func applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("applyForForfeitureDeposit")
	if len(args) != 3 {
		log.Error("args need three parameters")
		return shim.Error("args need three parameters")
	}
	//  需要判断是否基金会发起的
	//if !isFoundationInvoke(stub) {
	//	log.Error("please use foundation address")
	//	return shim.Error("please use foundation address")
	//}
	//  被没收地址
	forfeitureAddr := args[0]
	//  判断没收地址是否正确
	f, err := common.StringToAddress(forfeitureAddr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//  需要判断是否已经被没收过了
	listForForfeiture, err := GetListForForfeiture(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	//
	if listForForfeiture == nil {
		listForForfeiture = make(map[string]*Forfeiture)
	} else {
		//
		if _, ok := listForForfeiture[f.String()]; ok {
			return shim.Error("node was in the forfeiture list")
		}
	}
	//  被没收地址属于哪种类型
	role := args[1]
	//  没收理由
	extra := args[2]

	//  申请地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//  存储信息
	forfeiture := &Forfeiture{}
	forfeiture.ApplyAddress = invokeAddr.String()
	forfeiture.ForfeitureRole = role
	forfeiture.Extra = extra
	forfeiture.ApplyTime = getTiem(stub)
	listForForfeiture[f.String()] = forfeiture
	//  保存列表
	err = SaveListForForfeiture(stub, listForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

//是否在候选列表中
func isInCandidate(stub shim.ChaincodeStubInterface, invokeAddr string, candidate string) (bool, error) {
	list, err := getList(stub, candidate)
	if err != nil {
		log.Debugf("get list err: %s", err.Error())
		return false, err
	}
	if list == nil {
		return false, nil
	}
	if !list[invokeAddr] {
		return false, nil
	}
	return true, nil
}

//
func handleNode(stub shim.ChaincodeStubInterface, quitAddr common.Address, role string) error {
	//  移除退出列表
	listForQuit, err := GetListForQuit(stub)
	if err != nil {
		return err
	}
	delete(listForQuit, quitAddr.String())
	err = SaveListForQuit(stub, listForQuit)
	if err != nil {
		return err
	}
	//  获取该节点保证金数量
	b, err := GetNodeBalance(stub, quitAddr.String())
	if err != nil {
		return err
	}
	//  调用从合约把token转到请求地址
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	err = stub.PayOutToken(quitAddr.String(), modules.NewAmountAsset(b.Balance, gasToken), 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	list := ""
	if role == Developer {
		list = modules.DeveloperList
	}
	if role == Jury {
		list = modules.JuryList
	}
	//  移除候选列表
	err = moveCandidate(list, quitAddr.String(), stub)
	if err != nil {
		log.Error("moveCandidate err:", "error", err)
		return err
	}
	//  删除节点
	err = stub.DelState(string(constants.DEPOSIT_BALANCE_PREFIX) + quitAddr.String())
	if err != nil {
		log.Error("stub.DelState err:", "error", err)
		return err
	}
	return nil
}
func nodePayToDepositContract(stub shim.ChaincodeStubInterface, role string) pb.Response {
	log.Debug("enter nodePayToDepositContract")
	//  判断是否交付保证金交易
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("isContainDepositContractAddr err: ", "error", err)
		return shim.Error(err.Error())
	}

	gp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	cp := gp.ChainParameters
	//  交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  TODO 添加进入质押记录
	//err = pledgeDepositRep(stub, invokeAddr, invokeTokens.Amount)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//获取账户
	balance, err := GetNodeBalance(stub, invokeAddr.String())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	depositAmount := uint64(0)
	list := ""
	if role == Jury {
		depositAmount = cp.DepositAmountForJury
		list = modules.JuryList
	}
	if role == Developer {
		depositAmount = cp.DepositAmountForDeveloper
		list = modules.DeveloperList
	}
	//  第一次想加入
	if balance == nil {
		balance = &DepositBalance{}
		//  可以加入列表
		if invokeTokens.Amount != depositAmount {
			return shim.Error("Too many or too little.")
		}
		//  加入候选列表
		err = addCandaditeList(stub, invokeAddr, list)
		if err != nil {
			log.Error("addCandaditeList err: ", "error", err)
			return shim.Error(err.Error())
		}
		balance.EnterTime = getTiem(stub)
		//  没有
		balance.Balance = invokeTokens.Amount
		balance.Role = role
		err = SaveNodeBalance(stub, invokeAddr.String(), balance)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		//  追缴逻辑
		//if balance.Role != Jury {
		//	return shim.Error("not jury")
		//}
		all := balance.Balance + invokeTokens.Amount
		if all != depositAmount {
			return shim.Error("Too many or too little.")
		}
		//这里需要判断是否以及被基金会提前移除候选列表，即在规定时间内该节点没有追缴保证金
		b, err := isInCandidate(stub, invokeAddr.String(), list)
		if err != nil {
			log.Debugf("isInCandidate error: %s", err.Error())
			return shim.Error(err.Error())
		}
		if !b {
			//  加入jury候选列表
			err = addCandaditeList(stub, invokeAddr, list)
			if err != nil {
				log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
				return shim.Error(err.Error())
			}
		}
		balance.Balance = all
		err = SaveNodeBalance(stub, invokeAddr.String(), balance)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
}
