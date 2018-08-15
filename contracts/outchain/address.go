package outchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	//"github.com/palletone/go-palletone/adaptor/btc-adaptor"
	//"github.com/palletone/go-palletone/adaptor/eth-adaptor"
	"github.com/palletone/go-palletone/common/log"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

func ProcessOutChainAddress(chaincodeID string, outChainAddr *pb.OutChainAddress) (string, error) {
	var params OutChainMethod
	err := json.Unmarshal(outChainAddr.Params, &params)
	if err != nil {
		return "", fmt.Errorf("Get Request error zxl ==== ==== ", err.Error())
	}
	log.Debug(modName, "Get Request method zxl ==== ==== ", params.Method)

	outChainAddr.OutChainName = strings.ToLower(outChainAddr.OutChainName)
	switch outChainAddr.OutChainName {
	case "btc":
		return processAddressMethodBTC(chaincodeID, outChainAddr, &params)
	case "eth":
		return processAddressMethodETH(chaincodeID, outChainAddr, &params)
	}

	return "", errors.New("Unspport out chain.")
}

func processAddressMethodBTC(chaincodeID string, outChainAddr *pb.OutChainAddress,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	case "GetMultiAddr":
		//var createMultiSigParams adaptorbtc.CreateMultiSigParams
		//err := json.Unmarshal(outChainAddr.Params, &createMultiSigParams)
		//if err != nil {
		//	return "", fmt.Errorf("GetMultiAddr params error : %s", err.Error())
		//}
		//log.Debug(modName, "GetMultiAddr PublicKeys ==== ==== ", createMultiSigParams.PublicKeys)
		//
		//needJuryPubkeys := createMultiSigParams.N - len(createMultiSigParams.PublicKeys)
		//if needJuryPubkeys > 0 {
		//	pubkeys, err := ClolletJuryBTCPubkeysTest(chaincodeID)
		//	if err != nil {
		//		return "", err
		//	}
		//	if len(pubkeys) == 0 || needJuryPubkeys > len(pubkeys) {
		//		return "", adaptorbtc.NewError("Collect Jury Pubkeys error.")
		//	}
		//	for i := 0; i < needJuryPubkeys; i++ {
		//		createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, pubkeys[i])
		//	}
		//} else {
		//	return "", errors.New("params N error or Jury Pubkeys be set.")
		//}
		//
		//log.Debug(modName, "GetMultiAddr PublicKeys ==== ==== ", createMultiSigParams.PublicKeys)
		//var btcAdaptor adaptorbtc.AdaptorBTC
		//btcAdaptor.NetID = cfg.Ada.Btc.NetID
		//return btcAdaptor.CreateMultiSigAddress(&createMultiSigParams)
	}

	return "", errors.New("Unspport out chain Address method.")
}

func processAddressMethodETH(chaincodeID string, outChainAddr *pb.OutChainAddress,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	//case "GetMultiAddr":
	//	var createMultiSigAddressParams adaptoreth.CreateMultiSigAddressParams
	//	err := json.Unmarshal(outChainAddr.Params, &createMultiSigAddressParams)
	//	if err != nil {
	//		return "", fmt.Errorf("GetMultiAddr params error : %s", err.Error())
	//	}
	//	log.Debug(modName, "GetMultiAddr PublicKeys ==== ==== ", createMultiSigAddressParams.Addresses)
	//
	//	needJuryAddresses := createMultiSigAddressParams.N - len(createMultiSigAddressParams.Addresses)
	//	if needJuryAddresses > 0 {
	//		addresses, err := ClolletJuryETHAddressesTest(chaincodeID)
	//		if err != nil {
	//			return "", err
	//		}
	//		if len(addresses) == 0 || needJuryAddresses > len(addresses) {
	//			return "", adaptorbtc.NewError("Collect Jury Pubkeys error.")
	//		}
	//		for i := 0; i < needJuryAddresses; i++ {
	//			createMultiSigAddressParams.Addresses = append(createMultiSigAddressParams.Addresses, addresses[i])
	//		}
	//	} else {
	//		return "", errors.New("params N error or Jury Addresses be set.")
	//	}
	//
	//	log.Debug(modName, "GetMultiAddr PublicKeys ==== ==== ", createMultiSigAddressParams.Addresses)
	//	var ethAdaptor adaptoreth.AdaptorETH
	//	return ethAdaptor.CreateMultiSigAddress(&createMultiSigAddressParams)
	}

	return "", errors.New("Unspport out chain Address method.")
}
