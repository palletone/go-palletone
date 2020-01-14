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

package packetcc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/palletone/go-palletone/common/crypto"
	"sort"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

type PacketMgr struct {
}

const PacketPrefix = "P-"
const PacketBalancePrefix = "B-"
const PacketAllocationRecordPrefix = "R-"

func (p *PacketMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *PacketMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "createPacket": //创建红包
		fallthrough
	case "updatePacket": //调整红包的参数
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
		if f == "createPacket" {
			err = p.CreatePacket(stub, pubKey, count, minAmount, maxAmount, exp, args[5])
		} else { //update
			err = p.UpdatePacket(stub, pubKey, count, minAmount, maxAmount, exp, args[5])
		}
		if err != nil {
			return shim.Error("CreatePacket error:" + err.Error())
		}
		return shim.Success(nil)
	case "pullPacket": //领取红包
		if len(args) != 4 {
			return shim.Error("must input 4 args: pubKeyHex, message,signature,receiveAddress")
		}

		pubKey, err := hex.DecodeString(args[0])
		if err != nil {
			return shim.Error("Invalid pub key string:" + args[0])
		}
		message := args[1]
		//signature := []byte("signature")
		signature, err := hex.DecodeString(args[2])
		if err != nil {
			return shim.Error("Invalid signature hex string:" + args[2])
		}
		address, err := common.StringToAddress(args[3])
		if err != nil {
			return shim.Error("Invalid address string:" + args[3])
		}
		err = p.PullPacket(stub, pubKey, message, signature, address)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "getPacketInfo": //红包余额等信息
		if len(args) != 1 {
			return shim.Error("must input 1 args: pubKeyHex")
		}

		pubKey, err := hex.DecodeString(args[0])
		if err != nil {
			return shim.Error("Invalid pub key string:" + args[0])
		}
		result, err := p.GetPacketInfo(stub, pubKey)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "getPacketAllocationHistory": //红包领取记录
		if len(args) != 1 {
			return shim.Error("must input 1 args: pubKeyHex")
		}

		pubKey, err := hex.DecodeString(args[0])
		if err != nil {
			return shim.Error("Invalid pub key string:" + args[0])
		}
		result, err := p.GetPacketAllocationHistory(stub, pubKey)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	case "recyclePacket": //红包过期，需要回收未被领取的Token
		if len(args) != 1 {
			return shim.Error("must input 1 args: pubKeyHex")
		}

		pubKey, err := hex.DecodeString(args[0])
		if err != nil {
			return shim.Error("Invalid pub key string:" + args[0])
		}
		err = p.RecyclePacket(stub, pubKey)
		if err != nil {
			return shim.Error("QueryIsInBlacklist error:" + err.Error())
		}
		return shim.Success(nil)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}
func (p *PacketMgr) CreatePacket(stub shim.ChaincodeStubInterface, pubKey []byte, count int,
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
	pk, _ := getPacket(stub, pubKey)
	if pk != nil {
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
	err = savePacketBalance(stub, pubKey, packet.Amount, packet.Count)
	if err != nil {
		return err
	}
	return nil
}

//增加额度，调整红包产生等
func (p *PacketMgr) UpdatePacket(stub shim.ChaincodeStubInterface, pubKey []byte, count int,
	minAmount, maxAmount decimal.Decimal, expiredTime *time.Time, remark string) error {
	creator, _ := stub.GetInvokeAddress()
	packet, err := getPacket(stub, pubKey)
	if err != nil {
		return err
	}
	if packet.Creator != creator && !isFoundationInvoke(stub) {
		return errors.New("Only creator or admin can update")
	}
	adjustCount := int32(count) - int32(packet.Count)
	packet.Count = uint32(count)
	packet.MinPacketAmount = packet.Token.Uint64Amount(minAmount)
	packet.MaxPacketAmount = packet.Token.Uint64Amount(maxAmount)
	packet.Remark = remark
	if expiredTime == nil {
		packet.ExpiredTime = 0
	} else {
		packet.ExpiredTime = uint64(expiredTime.Unix())
	}
	tokenToPackets, err := stub.GetInvokeTokens()
	if err != nil {
		return err
	}
	if len(tokenToPackets) == 0 { //只调整参数，不增加额度
		err = savePacket(stub, packet)
		if err != nil {
			return err
		}
		return nil
	}

	if len(tokenToPackets) != 1 {
		return errors.New("Please pay one kind of token to this contract.")
	}
	tokenToPacket := tokenToPackets[0]
	if !tokenToPacket.Asset.Equal(packet.Token) {
		return errors.New("Please pay " + packet.Token.String() + " to this contract.")
	}
	packet.Amount += tokenToPacket.Amount

	err = savePacket(stub, packet)
	if err != nil {
		return err
	}
	bAmount, bCount, err := getPacketBalance(stub, pubKey)
	if err != nil {
		return err
	}
	newCount := int32(bCount) + adjustCount
	if newCount < 0 {
		return errors.New(fmt.Sprintf("Count must >=%d", bCount))
	}
	err = savePacketBalance(stub, pubKey, bAmount+tokenToPacket.Amount, uint32(newCount))
	if err != nil {
		return err
	}
	return nil
}
func (p *PacketMgr) PullPacket(stub shim.ChaincodeStubInterface,
	pubKey []byte, msg string, signature []byte,
	pullAddr common.Address) error {
	packet, err := getPacket(stub, pubKey)
	if err != nil {
		return errors.New("Packet not found")
	}
	//检查红包是否过期
	currentTime, _ := stub.GetTxTimestamp(10)
	if packet.ExpiredTime != 0 && packet.ExpiredTime <= uint64(currentTime.Seconds) {
		return errors.New("Packet already expired")
	}
	pass, err := crypto.MyCryptoLib.Verify(pubKey, signature, []byte(msg))
	if err != nil || !pass {
		return errors.New("validate signature failed")
	}
	balanceAmount, balanceCount, err := getPacketBalance(stub, packet.PubKey)
	if err != nil {
		return err
	}
	if balanceCount == 0 {
		return errors.New("Packet count is zero")
	}
	if balanceAmount == 0 {
		return errors.New("Packet balance is zero")
	}
	//验证通过，发送红包
	hash := common.HexToHash(stub.GetTxID())
	seed := util.BytesToUInt64(hash[0:8])
	payAmt := packet.GetPullAmount(int64(seed), balanceAmount, balanceCount)
	err = stub.PayOutToken(pullAddr.String(), &modules.AmountAsset{
		Amount: payAmt,
		Asset:  packet.Token,
	}, 0)
	if err != nil {
		return err
	}
	//调整红包余额
	err = savePacketBalance(stub, packet.PubKey, balanceAmount-payAmt, balanceCount-1)
	if err != nil {
		return err
	}
	//保存红包领取记录
	reqId := common.HexToHash(stub.GetTxID())
	timestamp, _ := stub.GetTxTimestamp(10)
	record := &PacketAllocationRecord{
		PubKey:      pubKey,
		Message:     msg,
		Amount:      payAmt,
		Token:       packet.Token,
		ToAddress:   pullAddr,
		RequestHash: reqId,
		Timestamp:   uint64(timestamp.Seconds),
	}
	err = savePacketAllocationRecord(stub, record)
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
func (p *PacketMgr) GetPacketInfo(stub shim.ChaincodeStubInterface, pubKey []byte) (*PacketJson, error) {
	packet, err := getPacket(stub, pubKey)
	if err != nil {
		return nil, err
	}
	balanceAmount, balanceCount, err := getPacketBalance(stub, pubKey)
	if err != nil {
		return nil, err
	}
	return convertPacket2Json(packet, balanceAmount, balanceCount), nil
}
func (p *PacketMgr) GetPacketAllocationHistory(stub shim.ChaincodeStubInterface,
	pubKey []byte) ([]*PacketAllocationRecordJson, error) {
	records, err := getPacketAllocationHistory(stub, pubKey)
	if err != nil {
		return nil, err
	}
	result := make([]*PacketAllocationRecordJson, len(records))
	for i, record := range records {
		result[i] = convertAllocationRecord2Json(record)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})
	return result, nil
}
func (p *PacketMgr) RecyclePacket(stub shim.ChaincodeStubInterface, pubKey []byte) error {
	packet, err := getPacket(stub, pubKey)
	if err != nil {
		return err
	}
	now, _ := stub.GetTxTimestamp(10)
	if packet.ExpiredTime > uint64(now.Seconds) { //红包未过期
		return errors.New("packet not expired")
	}
	balanceAmount, _, err := getPacketBalance(stub, pubKey)
	if err != nil {
		return err
	}
	if balanceAmount == 0 {
		return errors.New("no balance to recycle")
	}
	err = stub.PayOutToken(packet.Creator.String(), &modules.AmountAsset{Amount: balanceAmount, Asset: packet.Token}, 0)
	if err != nil {
		return err
	}
	//更新余额
	return savePacketBalance(stub, pubKey, 0, 0)
}
