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
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

func juryPayToDepositContract(stub shim.ChaincodeStubInterface, pubkey string) peer.Response {
	log.Debug("Start entering JuryPayToDepositContract func")
	var jdej core.JurorDepositExtraJson
	err := json.Unmarshal([]byte(pubkey), &jdej)
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

	//  交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}

	invokeAddrStr := invokeAddr.String()
	jde, err := jdej.Validate(invokeAddrStr)
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

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

	//获取账户
	balance, err := GetJuryBalance(stub, invokeAddrStr)
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  第一次想加入
	if balance == nil {
		//  可以加入列表
		if invokeTokens.Amount != cp.DepositAmountForJury {
			str := fmt.Errorf("jury needs to pay only %d  deposit.", cp.DepositAmountForJury)
			log.Error(str.Error())
			return shim.Error(str.Error())
		}
		//  加入候选列表
		err = addCandaditeList(stub, invokeAddr, modules.JuryList)
		if err != nil {
			log.Error("addCandaditeList err: ", "error", err)
			return shim.Error(err.Error())
		}

		balance = &modules.JurorDeposit{}
		balance.EnterTime = getTime(stub)
		//  没有
		balance.Balance = invokeTokens.Amount
		balance.Role = modules.Jury
		balance.Address = invokeAddrStr
		balance.JurorDepositExtra = jde
		err = saveJuryBalance(stub, invokeAddrStr, balance)
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
		if all != cp.DepositAmountForJury {
			str := fmt.Errorf("jury needs to pay only %d  deposit.", cp.DepositAmountForJury-balance.Balance)
			log.Error(str.Error())
			return shim.Error(str.Error())
		}
		//这里需要判断是否以及被基金会提前移除候选列表，即在规定时间内该节点没有追缴保证金
		b, err := isInCandidate(stub, invokeAddrStr, modules.JuryList)
		if err != nil {
			log.Debugf("isInCandidate error: %s", err.Error())
			return shim.Error(err.Error())
		}
		if !b {
			//  加入jury候选列表
			err = addCandaditeList(stub, invokeAddr, modules.JuryList)
			if err != nil {
				log.Error("addCandidateListAndPutStateForJury err: ", "error", err)
				return shim.Error(err.Error())
			}
		}
		balance.Balance = all
		err = saveJuryBalance(stub, invokeAddrStr, balance)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
}

func juryApplyQuit(stub shim.ChaincodeStubInterface) peer.Response {
	log.Debug("JuryApplyQuit")
	err := applyQuitList(modules.Jury, stub)
	if err != nil {
		log.Error("applyQuitList err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//  处理
func handleJury(stub shim.ChaincodeStubInterface, quitAddr common.Address) error {
	return handleNode(stub, quitAddr, modules.Jury)
}

func convertJuryDeposit2Json(juror *modules.JurorDeposit) *modules.JuryDepositJson {
	jrJson := &modules.JuryDepositJson{}

	dbJson := convertDepositBalance2Json(&juror.DepositBalance)
	jrJson.DepositBalanceJson = *dbJson
	jrJson.JurorDepositExtraJson = convertJuryDepositExtra2Json(&juror.JurorDepositExtra)
	jrJson.Address = juror.Address

	return jrJson
}

func convertJuryDepositExtra2Json(extra *core.JurorDepositExtra) (json core.JurorDepositExtraJson) {
	json.PublicKey = hex.EncodeToString(extra.PublicKey)
	return
}

//func updateJuryInfo(stub shim.ChaincodeStubInterface, args []string) peer.Response {
//	addr, err := stub.GetInvokeAddress()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	b, err := getJuryBalance(stub, addr.String())
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	//  TODO
//	if len(args) != 1 {
//		return shim.Error("need 1 parameter")
//	}
//	if len(args[0]) != 68 {
//		return shim.Error("public key is error")
//	}
//	//TODO 验证公钥和地址的关系
//	byte, err := hexutil.Decode(args[0])
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	b.PublicKey = byte
//	err = saveJuryBalance(stub, addr.String(), b)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success(nil)
//}
