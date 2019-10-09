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

package blacklistcc

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
)

type BlacklistMgr struct {
}

func (p *BlacklistMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *BlacklistMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "addBlacklist": //增加一个地址到黑名单
		if len(args) != 2 {
			return shim.Error("must input 2 args: blackAddress, reason")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		err = p.AddBlacklist(stub, addr, args[1])
		if err != nil {
			return shim.Error("AddBlacklist error:" + err.Error())
		}
		return shim.Success(nil)
	case "getBlacklistRecords": //列出黑名单列表
		result, err := p.GetBlacklistRecords(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getBlacklistAddress": //列出黑名单地址列表
		result, err := p.GetBlacklistAddress(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "payout": //付出Token
		if len(args) != 3 {
			return shim.Error("must input 3 args: Address,Amount,Asset")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		amount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid amount:" + args[1])
		}
		asset, err := modules.StringToAsset(args[2])
		if err != nil {
			return shim.Error("Invalid asset string:" + args[2])
		}
		err = p.Payout(stub, addr, amount, asset)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "queryIsInBlacklist": //判断某地址是否在黑名单中
		if len(args) != 1 {
			return shim.Error("must input 1 args: Address")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		result, err := p.QueryIsInBlacklist(stub, addr)
		if err != nil {
			return shim.Error("QueryIsInBlacklist error:" + err.Error())
		}
		if result {
			return shim.Success([]byte("true"))
		} else {
			return shim.Success([]byte("false"))
		}
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}
func (p *BlacklistMgr) AddBlacklist(stub shim.ChaincodeStubInterface, blackAddr common.Address, reason string) error {
	if !isFoundationInvoke(stub) {
		return errors.New("only foundation address can call this function")
	}
	exist, _ := p.QueryIsInBlacklist(stub, blackAddr)
	if exist { //不可重复添加同一个地址到黑名单
		return errors.New(blackAddr.String() + " already exist in blacklist")
	}
	tokenBalance, err := stub.GetTokenBalance(blackAddr.String(), nil)
	if err != nil {
		return errors.New("GetTokenBalance error:" + err.Error())
	}
	balance := make(map[modules.Asset]uint64)
	for _, aa := range tokenBalance {
		balance[*aa.Asset] = aa.Amount
	}
	balanceJson, _ := json.Marshal(balance)
	record := &BlacklistRecord{
		Address:     blackAddr,
		Reason:      reason,
		FreezeToken: string(balanceJson),
	}
	err = saveRecord(stub, record)
	if err != nil {
		return errors.New("saveRecord error:" + err.Error())
	}
	err = updateBlacklistAddressList(stub, blackAddr)
	if err != nil {
		return errors.New("updateBlacklistAddressList error:" + err.Error())
	}
	//发行对应冻结的Token给合约
	_, addr := stub.GetContractID()
	for asset, amount := range balance {
		err = stub.SupplyToken(asset.AssetId[:], asset.UniqueId[:], amount, addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *BlacklistMgr) GetBlacklistRecords(stub shim.ChaincodeStubInterface) ([]*BlacklistRecord, error) {
	return getAllRecords(stub)
}
func (p *BlacklistMgr) GetBlacklistAddress(stub shim.ChaincodeStubInterface) ([]common.Address, error) {
	return getBlacklistAddress(stub)
}
func (p *BlacklistMgr) Payout(stub shim.ChaincodeStubInterface, addr common.Address, amount decimal.Decimal, asset *modules.Asset) error {
	if !isFoundationInvoke(stub) {
		return errors.New("only foundation address can call this function")
	}
	uint64Amt := ptnjson.JsonAmt2AssetAmt(asset, amount)
	return stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: uint64Amt,
		Asset:  asset,
	}, 0)
}
func (p *BlacklistMgr) QueryIsInBlacklist(stub shim.ChaincodeStubInterface, addr common.Address) (bool, error) {
	blacklist, err := getBlacklistAddress(stub)
	if err != nil {
		return false, err
	}
	for _, b := range blacklist {
		if b.Equal(addr) {
			return true, nil
		}
	}
	return false, nil
}

type BlacklistRecord struct {
	Address     common.Address
	Reason      string
	FreezeToken string
}

const BLACKLIST_RECORD = "Blacklist-"

func saveRecord(stub shim.ChaincodeStubInterface, record *BlacklistRecord) error {
	data, _ := rlp.EncodeToBytes(record)
	return stub.PutState(BLACKLIST_RECORD+record.Address.String(), data)
}
func getAllRecords(stub shim.ChaincodeStubInterface) ([]*BlacklistRecord, error) {
	kvs, err := stub.GetStateByPrefix(BLACKLIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*BlacklistRecord, 0, len(kvs))
	for _, kv := range kvs {
		record := &BlacklistRecord{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
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
func updateBlacklistAddressList(stub shim.ChaincodeStubInterface, address common.Address) error {
	list, _ := getBlacklistAddress(stub)
	list = append(list, address)
	data, _ := rlp.EncodeToBytes(list)
	return stub.PutState(constants.BlacklistAddress, data)
}
func getBlacklistAddress(stub shim.ChaincodeStubInterface) ([]common.Address, error) {
	list := []common.Address{}
	dblist, err := stub.GetState(constants.BlacklistAddress)
	if err == nil && len(dblist) > 0 {
		err = rlp.DecodeBytes(dblist, &list)
		if err != nil {
			log.Errorf("rlp decode data[%x] to  []common.Address error", dblist)
			return nil, errors.New("rlp decode error:" + err.Error())
		}
	}
	return list, nil
}
