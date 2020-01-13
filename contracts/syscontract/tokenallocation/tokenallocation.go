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
 *  * @date 2018-2020
 *
 */

package tokenallocation

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type TokenAllocation struct {
}

const PacketPrefix = "Packet-"

func (p *TokenAllocation) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *TokenAllocation) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "createPacket": //创建红包
		if len(args) != 6 {
			return shim.Error("must input 6 args: pubKeyHex, packetCount,minAmt,maxAmt,expiredTime,remark")
		}
		pubKey, err := hex.DecodeString(args[0])
		if err != nil {
			return shim.Error("Invalid pub key string:" + args[0])
		}
		count, err := strconv.Atoi(args[1])
		if err != nil {
			return shim.Error("Invalid packet count string:" + args[1])
		}
		minAmount, err := decimal.NewFromString(args[2])
		if err != nil {
			return shim.Error("Invalid min amount string:" + args[2])
		}
		maxAmount, err := decimal.NewFromString(args[3])
		if err != nil {
			return shim.Error("Invalid max amount string:" + args[3])
		}
		var exp *time.Time = nil
		if args[4] != "" && args[4] != "0" {
			ti, err := time.ParseInLocation("2006-01-02 15:04:05", args[4], time.Local)
			if err != nil {
				return shim.Error("Invalid expired time format[YYYYmmdd HH:MM:ss]:" + args[4])
			}
			exp = &ti
		}
		err = p.CreatePacket(stub, pubKey, count, minAmount, maxAmount, exp, args[5])
		if err != nil {
			return shim.Error("CreatePacket error:" + err.Error())
		}
		return shim.Success(nil)
	case "pullPacket": //领取红包

		result, err := p.PullPacket(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getPacketBalance": //红包余额
		result, err := p.GetBlacklistAddress(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getPacketAllocation": //红包领取记录
		result, err := p.GetBlacklistAddress(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "updatePacket": //调整红包的参数
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
	case "recyclePacket": //红包过期，需要回收未被领取的Token
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
func (p *TokenAllocation) CreatePacket(stub shim.ChaincodeStubInterface, pubKey []byte, count int,
	minAmount, maxAmount decimal.Decimal, expiredTime *time.Time, remark string) error {
	creator, _ := stub.GetInvokeAddress()
	tokenToPackets, err := stub.GetInvokeTokens()
	if err != nil {
		return err
	}
	if len(tokenToPackets) != 1 {
		return errors.New("Please pay one kind of token to this contract.")
	}
	tokenToPacket := tokenToPackets[0]
	_, err = getPacket(stub, pubKey)
	if p != nil {
		return errors.New("PubKey already exist")
	}
	packet := &Packet{
		PubKey:          pubKey,
		Creator:         creator,
		Token:           tokenToPacket.Asset,
		Amount:          tokenToPacket.Amount,
		Count:           uint32(count),
		MinPacketAmount: tokenToPacket.Asset.Uint64Amount(minAmount),
		MaxPacketAmount: tokenToPacket.Asset.Uint64Amount(maxAmount),
		Remark:          remark,
	}
	if expiredTime == nil {
		packet.ExpiredTime = 0
	} else {
		packet.ExpiredTime = uint64(expiredTime.Unix())
	}
	err = savePacket(stub, packet)
	if err != nil {
		return err
	}
	return nil
}

func (p *TokenAllocation) PullPacket(stub shim.ChaincodeStubInterface, pubKey []byte, msg string, signature []byte,
	pullAddr common.Address) error {
	packet, err := getPacket(stub, pubKey)
	if err != nil {
		return errors.New("Packet not found")
	}
	currentTime, _ := stub.GetTxTimestamp(10)
	if packet.ExpiredTime != 0 && packet.ExpiredTime > uint64(currentTime.Seconds) {
		return errors.New("Packet already expired")
	}
	pass, err := crypto.MyCryptoLib.Verify(pubKey, signature, []byte(msg))
	if err != nil || !pass {
		return errors.New("validate signature failed")
	}
	//验证通过，发送红包
	payAmt := packet.GetPullAmount(nil)
	return stub.PayOutToken(pullAddr.String(), &modules.AmountAsset{
		Amount: payAmt,
		Asset:  packet.Token,
	}, 0)
}
func (p *TokenAllocation) GetBlacklistAddress(stub shim.ChaincodeStubInterface) ([]common.Address, error) {
	return getBlacklistAddress(stub)
}
func (p *TokenAllocation) Payout(stub shim.ChaincodeStubInterface, addr common.Address, amount decimal.Decimal, asset *modules.Asset) error {
	if !isFoundationInvoke(stub) {
		return errors.New("only foundation address can call this function")
	}
	uint64Amt := ptnjson.JsonAmt2AssetAmt(asset, amount)
	return stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: uint64Amt,
		Asset:  asset,
	}, 0)
}
func (p *TokenAllocation) QueryIsInBlacklist(stub shim.ChaincodeStubInterface, addr common.Address) (bool, error) {
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
