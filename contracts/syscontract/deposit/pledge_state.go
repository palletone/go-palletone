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

package deposit

import (
	"encoding/json"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

//质押相关的状态数据库操作，包括增加质押，质押分红，质押列表查询，质押提现等
func savePledgeRecord(stub shim.ChaincodeStubInterface, prefix string, node *modules.AddressAmount) error {
	b, err := json.Marshal(node)
	if err != nil {
		return err
	}
	err = stub.PutState(prefix+node.Address, b)
	if err != nil {
		return err
	}
	return nil
}

func getPledgeRecord(stub shim.ChaincodeStubInterface, prefix string, addr string) (*modules.AddressAmount, error) {
	b, err := stub.GetState(prefix + addr)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	node := &modules.AddressAmount{}
	err = json.Unmarshal(b, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}
func getAllPledgeRecords(stub shim.ChaincodeStubInterface, prefix string) ([]*modules.AddressAmount, error) {
	// 增加新的质押
	awards, err := stub.GetStateByPrefix(prefix)
	if err != nil {
		return nil, err
	}
	result := []*modules.AddressAmount{}

	for _, a := range awards {
		awardNode := &modules.AddressAmount{}
		err = json.Unmarshal(a.Value, awardNode)
		if err != nil {
			return nil, err
		}
		result = append(result, awardNode)
	}
	return result, nil
}
func savePledgeDepositRecord(stub shim.ChaincodeStubInterface, node *modules.AddressAmount) error {
	return savePledgeRecord(stub, string(constants.PLEDGE_DEPOSIT_PREFIX), node)
}
func delPledgeDepositRecord(stub shim.ChaincodeStubInterface, addr string) error {
	key := string(constants.PLEDGE_DEPOSIT_PREFIX) + addr
	return stub.DelState(key)
}
func getPledgeDepositRecord(stub shim.ChaincodeStubInterface, addr string) (*modules.AddressAmount, error) {
	addrAmt,err:= getPledgeRecord(stub, string(constants.PLEDGE_DEPOSIT_PREFIX), addr)
	if err!=nil{
		log.Error("getPledgeDepositRecord by %s return error:%s",addr,err.Error())
		return nil,err
	}
	if addrAmt!=nil {
		log.Debugf("getPledgeDepositRecord by %s,result:%d", addr, addrAmt.Amount)
	}
	return addrAmt,err
}
func getAllPledgeDepositRecords(stub shim.ChaincodeStubInterface) ([]*modules.AddressAmount, error) {
	return getAllPledgeRecords(stub, string(constants.PLEDGE_DEPOSIT_PREFIX))
}
func savePledgeWithdrawRecord(stub shim.ChaincodeStubInterface, node *modules.AddressAmount) error {
	return savePledgeRecord(stub, string(constants.PLEDGE_WITHDRAW_PREFIX), node)
}
func delPledgeWithdrawRecord(stub shim.ChaincodeStubInterface, addr string) error {
	key := string(constants.PLEDGE_WITHDRAW_PREFIX) + addr
	return stub.DelState(key)
}
func getPledgeWithdrawRecord(stub shim.ChaincodeStubInterface, addr string) (*modules.AddressAmount, error) {
	addrAmt,err:=  getPledgeRecord(stub, string(constants.PLEDGE_WITHDRAW_PREFIX), addr)
	if err!=nil{
		log.Error("getPledgeWithdrawRecord by %s return error:%s",addr,err.Error())
		return nil,err
	}
	if addrAmt!=nil {
		log.Debugf("getPledgeWithdrawRecord by %s,result:%d", addr, addrAmt.Amount)
	}
	return addrAmt,err
}
func getAllPledgeWithdrawRecords(stub shim.ChaincodeStubInterface) ([]*modules.AddressAmount, error) {
	return getAllPledgeRecords(stub, string(constants.PLEDGE_WITHDRAW_PREFIX))
}

//获得质押列表的最后更新日期yyyyMMdd
func getLastPledgeListDate(stub shim.ChaincodeStubInterface) (string, error) {
	date, err := stub.GetState(constants.PledgeListLastDate)
	if err != nil {
		return "", err
	}
	if date == nil {
		return "", nil
	}
	return string(date), nil
}
func saveLastPledgeListDate(stub shim.ChaincodeStubInterface, date string) error {
	return stub.PutState(constants.PledgeListLastDate, []byte(date))

}

//保存最新的质押列表
func saveLastPledgeList(stub shim.ChaincodeStubInterface, allM *modules.PledgeList) error {
	b, err := json.Marshal(allM)
	if err != nil {
		return err
	}
	err = stub.PutState(constants.PledgeList+allM.Date, b)
	if err != nil {
		return err
	}
	return saveLastPledgeListDate(stub, allM.Date)

}

//获得最新的质押列表
func getLastPledgeList(stub shim.ChaincodeStubInterface) (*modules.PledgeList, error) {
	date, err := getLastPledgeListDate(stub)
	if err != nil {
		return nil, err
	}
	return getPledgeListByDate(stub,date)
}
func getPledgeListByDate(stub shim.ChaincodeStubInterface,date string) (*modules.PledgeList, error) {
	b, err := stub.GetState(constants.PledgeList + date)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	allM := &modules.PledgeList{}
	err = json.Unmarshal(b, allM)
	if err != nil {
		return nil, err
	}
	return allM, nil
}
//查询历史上的所有质押列表记录
func getAllPledgeRewardHistory(stub shim.ChaincodeStubInterface) ([]*modules.PledgeList, error) {
	b, err := stub.GetStateByPrefix(constants.PledgeList)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	result := []*modules.PledgeList{}
	for _, kv := range b {
		allM := &modules.PledgeList{}
		err = json.Unmarshal(kv.Value, allM)
		if err != nil {
			return nil, err
		}
		result = append(result, allM)
	}

	return result, nil
}
