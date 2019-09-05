package outchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/palletone/go-palletone/common/log"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"

	"github.com/palletone/adaptor"
	//"github.com/palletone/btc-adaptor"
	"github.com/palletone/eth-adaptor"
)

var (
	allChain      = map[string]adaptor.ICryptoCurrency{"btc": GetBTCAdaptor(), "eth": GetETHAdaptor(), "erc20": GetERC20Adaptor()}
	exceptMethond = map[string]CallbackExcpet{ //some method need private key, list in map, first key is methodName
		"GetJuryPubkey": GetJuryPubkey, "GetJuryAddr": GetJuryAddr,
		"SignTransaction": SignTransaction, "SignMessage": SignMessage,
	}
)

type Callback func() adaptor.ICryptoCurrency

type CallbackExcpet func(chaincodeID, chainName string, params []byte) (string, error)

func GetBTCAdaptor() adaptor.ICryptoCurrency {
	//var btcAdaptor adaptorbtc.AdaptorBTC
	//btcAdaptor.NetID = cfg.Ada.Btc.NetID
	//btcAdaptor.Host = cfg.Ada.Btc.Host //
	//btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
	//btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
	//btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
	//return &btcAdaptor
	return nil
}
func GetETHAdaptor() adaptor.ICryptoCurrency {
	var ethAdaptor ethadaptor.AdaptorETH
	ethAdaptor.NetID = cfg.Ada.Eth.NetID
	ethAdaptor.Rawurl = cfg.Ada.Eth.Rawurl
	ethAdaptor.TxQueryUrl = cfg.Ada.Eth.TxQueryUrl
	return &ethAdaptor
}
func GetERC20Adaptor() adaptor.ICryptoCurrency {
	var ethAdaptor ethadaptor.AdaptorErc20
	ethAdaptor.NetID = cfg.Ada.Eth.NetID
	ethAdaptor.Rawurl = cfg.Ada.Eth.Rawurl
	ethAdaptor.TxQueryUrl = cfg.Ada.Eth.TxQueryUrl
	return &ethAdaptor
}

func ProcessOutChainCall(chaincodeID string, outChainCall *pb.OutChainCall) (string, error) {
	log.Debugf("Get Request method : %s", outChainCall.Method)

	chainName := strings.ToLower(outChainCall.OutChainName)
	if _, existChain := exceptMethond[chainName]; existChain {
		ef, existMethod := exceptMethond[outChainCall.Method]
		if existMethod {
			return ef(chaincodeID, chainName, outChainCall.Params)
		}
	}

	return adaptorCall(chainName, outChainCall.Method, outChainCall.Params)
}
func GetJuryPubkey(chaincodeID string, chainName string, params []byte) (string, error) {
	adaptorObj := allChain[chainName] //ProcessOutChainCall has checked
	priKey, err := GetJuryKeyInfo(chaincodeID, chainName, params, adaptorObj)
	if err != nil {
		return "", err
	}
	ouptPub, err := adaptorObj.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: priKey})
	if err != nil {
		return "", err
	}
	resultJson, err := json.Marshal(ouptPub)
	if err != nil {
		return "", err
	}
	return string(resultJson), nil
}

func SignTransaction(chaincodeID string, chainName string, params []byte) (string, error) {
	adaptorObj := allChain[chainName] //ProcessOutChainCall has checked
	//
	priKey, err := GetJuryKeyInfo(chaincodeID, chainName, params, adaptorObj)
	if err != nil {
		return "", err
	}
	//
	var input adaptor.SignTransactionInput
	input.PrivateKey = priKey

	//
	result, err := adaptorObj.SignTransaction(&input)
	if err != nil {
		return "", err
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(resultJson), nil
}
func GetJuryAddr(chaincodeID string, chainName string, params []byte) (string, error) {
	adaptorObj := allChain[chainName] //ProcessOutChainCall has checked
	addr, err := GetJuryAddress(chaincodeID, chainName, params, adaptorObj)
	if err != nil {
		return "", err
	}
	return addr, nil
}

func SignMessage(chaincodeID string, chainName string, params []byte) (string, error) {
	adaptorObj := allChain[chainName] //ProcessOutChainCall has checked
	//
	priKey, err := GetJuryKeyInfo(chaincodeID, chainName, params, adaptorObj)
	if err != nil {
		return "", err
	}
	//
	var input adaptor.SignMessageInput
	input.PrivateKey = priKey

	//
	result, err := adaptorObj.SignMessage(&input)
	if err != nil {
		return "", err
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(resultJson), nil
}

func adaptorCall(chainName, methodName string, params []byte) (string, error) {
	adaptorObj, has := allChain[chainName]
	if !has {
		log.Debugf("Not implement this Chain")
		return "", errors.New("Not implement this Chain")
	}
	//adaptorObj := f()
	adaptorObjValue := reflect.ValueOf(adaptorObj)
	adaptorObjMethod := adaptorObjValue.MethodByName(methodName)
	if !adaptorObjMethod.IsValid() {
		log.Debugf("Not exist this Method")
		return "", errors.New("Not exist this Method")
	}
	log.Debugf("method parms's num %d", adaptorObjMethod.Type().NumIn())
	if adaptorObjMethod.Type().NumIn() > 1 {
		log.Debugf("Implement of method %s is invalid, only support one params", methodName)
		return "", errors.New("Implement of method is invalid, only support one params")
	}

	var res []reflect.Value
	if 0 == adaptorObjMethod.Type().NumIn() {
		res = adaptorObjMethod.Call(nil)
	} else {
		inParaType := adaptorObjMethod.Type().In(0)
		if inParaType.Kind() == reflect.Ptr {
			rvf := reflect.New(inParaType.Elem())
			err := json.Unmarshal(params, rvf.Interface())
			if nil != err {
				log.Debugf("Method's param is invalid")
				return "", errors.New("Method's param is invalid")
			}
			params := []reflect.Value{rvf}
			res = adaptorObjMethod.Call(params)
		} else if inParaType.Kind() == reflect.String {
			targetValue := reflect.ValueOf(params)
			params := []reflect.Value{targetValue}
			res = adaptorObjMethod.Call(params)
		} else {
			log.Debugf("Implement of method %s is invalid", methodName)
			return "", errors.New("Implement of method is invalid")
		}
	}
	if len(res) > 1 {
		if res[1].IsNil() {
			//return res[0].String(), nil
			result, err := json.Marshal(res[0].Interface())
			return string(result), err
		} else {
			return "", fmt.Errorf("%s", res[1].Interface())
		}
	} else {
		return res[0].String(), nil
	}
}
