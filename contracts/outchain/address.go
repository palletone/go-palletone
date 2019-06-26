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

func ProcessOutChainAddress(chaincodeID string, outChainAddr *pb.OutChainAddress) (string, error) {
	var params OutChainMethod
	err := json.Unmarshal(outChainAddr.Params, &params)
	if err != nil {
		return "", fmt.Errorf("Get Request error zxl ==== ==== %s", err.Error())
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
	case "CreateMultiSigAddress":
		var createMultiSigParams adaptor.CreateMultiSigParams
		err := json.Unmarshal(outChainAddr.Params, &createMultiSigParams)
		if err != nil {
			return "", fmt.Errorf("CreateMultiSigAddress params error : %s", err.Error())
		}
		log.Debug(modName, "CreateMultiSigAddress PublicKeys ==== ==== ", createMultiSigParams.PublicKeys)

		needJuryPubkeys := createMultiSigParams.N - len(createMultiSigParams.PublicKeys)
		if needJuryPubkeys > 0 {
			pubkeys, err := ClolletJuryBTCPubkeysTest(chaincodeID)
			if err != nil {
				return "", err
			}
			if len(pubkeys) == 0 || needJuryPubkeys > len(pubkeys) {
				return "", errors.New("Collect Jury Pubkeys error.")
			}
			for i := 0; i < needJuryPubkeys; i++ {
				createMultiSigParams.PublicKeys = append(createMultiSigParams.PublicKeys, pubkeys[i])
			}
		}

		log.Debug(modName, "CreateMultiSigAddress PublicKeys ==== ==== ", createMultiSigParams.PublicKeys)
		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		//
		result, err := btcAdaptor.CreateMultiSigAddress(&createMultiSigParams)
		if err != nil {
			return "", err
		}

		return result, nil

	case "GetAddressByPubkey":
		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.GetAddressByPubkey(string(outChainAddr.Params))
	case "GetJuryBTCPubkey":
		pubkeys, err := ClolletJuryBTCPubkeysTest(chaincodeID)
		if err != nil {
			return "", err
		}
		if len(pubkeys) == 0 {
			return "", errors.New("Collect Jury Pubkeys error.")
		}
		return pubkeys[0], nil
	}

	return "", errors.New("Unspport out chain Address method.")
}

type GetJuryAddressParams struct {
	Addresses []string `json:"addresses"`
	N         int      `json:"n"`
	M         int      `json:"m"`
}

func processAddressMethodETH(chaincodeID string, outChainAddr *pb.OutChainAddress,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	case "CreateMultiSigAddress":
		var createMultiSigAddressParams adaptor.CreateMultiSigAddressParams
		err := json.Unmarshal(outChainAddr.Params, &createMultiSigAddressParams)
		if err != nil {
			return "", fmt.Errorf("CreateMultiSigAddress params error : %s", err.Error())
		}
		log.Debug(modName, "CreateMultiSigAddress Address ==== ==== ", createMultiSigAddressParams.Addresses)

		needJuryAddresses := createMultiSigAddressParams.N - len(createMultiSigAddressParams.Addresses)
		if needJuryAddresses > 0 {
			addresses, err := ClolletJuryETHAddressesTest(chaincodeID)
			if err != nil {
				return "", err
			}
			if len(addresses) == 0 || needJuryAddresses > len(addresses) {
				return "", errors.New("Collect Jury Address error.")
			}
			for i := 0; i < needJuryAddresses; i++ {
				createMultiSigAddressParams.Addresses = append(createMultiSigAddressParams.Addresses, addresses[i])
			}
		} else {
			return "", errors.New("params N error or Jury Addresses not be reserved.")
		}

		log.Debug(modName, "CreateMultiSigAddress Address ==== ==== ", createMultiSigAddressParams.Addresses)
		var ethAdaptor adaptoreth.AdaptorETH
		return ethAdaptor.CreateMultiSigAddress(&createMultiSigAddressParams)

	case "GetJuryETHAddr":
		addrs, err := ClolletJuryETHAddressesTest(chaincodeID)
		if err != nil {
			return "", err
		}
		if len(addrs) == 0 {
			return "", errors.New("Collect Jury address error.")
		}
		return addrs[0], nil
	}

	return "", errors.New("Unspport out chain Address method.")
}
