package exchangecc

import (
	"encoding/json"
	"errors"
	"strconv"
	"bytes"
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

type ExchangeMgr struct {
	Address     common.Address
	sale        *modules.Asset 
	sale_amount        int
	want        *modules.Asset 
	want_amount        int
	exchange_sn []byte
}
const EXCHANGELIST_RECORD = "Exchangelist-"
type ExchangelistRecord struct {
	Address     common.Address
	sale        *modules.Asset 
	sale_amount        int
	want        *modules.Asset 
	want_amount        int
	exchange_sn []byte
}
func saveRecord(stub shim.ChaincodeStubInterface, record *ExchangeMgr) error {
	data, _ := rlp.EncodeToBytes(record)
	return stub.PutState(EXCHANGELIST_RECORD+record.Address.String(), data)
}
func (p *ExchangeMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *ExchangeMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "maker": //create a maker
		if len(args) != 5 {
			return shim.Error("must input 2 args: blackAddress, reason")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want, err := modules.StringToAsset(args[1])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want_amount, err := strconv.Atoi(args[2])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale, err := modules.StringToAsset(args[3])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale_amount, err := strconv.Atoi(args[4])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		newsheet := ExchangeMgr{}
		newsheet.Address = addr
		newsheet.sale = sale
		newsheet.sale_amount = sale_amount
		newsheet.want = want
		newsheet.want_amount = want_amount
		txid := stub.GetTxID()
	    newsheet.exchange_sn = common.Hex2Bytes(txid[2:])
		err = p.AddExchangelist(stub, &newsheet)
		if err != nil {
			return shim.Error("AddBlacklist error:" + err.Error())
		}

		return shim.Success(nil)
	case "taker": //获取挂单信息,token互换
		if len(args) != 5 {
			return shim.Error("must input 2 args: blackAddress, reason")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		addr = addr
		want, err := modules.StringToAsset(args[1])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want=want
		want_amount, err := strconv.Atoi(args[2])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want_amount=want_amount
		sale, err := modules.StringToAsset(args[3])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale=sale
		sale_amount, err := strconv.Atoi(args[4])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale_amount=sale_amount
		//todo  check if have maker want token
		//todo  send to maker contract addr  want token  
		    return shim.Success(nil)
	case "getExchangelist": //列出黑名单地址列表
		result, err := p.GetExchangelistRecords(stub)
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
	case "QueryIsInExchangelist": //判断某地址是否在黑名单中
		if len(args) != 1 {
			return shim.Error("must input 1 args: Address")
		}
		exchange_sn := []byte(args[0])
		
		result, err := p.QueryIsInExchangelist(stub, exchange_sn)
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
func getExchangeRecords(stub shim.ChaincodeStubInterface) ([]*ExchangelistRecord, error) {
	kvs, err := stub.GetStateByPrefix(EXCHANGELIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*ExchangelistRecord, 0, len(kvs))
	for _, kv := range kvs {
		record := &ExchangelistRecord{}
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

func (p *ExchangeMgr) GetExchangelistRecords(stub shim.ChaincodeStubInterface) ([]*ExchangelistRecord, error) {
	return getExchangeRecords(stub)
}
func (p *ExchangeMgr) GetExchangelist(stub shim.ChaincodeStubInterface) ([]*ExchangeMgr, error) {
	return getExchangelistAddress(stub)
}
func (p *ExchangeMgr) Payout(stub shim.ChaincodeStubInterface, addr common.Address, amount decimal.Decimal, asset *modules.Asset) error {
	if !isFoundationInvoke(stub) {
		return errors.New("only foundation address can call this function")
	}
	uint64Amt := ptnjson.JsonAmt2AssetAmt(asset, amount)
	return stub.PayOutToken(addr.String(), &modules.AmountAsset{
		Amount: uint64Amt,
		Asset:  asset,
	}, 0)
}
func (p *ExchangeMgr) QueryIsInExchangelist(stub shim.ChaincodeStubInterface, exchange_sn []byte) (bool, error) {
	exchangelist, err := getExchangelistAddress(stub)
	if err != nil {
		return false, err
	}
	for _, b := range exchangelist {
		if bytes.Equal(b.exchange_sn,exchange_sn) {
			return true, nil
		}
	}
	return false, nil
}
func updateExchangeAddressList(stub shim.ChaincodeStubInterface, sheet *ExchangeMgr) error {
	list, _ := getExchangelistAddress(stub)
	list = append(list, sheet)
	data, _ := rlp.EncodeToBytes(list)
	return stub.PutState(constants.ExchangelistAddress, data)
}

func (p *ExchangeMgr) AddExchangelist(stub shim.ChaincodeStubInterface, sheet *ExchangeMgr) error {
	if !isFoundationInvoke(stub) {
		return errors.New("only foundation address can call this function")
	}//单号
	addr := sheet.Address
	exist, _ := p.QueryIsInExchangelist(stub, sheet.exchange_sn)
	if exist { //不可重复添加同一个交易单
		return errors.New(addr.String() + " already exist in exchangelist")
	}
	tokenBalance, err := stub.GetTokenBalance(addr.String(), nil)
	if err != nil {
		return errors.New("GetTokenBalance error:" + err.Error())
	}
	balance := make(map[modules.Asset]uint64)
	for _, aa := range tokenBalance {
		balance[*aa.Asset] = aa.Amount
	}
	record := &ExchangeMgr{
		Address  :  sheet.Address,
	    sale     :  sheet.sale,
	    sale_amount:sheet.sale_amount,
	    want     :  sheet.want,
	    want_amount:sheet.want_amount,
	    exchange_sn:sheet.exchange_sn,
	}
	err = saveRecord(stub, record)
	if err != nil {
		return errors.New("saveRecord error:" + err.Error())
	}
	err = updateExchangeAddressList(stub, record)
	if err != nil {
		return errors.New("updateExchangeAddressList error:" + err.Error())
	}
	return nil
}


func getExchangelistAddress(stub shim.ChaincodeStubInterface) ([]*ExchangeMgr, error) {
	list := []*ExchangeMgr{} 
	dblist, err := stub.GetState(constants.ExchangelistAddress)
	if err == nil && len(dblist) > 0 {
		err = rlp.DecodeBytes(dblist, list)
		if err != nil {
			log.Errorf("rlp decode data[%x] to  []common.Address error", dblist)
			return nil, errors.New("rlp decode error:" + err.Error())
		}
	}
	return list, nil
}