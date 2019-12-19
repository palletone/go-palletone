package exchangecc

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Sale        *modules.Asset 
	SaleAmount        uint64
	Want        *modules.Asset 
	WantAmount        uint64
	ExchangeSn string
}
type ExchangeSheetJson struct {
	Address     common.Address
	Sale        *modules.Asset 
	SaleAmount        decimal.Decimal
	Want        *modules.Asset 
	WantAmount        decimal.Decimal
	ExchangeSn  string
}
const EXCHANGELIST_RECORD = "Exchangelist-"
func saveRecord(stub shim.ChaincodeStubInterface, record *ExchangeMgr) error {
	data, _ := rlp.EncodeToBytes(record)
	return stub.PutState(EXCHANGELIST_RECORD/*+record.Address.String()*/+record.ExchangeSn, data)
}

func  convertSheet(exm ExchangeMgr) *ExchangeSheetJson{
	   newSheet := ExchangeSheetJson{}
	   newSheet.Address = exm.Address
	   newSheet.Sale = exm.Sale
	   newSheet.SaleAmount = decimal.New(int64(exm.SaleAmount), 0)
	   newSheet.Want = exm.Want
	   newSheet.WantAmount = decimal.New(int64(exm.WantAmount), 0)
	   newSheet.ExchangeSn = exm.ExchangeSn
	   return &newSheet
}
func (p *ExchangeMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (p *ExchangeMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "maker": //create a maker
		if len(args) != 5 {
			return shim.Error("must input 5 args: blackAddress, reason")
		}
		addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want, err := modules.StringToAsset(args[1])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want_amount, err := decimal.NewFromString(args[2])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale, err := modules.StringToAsset(args[3])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale_amount, err := decimal.NewFromString(args[4])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		newsheet := ExchangeMgr{}
		newsheet.Address = addr
		newsheet.Sale = sale
		//amount = amount.Mul(decimal.New(100000000, 0))
		newsheet.SaleAmount = uint64(sale_amount.IntPart())
		newsheet.Want = want 
		newsheet.WantAmount = uint64(want_amount.IntPart())
		txid := stub.GetTxID()
		newsheet.ExchangeSn = fmt.Sprintf("%x", common.Hex2Bytes(txid[2:]))
		err = p.AddExchangelist(stub, &newsheet)
		if err != nil {
			return shim.Error("AddExchangelist error:" + err.Error())
		}

		return shim.Success(nil)
	case "taker": //获取挂单信息,token互换
		if len(args) != 5 {
			return shim.Error("must input 2 args: AddExchangelist, reason")
		}
		taker_addr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		taker_addr=taker_addr
		want, err := modules.StringToAsset(args[1])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want=want
		want_amount, err := decimal.NewFromString(args[2])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		want_amount = want_amount
		sale, err := modules.StringToAsset(args[3])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale=sale
		sale_amount, err := decimal.NewFromString(args[4])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		sale_amount=sale_amount
		ExchangeSn, err := decimal.NewFromString(args[5])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		ExchangeSn = ExchangeSn
		//todo  check if have maker want token

		//todo  send to maker contract addr  want token  
		    return shim.Success(nil)
	case "getExchangelist": //列出黑名单地址列表
		result, err := p.GetExchangeMgrs(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)
	case "withDrawExchange": //列出黑名单地址列表
		if len(args) != 1 {
			return shim.Error("must input 1 args: exchange_sn")
		}
	    result := p.DelExchangeRecord(stub,args[0])
	    if result != true {
			return shim.Error("Invalid exchange_sn :" + args[0])
		}
		return shim.Success([]byte("1"))
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
		exchange_sn := args[0]
		
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
func getExchangeRecords(stub shim.ChaincodeStubInterface) ([]*ExchangeSheetJson, error) {
	kvs, err := stub.GetStateByPrefix(EXCHANGELIST_RECORD)
	if err != nil {
		return nil, err
	}
	result := make([]*ExchangeSheetJson, 0, len(kvs))
	for _, kv := range kvs {
		record := &ExchangeMgr{}
		err = rlp.DecodeBytes(kv.Value, record)
		if err != nil {
			return nil, err
		}
		jsSheet := convertSheet(*record)
		result = append(result, jsSheet)
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

func (p *ExchangeMgr) GetExchangeMgrs(stub shim.ChaincodeStubInterface) ([]*ExchangeSheetJson, error) {
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
func (p *ExchangeMgr) QueryIsInExchangelist(stub shim.ChaincodeStubInterface, exchange_sn string) (bool, error) {
	exchangelist, err := getExchangelistAddress(stub)
	if err != nil {
		return false, err
	}
	for _, b := range exchangelist {
		if b.ExchangeSn == exchange_sn {
			return true, nil
		}
	}
	return false, nil
}
func (p *ExchangeMgr) DelExchangeRecord(stub shim.ChaincodeStubInterface, exchange_sn string) bool {
	err := stub.DelState(EXCHANGELIST_RECORD+exchange_sn)
	if err == nil {
		return true
	}
	return false
}
func (p *ExchangeMgr) FindInExchangelist(stub shim.ChaincodeStubInterface, exchange_sn string) (*ExchangeMgr, error) {
	exchangelist, err := getExchangelistAddress(stub)
	if err != nil {
		return nil, err
	}
	for _, b := range exchangelist {
		if b.ExchangeSn == exchange_sn {
			return b, nil
		}
	}
	return nil, nil
}
func updateExchangeAddressList(stub shim.ChaincodeStubInterface, sheet *ExchangeMgr) error {
	list, _ := getExchangelistAddress(stub)
	list = append(list, sheet)
	data, _ := rlp.EncodeToBytes(list)
	return stub.PutState(constants.ExchangelistAddress, data)
}
func (p *ExchangeMgr) AddExchangelist(stub shim.ChaincodeStubInterface, sheet *ExchangeMgr) error {
	addr := sheet.Address
	exist, _ := p.QueryIsInExchangelist(stub, sheet.ExchangeSn)
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
	    Sale     :  sheet.Sale,
	    SaleAmount:sheet.SaleAmount,
	    Want     :  sheet.Want,
	    WantAmount:sheet.WantAmount,
	    ExchangeSn:sheet.ExchangeSn,
	}
	err = saveRecord(stub, record)
	if err != nil {
		return errors.New("saveRecord error:" + err.Error())
	}
	/*err = updateExchangeAddressList(stub, record)
	if err != nil {
		return errors.New("updateExchangeAddressList error:" + err.Error())
	}*/
	return nil
}


func getExchangelistAddress(stub shim.ChaincodeStubInterface) ([]*ExchangeMgr, error) {
	list := []*ExchangeMgr{} 
	dblist, err := stub.GetState(constants.ExchangelistAddress)
	if err == nil && len(dblist) > 0 {
		err = rlp.DecodeBytes(dblist, &list)
		if err != nil {
			log.Errorf("rlp decode data[%x] to  []common.Address error", dblist)
			return nil, errors.New("rlp decode error:" + err.Error())
		}
	}
	return list, nil
}
