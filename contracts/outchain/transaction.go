package outchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	//"github.com/palletone/go-palletone/adaptor/btc-adaptor"
	"github.com/palletone/go-palletone/common/log"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

func ProcessOutChainTransaction(chaincodeID string, outChainAddr *pb.OutChainTransaction) (string, error) {
	var params OutChainMethod
	err := json.Unmarshal(outChainAddr.Params, &params)
	if err != nil {
		return "", fmt.Errorf("Get Request error zxl ==== ==== ", err.Error())
	}
	log.Debug(modName, "Get Request method zxl ==== ==== ", params.Method)

	outChainAddr.OutChainName = strings.ToLower(outChainAddr.OutChainName)
	switch outChainAddr.OutChainName {
	case "btc":
		return processTransactionMethodBTC(chaincodeID, outChainAddr, &params)
	case "eth":
	}

	return "", errors.New("Unspport out chain.")
}

func processTransactionMethodBTC(chaincodeID string, outChainAddr *pb.OutChainTransaction,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	//case "SignTransaction":
	//	var signTransactionParams adaptorbtc.SignTransactionParams
	//	err := json.Unmarshal(outChainAddr.Params, &signTransactionParams)
	//	if err != nil {
	//		return "", fmt.Errorf("GetBalance params error : %s", err.Error())
	//	}
	//	prikey, err := GetJuryBTCPrikeyTest(chaincodeID)
	//	if err != nil {
	//		return "", err
	//	}
	//	signTransactionParams.Privkeys = append(signTransactionParams.Privkeys, prikey)
	//	log.Debug(modName, "SignTransaction Privkeys ==== ==== ", signTransactionParams.Privkeys)
	//
	//	var btcAdaptor adaptorbtc.AdaptorBTC
	//	btcAdaptor.NetID = cfg.Ada.Btc.NetID
	//	btcAdaptor.Host = cfg.Ada.Btc.Host
	//	btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
	//	btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
	//	btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
	//	return btcAdaptor.SignTransaction(&signTransactionParams)
	}

	return "", errors.New("Unspport out chain Transaction method.")
}
