package auctioncc

import (
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
)

//todo tmp
var  DestructionAddress = "PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX"//""PCLOST00000000000000000000000000000"  //销毁地址
var   AuctionTransactionGas = uint64(1)  //ptn  临时的

func getPayToContract(stub shim.ChaincodeStubInterface) (*modules.Asset, uint64, error) {
	payAssets, _ := stub.GetInvokeTokens()
	for _, invokeAA := range payAssets {
		if invokeAA.Address == myContractAddr {
			return invokeAA.Asset, invokeAA.Amount, nil
		}
	}
	return nil, 0, errors.New("no pay to contract token")
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
