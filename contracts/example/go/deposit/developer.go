package deposit

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"strings"
	"time"
)

func developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("GetInvokeFromAddr error:")
	}
	//交付数量
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Success([]byte("GetPayToContractPtnTokens error:"))
	}
	//获取账户
	balance, err := stub.GetDepositBalance(invokeAddr)
	if err != nil {
		return shim.Error(err.Error())
	}
	isDeveloper := false
	if balance == nil {
		balance = &modules.DepositBalance{}
		if invokeTokens.Amount >= depositAmountsForDeveloper {
			//加入列表
			//addList("Jury", invokeAddr, stub)
			err = addCandaditeList(invokeAddr, stub, "DeveloperList")
			if err != nil {
				return shim.Error(err.Error())
			}
			isDeveloper = true
			balance.EnterTime = time.Now().UTC()
		}
		updateForPayValue(balance, invokeTokens)
	} else {
		//账户已存在，进行信息的更新操作
		if balance.TotalAmount >= depositAmountsForDeveloper {
			//原来就是jury
			isDeveloper = true
			//TODO 再次交付保证金时，先计算当前余额的币龄奖励
			awards := award.GetAwardsWithCoins(balance.TotalAmount, balance.LastModifyTime.Unix())
			balance.TotalAmount += awards

		}
		//处理交付保证金数据
		updateForPayValue(balance, invokeTokens)
	}
	if !isDeveloper {
		//判断交了保证金后是否超过了jury
		if balance.TotalAmount >= depositAmountsForDeveloper {
			//addList("Jury", invokeAddr, stub)
			err = addCandaditeList(invokeAddr, stub, "DeveloperList")
			if err != nil {
				return shim.Error(err.Error())
			}
			balance.EnterTime = time.Now().UTC()
		}
	}
	err = marshalAndPutStateForBalance(stub, invokeAddr, balance)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("developer pay ok."))
}

func developerApplyCashback(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	err := applyCashbackList("Developer", stub, args)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("apply for cashback success."))
}

//基金会处理
func handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//地址，申请时间，是否同意
	if len(args) != 3 {
		return shim.Error("Input parameter error,need three parameters.")
	}
	//基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//判断没收请求地址是否是基金会地址
	if strings.Compare(invokeAddr, foundationAddress) != 0 {
		return shim.Error("请求地址不正确，请使用基金会的地址")
	}
	//获取一下该用户下的账簿情况
	addr := args[0]
	balance, err := stub.GetDepositBalance(addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//判断没收节点账户是否为空
	if balance == nil {
		return shim.Error("you have not depositWitnessPay for deposit.")
	}
	//获取申请时间戳
	strTime := args[1]
	applyTime, err := strconv.ParseInt(strTime, 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	check := args[2]
	if check == "ok" {
		//对余额处理
		err = handleDeveloper(stub, addr, applyTime, balance)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr, applyTime)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	return shim.Success([]byte("handle for cashback success."))
}

func handleDeveloper(stub shim.ChaincodeStubInterface, cashbackAddr string, applyTime int64, balance *modules.DepositBalance) error {
	//获取请求列表
	listForCashback, err := stub.GetListForCashback()
	if err != nil {
		return err
	}
	if listForCashback == nil {
		return fmt.Errorf("%s", "listForCashback is nil.")
	}
	isExist := isInCashbacklist(cashbackAddr, listForCashback)
	if !isExist {
		return fmt.Errorf("%s", "node is not exist in the list.")
	}
	//获取节点信息
	cashbackNode := &modules.Cashback{}
	for _, m := range listForCashback {
		if m.CashbackAddress == cashbackAddr && m.CashbackTime == applyTime {
			cashbackNode = m
			break
		}
	}
	newList := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
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
	//还得判断一下是否超过余额
	if cashbackNode.CashbackTokens.Amount > balance.TotalAmount {
		return fmt.Errorf("%s", "退款大于账户余额")
	}
	err = handleDeveloperDepositCashback(stub, cashbackAddr, cashbackNode, balance)
	if err != nil {
		return err
	}
	return nil
}

//对Developer退保证金的处理
func handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *modules.Cashback, balance *modules.DepositBalance) error {
	if balance.TotalAmount >= depositAmountsForJury {
		//已在列表中
		err := handleDeveloperFromList(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			return err
		}
	} else {
		////TODO 不在列表中,没有奖励，直接退
		err := handleCommonJuryOrDev(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			return err
		}
	}
	return nil
}

//Developer已在列表中
func handleDeveloperFromList(stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *modules.Cashback, balance *modules.DepositBalance) error {
	//退出列表
	var err error
	//计算余额
	result := balance.TotalAmount - cashbackValue.CashbackTokens.Amount
	//判断是否退出列表
	if result == 0 {
		//加入列表时的时间
		startTime := balance.EnterTime.YearDay()
		//当前退出时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已到期
		if endTime-startTime >= depositPeriod {
			//退出全部，即删除cashback，利息计算好了
			err = cashbackAllDeposit("Developer", stub, cashbackAddr, cashbackValue.CashbackTokens, balance)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s", "未到期，不能退出列表")
		}
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中，还没有计算利息
		//d.addListForCashback("Jury", stub, cashbackAddr, invokeTokens)
		err = cashbackSomeDeposit("Developer", stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			return err
		}
	}
	return nil
}
