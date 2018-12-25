package deposit

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"strings"
	"time"
)

//处理交付保证金数据
func updateForPayValue(balance *modules.DepositBalance, invokeTokens *modules.InvokeTokens) {
	balance.TotalAmount += invokeTokens.Amount
	balance.LastModifyTime = time.Now().UTC()

	payTokens := &modules.InvokeTokens{}
	payValue := &modules.PayValue{PayTokens: payTokens}
	payValue.PayTokens.Amount = invokeTokens.Amount
	payValue.PayTokens.Asset = invokeTokens.Asset
	payValue.PayTime = time.Now().UTC()

	balance.PayValues = append(balance.PayValues, payValue)
}

//对结果序列化并更新数据
func marshalAndPutStateForBalance(stub shim.ChaincodeStubInterface, nodeAddr string, balance *modules.DepositBalance) error {
	balanceByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	err = stub.PutState(nodeAddr, balanceByte)
	if err != nil {
		return err
	}
	return nil
}

//加入申请提取列表
func addListAndPutStateForCashback(role string, stub shim.ChaincodeStubInterface, invokeAddr string, invokeTokens *modules.InvokeTokens) error {
	//先获取申请列表
	listForCashback, err := stub.GetListForCashback()
	if err != nil {
		return err
	}
	////序列化
	cashback := new(modules.Cashback)
	cashback.CashbackAddress = invokeAddr
	cashback.CashbackTokens = invokeTokens
	cashback.Role = role
	cashback.CashbackTime = time.Now().UTC().Unix()
	if listForCashback == nil {
		listForCashback = []*modules.Cashback{cashback}
	} else {
		isExist := isInCashbacklist(invokeAddr, listForCashback)
		if isExist {
			return fmt.Errorf("%s", "node is exist in the list.")
		}
		listForCashback = append(listForCashback, cashback)
	}
	//反序列化
	listForCashbackByte, err := json.Marshal(listForCashback)
	if err != nil {
		return err
	}
	err = stub.PutState("ListForCashback", listForCashbackByte)
	if err != nil {
		return err
	}
	return nil
}

//查找节点是否在列表中
func isInCashbacklist(addr string, list []*modules.Cashback) bool {
	for _, m := range list {
		if strings.Compare(addr, m.CashbackAddress) == 0 {
			return true
		}
	}
	return false
}

func applyCashbackList(role string, stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%s", "arg need one parameter.")
	}
	//获取 请求 调用 地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return err
	}
	//数量
	ptnAccount, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	//TODO 是否传进来
	asset := modules.NewPTNAsset()
	invokeTokens := &modules.InvokeTokens{
		Amount: ptnAccount,
		Asset:  asset,
	}
	//先获取数据库信息
	balance, err := stub.GetDepositBalance(invokeAddr)
	if err != nil {
		return err
	}
	if balance == nil {
		return fmt.Errorf("%s", "Your balance is nil.")
	}
	if balance.TotalAmount < invokeTokens.Amount {
		return fmt.Errorf("%s", "Your balance is not enough.")
	}
	if strings.Compare(role, "Mediator") == 0 {
		if balance.TotalAmount-invokeTokens.Amount < depositAmountsForMediator {
			return fmt.Errorf("%s", "Can not cashback some.")
		}
	}
	err = addListAndPutStateForCashback(role, stub, invokeAddr, invokeTokens)
	if err != nil {
		return err
	}
	return nil
}

//从 申请提取保证金列表中移除节点
func moveAndPutStateFromCashbackList(stub shim.ChaincodeStubInterface, cashbackAddr string, applyTime int64) error {
	//获取没收列表
	listForCashback, err := stub.GetListForCashback()
	if err != nil {
		return err
	}
	if listForCashback == nil {
		return fmt.Errorf("%s", "listForCashback is nil")
	}
	isExist := isInCashbacklist(cashbackAddr, listForCashback)
	if !isExist {
		return fmt.Errorf("%s", "node is not exist in the list.")
	}
	newList, isOk := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
	if !isOk {
		log.Error("Apply time is wrong.")
		return fmt.Errorf("%s", "Apply time is wrong.")
	}
	listForCashbackByte, err := json.Marshal(newList)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return err
	}
	//更新列表
	err = stub.PutState("ListForCashback", listForCashbackByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	return nil
}

//提取一部分保证金
func cashbackSomeDeposit(role string, stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *modules.Cashback, balance *modules.DepositBalance) error {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(cashbackAddr, cashbackValue.CashbackTokens, 0)
	if err != nil {
		return err
	}
	awards := award.GetAwardsWithCoins(balance.TotalAmount, balance.LastModifyTime.Unix())
	balance.LastModifyTime = time.Now().UTC()
	//加上利息奖励
	balance.TotalAmount += awards
	//减去提取部分
	balance.TotalAmount -= cashbackValue.CashbackTokens.Amount
	//TODO 如果推出后低于保证金，则退出列表
	if role == "Jury" {
		//如果推出后低于保证金，则退出列表
		if balance.TotalAmount < depositAmountsForJury {
			//handleMember("Jury", cashbackAddr, stub)
			err = moveCandidate("JuryList", cashbackAddr, stub)
			if err != nil {
				return err
			}
		}
	} else if role == "Developer" {
		//如果推出后低于保证金，则退出列表
		if balance.TotalAmount < depositAmountsForDeveloper {
			//handleMember("Developer", cashbackAddr, stub)
			err = moveCandidate("DeveloperList", cashbackAddr, stub)
			if err != nil {
				return err
			}
		}
	}
	//TODO 加入提款记录
	balance.CashbackValues = append(balance.CashbackValues, cashbackValue)
	//序列化
	err = marshalAndPutStateForBalance(stub, cashbackAddr, balance)
	if err != nil {
		return err
	}
	return nil
}

//同意提取保证金处理
//func handleCashback(stub shim.ChaincodeStubInterface, foundationAddr, cashbackAddr string, applyTime int64, balance *modules.DepositBalance) error {
//	//获取请求列表
//	listForCashback, err := stub.GetListForCashback()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	if listForCashback == nil {
//		return shim.Error("listForCashback is nil.")
//	}
//	//在申请退款保证金列表中移除该节点
//	//fmt.Println(listForCashback)
//	//fmt.Println(cashbackAddr)
//	//fmt.Println(applyTime)
//	cashbackValue, err := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	//fmt.Println(cashbackValue)
//	//fmt.Printf("%#v\n\n", cashbackValue)
//	if cashbackValue == nil {
//		return shim.Error("列表里没有该申请")
//	}
//	//还得判断一下是否超过余额
//	if cashbackValue.CashbackTokens.Amount > balance.TotalAmount {
//		return shim.Error("退款大于账户余额")
//	}
//
//}

//处理申请提保证金请求并移除列表
func cashbackAllDeposit(role string, stub shim.ChaincodeStubInterface, cashbackAddr string, invokeTokens *modules.InvokeTokens, balance *modules.DepositBalance) error {
	//计算保证金全部利息
	//获取币龄
	endTime := time.Now().UTC()
	coinDays := award.GetCoinDay(balance.TotalAmount, balance.LastModifyTime, endTime)
	//计算币龄收益
	awards := award.CalculateAwardsForDepositContractNodes(coinDays)
	//本金+利息
	invokeTokens.Amount += awards
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(cashbackAddr, invokeTokens, 0)
	if err != nil {
		return err
	}
	//移除出列表
	err = moveCandidate(role, cashbackAddr, stub)
	if err != nil {
		return err
	}
	//删除节点
	err = stub.DelState(cashbackAddr)
	if err != nil {
		return err
	}
	return nil
}

//Jury or developer 可以随时退保证金，只是不在列表中的话，没有奖励
func handleCommonJuryOrDev(stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *modules.Cashback, balance *modules.DepositBalance) error {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(cashbackAddr, cashbackValue.CashbackTokens, 0)
	if err != nil {
		return err
	}
	//fmt.Printf("balanceValue=%s\n", balanceValue)
	//v := handleValues(balanceValue.Values, tokens)
	//balanceValue.Values = v
	balance.LastModifyTime = time.Now().UTC()
	balance.TotalAmount -= cashbackValue.CashbackTokens.Amount
	//fmt.Printf("balanceValue=%s\n", balanceValue)
	//TODO
	balance.CashbackValues = append(balance.CashbackValues, cashbackValue)

	err = marshalAndPutStateForBalance(stub, cashbackAddr, balance)
	if err != nil {
		return err
	}
	return nil
}

func addCandaditeList(invokeAddr string, stub shim.ChaincodeStubInterface, candidate string) error {
	list, err := stub.GetCandidateList(candidate)
	if err != nil {
		return err
	}
	if list == nil {
		list = []string{invokeAddr}
	} else {
		list = append(list, invokeAddr)
	}
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

func moveCandidate(candidate string, invokeFromAddr string, stub shim.ChaincodeStubInterface) error {
	list, err := stub.GetCandidateList(candidate)
	if err != nil {
		return err
	}
	if list == nil {
		return fmt.Errorf("%s", "list is nil.")
	}
	for i := 0; i < len(list); i++ {
		if list[i] == invokeFromAddr {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	listBytes, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(candidate, listBytes)
	if err != nil {
		return err
	}
	return nil

}

//从申请没收保证金列表中移除
func moveInApplyForForfeitureList(stub shim.ChaincodeStubInterface, listForForfeiture []*modules.Forfeiture, forfeitureAddr string, applyTime int64) (newList []*modules.Forfeiture, isOk bool) {
	for i := 0; i < len(listForForfeiture); i++ {
		if listForForfeiture[i].ApplyTime == applyTime && listForForfeiture[i].ForfeitureAddress == forfeitureAddr {
			newList = append(listForForfeiture[:i], listForForfeiture[i+1:]...)
			isOk = true
			break
		}
	}
	return
}

//从申请没收保证金列表中移除
func moveInApplyForCashbackList(stub shim.ChaincodeStubInterface, listForCashback []*modules.Cashback, cashbackAddr string, applyTime int64) (newList []*modules.Cashback, isOk bool) {
	//
	for i := 0; i < len(listForCashback); i++ {
		if listForCashback[i].CashbackTime == applyTime && listForCashback[i].CashbackAddress == cashbackAddr {
			newList = append(listForCashback[:i], listForCashback[i+1:]...)
			isOk = true
			break
		}
	}
	return
	//listForCashbackByte, err := json.Marshal(listForCashback)
	//if err != nil {
	//	return nil, err
	//}
	////更新列表
	//err = stub.PutState("ListForCashback", listForCashbackByte)
	//if err != nil {
	//	return nil, err
	//}
	//return cashback, nil
}
