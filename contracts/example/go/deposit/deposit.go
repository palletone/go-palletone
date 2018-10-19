package deposit

import (
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
)

type DepositChaincode struct{}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("***system contract init about DepositChaincode***")
	//获取配置文件
	depositConfigBytes, err := stub.GetDepositConfig()
	if err != nil {
		fmt.Println("deposit error: ", err.Error())
		return shim.Error(err.Error())
	}
	fmt.Println("deposit=", string(depositConfigBytes))
	return shim.Success(nil)
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "deposit_witness_pay":
		//交付保证金
		//handle witness pay
		//void deposit_witness_pay(const witness_object& wit, token_type amount)

		//获取用户地址
		userAddr, err := stub.GetPayToContractAddr()
		if err != nil {
			fmt.Println("GetPayToContractAddr error: ", err.Error())
			return shim.Error(err.Error())
		}
		fmt.Println("GetPayToContractAddr=", string(userAddr))
		//获取 Token 数量
		tokenAmount, err := stub.GetPayToContractTokens()
		if err != nil {
			fmt.Println("GetPayToContractTokens error: ", err.Error())
		}
		fmt.Println("GetPayToContractTokens=", string(tokenAmount))

		return d.deposit_witness_pay(stub, args)
	case "deposit_cashback":
		//保证金退还
		//handle cashback rewards
		//void deposit_cashback(const account_object& acct, token_type amount, bool require_vesting = true)
		return d.deposit_cashback(stub, args)
	case "forfeiture_deposit":
		//保证金没收
		//void forfeiture_deposit(const witness_object& wit, token_type amount)
		return d.forfeiture_deposit(stub, args)
	default:
		return shim.Error("Invoke error")
	}
}

func (d *DepositChaincode) deposit_witness_pay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//判断参数是否准确
	if len(args) != 2 {
		return shim.Error("input error: need two args (witnessAddr and ptnAmount)")
	}
	//将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error("ptnAccount input error: " + err.Error())
	}
	//TODO 获取一下该用户下的账簿情况
	accBalByte, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("get account balance from ledger error: " + err.Error())
	}
	if accBalByte == nil {
		stub.PutState(args[0], []byte(args[1]))
		return shim.Success([]byte("ok"))
	}
	accBalStr := string(accBalByte)
	//将 string 转 uint64
	accBal, err := strconv.ParseUint(accBalStr, 10, 64)
	if err != nil {
		return shim.Error("string parse to uint64 error: " + err.Error())
	}
	//写入写集
	result := accBal + ptnAccount
	resultStr := strconv.FormatUint(result, 10)
	stub.PutState(args[0], []byte(resultStr))
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) deposit_cashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//有可能从mediator 退出成为 jury,把金额退出一半
	if len(args) != 2 {
		return shim.Error("input error: need two args (witnessAddr and ptnAmount)")
	}
	//将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error("ptnAccount input error: " + err.Error())
	}
	//TODO 获取一下该用户下的账簿情况
	accBalByte, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("get account balance from ledger error: " + err.Error())
	}
	if accBalByte == nil {
		return shim.Error("your deposit does not exist.")
	}
	accBalStr := string(accBalByte)
	//将 string 转 uint64
	accBal, err := strconv.ParseUint(accBalStr, 10, 64)
	if err != nil {
		return shim.Error("string parse to uint64 error: " + err.Error())
	}
	if accBal-ptnAccount < 0 {
		return shim.Error("deposit does not enough.")
	}
	//写入写集
	result := accBal - ptnAccount
	resultStr := strconv.FormatUint(result, 10)
	stub.PutState(args[0], []byte(resultStr))
	return shim.Success([]byte("ok"))
}

func (d DepositChaincode) forfeiture_deposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("input error: only need one arg (witnessAddr)")
	}
	//直接把保证金没收
	stub.PutState(args[0], []byte("0"))
	return shim.Success([]byte("ok"))
}
