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
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/syscontract"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
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
			return invokeTo, nil
		}
	}
	return nil, fmt.Errorf("it is not a depositContract invoke transaction")
}

//  处理部分保证金逻辑
func applyCashbackList(role string, stub shim.ChaincodeStubInterface, args []string) error {
	//  判断参数是否正确
	if len(args) != 1 {
		return fmt.Errorf("%s", "arg need one parameter")
	}
	//  转换保证金数量
	ptnAccount, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	//  获取资产类型
	fees, err := stub.GetInvokeFees()
	if err != nil {
		return err
	}
	invokeTokens := &modules.AmountAsset{
		Amount: ptnAccount,
		Asset:  fees.Asset,
	}
	//  获取请求调用地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return err
	}

	var balance uint64
	if role == Mediator {
		md, _ := GetMediatorDeposit(stub, invokeAddr.String())
		if md == nil {
			return fmt.Errorf("%s", "mediator balance is nil")
		}
		if !isOverDeadline(stub, md.EnterTime) {
			return fmt.Errorf("does not over deadline")
		}
		balance = md.Balance
	} else {
		//  先获取账户信息
		deposit, _ := GetNodeBalance(stub, invokeAddr.String())
		if deposit == nil {
			return fmt.Errorf("%s", "balance is nil")
		}
		//  如果jury或者Dev 已经加入了候选列表，需要判断是否超过质押期限
		if deposit.EnterTime != "" {
			if !isOverDeadline(stub, deposit.EnterTime) {
				return fmt.Errorf("does not over deadline")
			}
		}
		balance = deposit.Balance
	}

	//  判断余额与当前退还的比较
	if balance < invokeTokens.Amount {
		return fmt.Errorf("%s", "balance is not enough")
	}

	//  对mediator的特殊处理
	if role == Mediator {
		//  获取保证金下限
		//depositAmountsForMediatorStr, err := stub.GetSystemConfig(modules.DepositAmountForMediator)
		//if err != nil {
		//	return err
		//}
		////  转换
		//depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
		//if err != nil {
		//	return err
		//}
		cp, err := stub.GetSystemConfig()
		if err != nil {
			//log.Error("strconv.ParseUint err:", "error", err)
			return err
		}
		depositAmountsForMediator := cp.DepositAmountForMediator
		//  判断退还后是否还在保证金下线之上
		if balance-invokeTokens.Amount < depositAmountsForMediator {
			return fmt.Errorf("%s", "can not cashback some")
		}
	}
	//  先获取申请列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		return err
	}
	// 判断列表是否为空
	if listForCashback == nil {
		listForCashback = make(map[string]*Cashback)
	} else {
		if _, ok := listForCashback[invokeAddr.String()]; ok {
			return fmt.Errorf("node is exist in the list")
		}
	}
	cashback := &Cashback{}
	cashback.CashbackTokens = invokeTokens
	cashback.Role = role
	cashback.CashbackTime = TimeStr()

	//  保存退还列表
	listForCashback[invokeAddr.String()] = cashback
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		return err
	}
	return nil
}

//  从申请提取保证金列表中移除节点
func moveAndPutStateFromCashbackList(stub shim.ChaincodeStubInterface, cashbackAddr common.Address) error {
	//获取没收列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("stub.GetListForCashback err:", "error", err)
		return err
	}
	if listForCashback == nil {
		log.Error("listForCashback is nil")
		return fmt.Errorf("%s", "listForCashback is nil")
	}
	if _, ok := listForCashback[cashbackAddr.String()]; !ok {
		log.Error("node is not exist in the cashback list.")
		return fmt.Errorf("%s", "node is not exist in the cashback list.")
	}
	delete(listForCashback, cashbackAddr.String())
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	return nil
}

//Jury or developer 可以随时退保证金，只是不在列表中的话，没有奖励
func handleCommonJuryOrDev(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//  这里计算这一次操作的币龄利息
	awards := caculateAwards(stub, balance.Balance, balance.LastModifyTime)
	balance.Balance += awards
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(cashbackAddr.String(), cashbackValue.CashbackTokens, 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	balance.LastModifyTime = TimeStr()
	balance.Balance -= cashbackValue.CashbackTokens.Amount

	err = SaveNodeBalance(stub, cashbackAddr.String(), balance)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
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
func SaveListForCashback(stub shim.ChaincodeStubInterface, list map[string]*Cashback) error {
	byte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(ListForCashback, byte)
	if err != nil {
		return err
	}
	return nil
}

//  获取退款列表
func GetListForCashback(stub shim.ChaincodeStubInterface) (map[string]*Cashback, error) {
	byte, err := stub.GetState(ListForCashback)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]*Cashback)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//  社区申请没收某节点的保证金数量
func applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("applyForForfeitureDeposit")
	if len(args) != 4 {
		log.Error("args need four parameters")
		return shim.Error("args need four parameters")
	}
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
	//  被没收数量
	amount := args[2]
	//  获取没收保证金数量，将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	var balance uint64
	if role == Mediator {
		md, err := GetMediatorDeposit(stub, forfeitureAddr)
		if err != nil {
			return shim.Error(err.Error())
		}
		balance = md.Balance
	} else {
		//  获取该节点账户
		db, err := GetNodeBalance(stub, forfeitureAddr)
		if err != nil {
			return shim.Error(err.Error())
		}
		balance = db.Balance
	}

	//  比较没收数量
	if ptnAccount > balance {
		return shim.Error("forfeituring to many ")
	}

	//  TODO 如果时没收mediator则，要么没收所有，要么没收后，该节点的保证金还在规定的下限之上
	//if role == Mediator {
	//	//
	//	amount, err := stub.GetSystemConfig(modules.DepositAmountForMediator)
	//	if err != nil {
	//		return shim.Error(err.Error())
	//	}
	//	//  转换保证金数量
	//	depositAmountsForMediator, err := strconv.ParseUint(amount, 10, 64)
	//	if err != nil {
	//		log.Error("strconv.ParseUint err:", "error", err)
	//		return shim.Error(err.Error())
	//	}
	//	result := balance - ptnAccount
	//	if result < depositAmountsForMediator {
	//		return shim.Error("can not forfeiture some deposit amount for mediator")
	//	}
	//}
	fees, err := stub.GetInvokeFees()
	if err != nil {
		log.Error("stub.GetInvokeFees err:", "error", err)
		return shim.Error(err.Error())
	}
	invokeTokens := &modules.AmountAsset{
		Amount: ptnAccount,
		Asset:  fees.Asset,
	}
	//  没收理由
	extra := args[3]
	//  需要判断是否基金会发起的
	//if !isFoundationInvoke(stub) {
	//	log.Error("please use foundation address")
	//	return shim.Error("please use foundation address")
	//}
	//  申请地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//  存储信息
	forfeiture := &Forfeiture{}
	forfeiture.ApplyAddress = invokeAddr.String()
	forfeiture.ForfeitureAddress = forfeitureAddr
	forfeiture.ApplyTokens = invokeTokens
	forfeiture.ForfeitureRole = role
	forfeiture.Extra = extra
	forfeiture.ApplyTime = TimeStr()
	listForForfeiture[f.String()] = forfeiture
	//  保存列表
	err = SaveListForForfeiture(stub, listForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

func mediatorDepositKey(medAddr string) string {
	return string(constants.MEDIATOR_INFO_PREFIX) + string(constants.DEPOSIT_BALANCE_PREFIX) + medAddr
}

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

func DelMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string) error {
	err := stub.DelState(mediatorDepositKey(medAddr))
	if err != nil {
		return err
	}

	return nil
}

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

func DelNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string) error {
	err := stub.DelState(string(constants.DEPOSIT_BALANCE_PREFIX) + balanceAddr)
	if err != nil {
		return err
	}
	return nil
}

//  将字符串时间格式转换为time类型
func StrToTime(strT string) time.Time {
	t, _ := time.Parse(Layout2, strT[:19])
	return t
}

//  将当前的time格式类型转换为字符串格式
func TimeStr() string {
	timeStr := time.Now().UTC().Format(Layout1)
	tt, _ := time.Parse(Layout1, timeStr)
	return tt.String()
}

// 判读是否超过了抵押日期
func isOverDeadline(stub shim.ChaincodeStubInterface, enterTime string) bool {
	//  判断是否超过了质押周期
	//depositPeriod, err := stub.GetSystemConfig(DepositPeriod)
	//if err != nil {
	//	log.Error("get deposit period err: ", "error", err)
	//	return false
	//}
	////
	//day, err := strconv.Atoi(depositPeriod)
	//if err != nil {
	//	log.Error("strconv.Atoi err: ", "error", err)
	//	return false
	//}
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	day := cp.DepositPeriod
	//nowT := time.Now().UTC()
	enterT := StrToTime(enterTime)
	dur := int(time.Since(enterT).Hours())
	//duration := nowT.Sub(enterT).Hours()
	if dur/24 < day {
		return false
	}
	return true
}

//  通过最后修改时间计算币龄收益
func caculateAwards(stub shim.ChaincodeStubInterface, balance uint64, lastModifyTime string) uint64 {
	endTime := StrToTime(lastModifyTime)
	//  获取保证金年利率
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return 0
	}
	depositRateFloat64 := cp.DepositRate
	//  计算币龄收益
	return award.GetAwardsWithCoins(balance, endTime, depositRateFloat64)
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
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := cp.FoundationAddress
	// 判断当前请求的是否为基金会
	if invokeAddr.String() != foundationAddress {
		log.Error("please use foundation address")
		return false
	}
	return true
}

func getNor(stub shim.ChaincodeStubInterface, invokeA string) (*NorNodBal, error) {
	b, err := stub.GetState(string(constants.DEPOSIT_NORMAL_PREFIX) + invokeA)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	nor := &NorNodBal{}
	err = json.Unmarshal(b, nor)
	if err != nil {
		return nil, err
	}
	return nor, nil
}
func saveNor(stub shim.ChaincodeStubInterface, invokeA string, nor *NorNodBal) error {
	b, err := json.Marshal(nor)
	if err != nil {
		return err
	}
	err = stub.PutState(string(constants.DEPOSIT_NORMAL_PREFIX)+invokeA, b)
	if err != nil {
		return err
	}
	return nil
}

func getNorMap(stub shim.ChaincodeStubInterface) (map[string]*modules.AmountAsset, error) {
	b, err := stub.GetState(NormalNodeList)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	norMap := make(map[string]*modules.AmountAsset)
	err = json.Unmarshal(b, &norMap)
	if err != nil {
		return nil, err
	}
	return norMap, nil
}

func saveNorMap(stub shim.ChaincodeStubInterface, norMap map[string]*modules.AmountAsset) error {
	b, err := json.Marshal(norMap)
	if err != nil {
		return err
	}
	err = stub.PutState(NormalNodeList, b)
	if err != nil {
		return err
	}
	return nil
}

func getExtPtn(stub shim.ChaincodeStubInterface) (map[string]*extractPtn, error) {
	b, err := stub.GetState(ExtractPtnList)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	extP := make(map[string]*extractPtn)
	err = json.Unmarshal(b, &extP)
	if err != nil {
		return nil, err
	}
	return extP, nil
}

func saveExtPtn(stub shim.ChaincodeStubInterface, extPtnL map[string]*extractPtn) error {
	b, err := json.Marshal(extPtnL)
	if err != nil {
		return err
	}
	err = stub.PutState(ExtractPtnList, b)
	if err != nil {
		return err
	}
	return nil
}

func getPledgeVotes(stub shim.ChaincodeStubInterface) (int64, error) {
	b, err := stub.GetState(AllPledgeVotes)
	if err != nil {
		return 0, err
	}
	votes, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0, err
	}
	return votes, nil
}

func savePledgeVotes(stub shim.ChaincodeStubInterface, votes int64) error {
	str := strconv.FormatInt(votes, 10)
	err := stub.PutState(AllPledgeVotes, []byte(str))
	if err != nil {
		return err
	}
	return nil
}
