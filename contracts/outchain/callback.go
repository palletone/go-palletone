package outchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/palletone/adaptor"
	"github.com/palletone/btc-adaptor"
	"github.com/palletone/eth-adaptor"

	"github.com/palletone/go-palletone/common/log"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

var (
	allChain      = map[string]Callback{"btc": GetBTCAdaptor, "eth": GetETHAdaptor}
	exceptMethond = map[string]map[string]CallbackExcpet{ //some method need private key, list in map, first key is chainName
		"btc": {"GetJuryBTCPubkey": GetJuryBTCPubkey, "SignTransaction": SignTransaction},
		"eth": {"GetJuryETHAddr": GetJuryETHAddr, "Keccak256HashPackedSig": Keccak256HashPackedSig},
	}
)

type Callback func() interface{}

type CallbackExcpet func(chaincodeID, chainName string, params []byte) (string, error)

func GetBTCAdaptor() interface{} {
	var btcAdaptor adaptorbtc.AdaptorBTC
	btcAdaptor.NetID = cfg.Ada.Btc.NetID
	btcAdaptor.Host = cfg.Ada.Btc.Host //
	btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
	btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
	btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
	return &btcAdaptor
}
func GetETHAdaptor() interface{} {
	var ethAdaptor adaptoreth.AdaptorETH
	ethAdaptor.NetID = cfg.Ada.Eth.NetID
	ethAdaptor.Rawurl = cfg.Ada.Eth.Rawurl
	return &ethAdaptor
}

func ProcessOutChainCall(chaincodeID string, outChainCall *pb.OutChainCall) (string, error) {
	log.Debugf("Get Request method : %s", outChainCall.Method)

	chainName := strings.ToLower(outChainCall.OutChainName)
	if _, existChain := exceptMethond[chainName]; existChain {
		ef, existMethod := exceptMethond[chainName][outChainCall.Method]
		if existMethod {
			return ef(chaincodeID, chainName, outChainCall.Params)
		}
	}

	return adaptorCall(chainName, outChainCall.Method, outChainCall.Params)
}
func GetJuryBTCPubkey(chaincodeID string, methodName string, params []byte) (string, error) {
	pubkeys, err := ClolletJuryBTCPubkeysTest(chaincodeID)
	if err != nil {
		return "", err
	}
	if len(pubkeys) == 0 {
		return "", errors.New("Collect Jury Pubkeys error.")
	}
	return pubkeys[0], nil
}
func SignTransaction(chaincodeID string, methodName string, params []byte) (string, error) {
	var signTransactionParams adaptor.SignTransactionParams
	err := json.Unmarshal(params, &signTransactionParams)
	if err != nil {
		return "", fmt.Errorf("SignTransaction params error : %s", err.Error())
	}
	prikey, err := GetJuryBTCPrikeyTest(chaincodeID)
	if err != nil {
		return "", err
	}
	signTransactionParams.Privkeys = append(signTransactionParams.Privkeys, prikey)

	var btcAdaptor adaptorbtc.AdaptorBTC
	btcAdaptor.NetID = cfg.Ada.Btc.NetID
	return btcAdaptor.SignTransaction(&signTransactionParams)
}
func GetJuryETHAddr(chaincodeID string, methodName string, params []byte) (string, error) {
	addrs, err := ClolletJuryETHAddressesTest(chaincodeID)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", errors.New("Collect Jury address error.")
	}
	return addrs[0], nil
}

func Keccak256HashPackedSig(chaincodeID string, methodName string, params []byte) (string, error) {
	var sigParams adaptor.Keccak256HashPackedSigParams
	err := json.Unmarshal(params, &sigParams)
	if err != nil {
		return "", fmt.Errorf("Keccak256HashPackedSig params error : %s", err.Error())
	}
	prikey, err := GetJuryETHPrikeyTest(chaincodeID)
	if err != nil {
		return "", err
	}
	sigParams.PrivateKeyHex = prikey

	var ethAdaptor adaptoreth.AdaptorETH
	ethAdaptor.NetID = cfg.Ada.Eth.NetID
	return ethAdaptor.Keccak256HashPackedSig(&sigParams)
}

func adaptorCall(chainName, methodName string, params []byte) (string, error) {
	f, has := allChain[chainName]
	if !has {
		log.Debugf("Not implement this Chain")
		return "", errors.New("Not implement this Chain")
	}
	adaptorObj := f()
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
			return res[0].String(), nil
		} else {
			return "", errors.New(res[1].String()) //
		}
	} else {
		return res[0].String(), nil
	}
}
