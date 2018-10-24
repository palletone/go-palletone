package outchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/palletone/adaptor"
	"github.com/palletone/btc-adaptor"
	"github.com/palletone/eth-adaptor"
	"github.com/palletone/go-palletone/common/log"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

func ProcessOutChainQuery(chaincodeID string, outChainAddr *pb.OutChainQuery) (string, error) {
	var params OutChainMethod
	err := json.Unmarshal(outChainAddr.Params, &params)
	if err != nil {
		return "", fmt.Errorf("Get Request error zxl ==== ==== %s", err.Error())
	}
	log.Debug(modName, "Get Request method zxl ==== ==== ", params.Method)

	outChainAddr.OutChainName = strings.ToLower(outChainAddr.OutChainName)
	switch outChainAddr.OutChainName {
	case "btc":
		return processQueryMethodBTC(chaincodeID, outChainAddr, &params)
	case "eth":
		return processQueryMethodETH(chaincodeID, outChainAddr, &params)
	}

	return "", errors.New("Unspport out chain.")
}

func processQueryMethodBTC(chaincodeID string, outChainAddr *pb.OutChainQuery,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	case "GetBalance":
		var getBalanceParams adaptor.GetBalanceParams
		err := json.Unmarshal(outChainAddr.Params, &getBalanceParams)
		if err != nil {
			return "", fmt.Errorf("GetBalance params error : %s", err.Error())
		}
		log.Debug(modName, "GetBalance Address ==== ==== ", getBalanceParams.Address)

		err1 := GetConfigTest()
		if err != nil {
			log.Error("loadconfig() failed !!!!!!")
			return "", err1
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		btcAdaptor.Host = cfg.Ada.Btc.Host
		btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
		btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
		btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
		return btcAdaptor.GetBalance(&getBalanceParams)
	}
	return "", errors.New("Unspport out chain Query method.")
}

func processQueryMethodETH(chaincodeID string, outChainAddr *pb.OutChainQuery,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	case "GetBalance":
		var getHeaderParams adaptor.GetBestHeaderParams
		err := json.Unmarshal(outChainAddr.Params, &getHeaderParams)
		if err != nil {
			return "", fmt.Errorf("GetBalance params error : %s", err.Error())
		}

		err1 := GetConfigTest()
		if err != nil {
			log.Error("loadconfig() failed !!!!!!")
			return "", err1
		}

		var ethAdaptor adaptoreth.AdaptorETH
		ethAdaptor.Rawurl = cfg.Ada.Eth.Rawurl
		return ethAdaptor.GetBestHeader(&getHeaderParams)
	}
	return "", errors.New("Unspport out chain Query method.")
}
