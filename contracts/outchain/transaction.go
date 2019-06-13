package outchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/palletone/adaptor"
	"github.com/palletone/btc-adaptor"
	"github.com/palletone/go-palletone/common/log"

	"github.com/palletone/eth-adaptor"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

func ProcessOutChainTransaction(chaincodeID string, outChainTX *pb.OutChainTransaction) (string, error) {
	var params OutChainMethod
	err := json.Unmarshal(outChainTX.Params, &params)
	if err != nil {
		return "", fmt.Errorf("Get Request error zxl ==== ==== %s", err.Error())
	}
	log.Debug(modName, "Get Request method zxl ==== ==== ", params.Method)

	outChainTX.OutChainName = strings.ToLower(outChainTX.OutChainName)
	switch outChainTX.OutChainName {
	case "btc":
		return processTransactionMethodBTC(chaincodeID, outChainTX, &params)
	case "eth":
		return processTransactionMethodETH(chaincodeID, outChainTX, &params)
	}

	return "", errors.New("Unspport out chain.")
}

func processTransactionMethodBTC(chaincodeID string, outChainTX *pb.OutChainTransaction,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	case "SignTransaction":
		var signTransactionParams adaptor.SignTransactionParams
		err := json.Unmarshal(outChainTX.Params, &signTransactionParams)
		if err != nil {
			return "", fmt.Errorf("SignTransaction params error : %s", err.Error())
		}
		prikey, err := GetJuryBTCPrikeyTest(chaincodeID)
		if err != nil {
			return "", err
		}
		signTransactionParams.Privkeys = append(signTransactionParams.Privkeys, prikey)
		//log.Debug(modName, "SignTransaction Privkeys ==== ==== ", signTransactionParams.Privkeys)

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.SignTransaction(&signTransactionParams)
	case "SendTransactionHttp":
		var sendTransactionParams adaptor.SendTransactionHttpParams
		err := json.Unmarshal(outChainTX.Params, &sendTransactionParams)
		if err != nil {
			return "", fmt.Errorf("SendTransactionHttp params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.SendTransactionHttp(&sendTransactionParams)
	case "SignTxSend":
		var signTxSendParams adaptor.SignTxSendParams
		err := json.Unmarshal(outChainTX.Params, &signTxSendParams)
		if err != nil {
			return "", fmt.Errorf("SignTransaction params error : %s", err.Error())
		}
		prikey, err := GetJuryBTCPrikeyTest(chaincodeID)
		if err != nil {
			return "", err
		}
		signTxSendParams.Privkeys = append(signTxSendParams.Privkeys, prikey)
		log.Debug(modName, "SignTxSend Privkeys ==== ==== ", signTxSendParams.Privkeys)

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		btcAdaptor.Host = cfg.Ada.Btc.Host
		btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
		btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
		btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
		return btcAdaptor.SignTxSend(&signTxSendParams)

	case "DecodeRawTransaction":
		var decodeRawTransactionParams adaptor.DecodeRawTransactionParams
		err := json.Unmarshal(outChainTX.Params, &decodeRawTransactionParams)
		if err != nil {
			return "", fmt.Errorf("DecodeRawTransaction params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.Host = cfg.Ada.Btc.Host
		btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
		btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
		btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
		return btcAdaptor.DecodeRawTransaction(&decodeRawTransactionParams)

	case "GetTransactionByHash":
		var getTransactionByHashParams adaptor.GetTransactionByHashParams
		err := json.Unmarshal(outChainTX.Params, &getTransactionByHashParams)
		if err != nil {
			return "", fmt.Errorf("GetTransactionByHash params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.Host = cfg.Ada.Btc.Host
		btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
		btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
		btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
		return btcAdaptor.GetTransactionByHash(&getTransactionByHashParams)
	case "GetTransactionHttp":
		var getTransactionByHashParams adaptor.GetTransactionHttpParams
		err := json.Unmarshal(outChainTX.Params, &getTransactionByHashParams)
		if err != nil {
			return "", fmt.Errorf("GetTransactionHttp params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.GetTransactionHttp(&getTransactionByHashParams)
	case "GetTransactions":
		var getTransactions adaptor.GetTransactionsParams
		err := json.Unmarshal(outChainTX.Params, &getTransactions)
		if err != nil {
			return "", fmt.Errorf("GetTransactionByHash params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		btcAdaptor.Host = cfg.Ada.Btc.Host
		btcAdaptor.RPCUser = cfg.Ada.Btc.RPCUser
		btcAdaptor.RPCPasswd = cfg.Ada.Btc.RPCPasswd
		btcAdaptor.CertPath = cfg.Ada.Btc.CertPath
		return btcAdaptor.GetTransactions(&getTransactions)
	case "RawTransactionGen":
		var rawTransaction adaptor.RawTransactionGenParams
		err := json.Unmarshal(outChainTX.Params, &rawTransaction)
		if err != nil {
			return "", fmt.Errorf("RawTransactionGen params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.RawTransactionGen(&rawTransaction)
	case "VerifyMessage":
		var verifyMessage adaptor.VerifyMessageParams
		err := json.Unmarshal(outChainTX.Params, &verifyMessage)
		if err != nil {
			return "", fmt.Errorf("VerifyMessage params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.VerifyMessage(&verifyMessage)
	case "MergeTransaction":
		var mergeTransaction adaptor.MergeTransactionParams
		err := json.Unmarshal(outChainTX.Params, &mergeTransaction)
		if err != nil {
			return "", fmt.Errorf("MergeTransaction params error : %s", err.Error())
		}

		var btcAdaptor adaptorbtc.AdaptorBTC
		btcAdaptor.NetID = cfg.Ada.Btc.NetID
		return btcAdaptor.MergeTransaction(&mergeTransaction)
	}

	return "", errors.New("Unspport out chain Transaction method.")
}

func processTransactionMethodETH(chaincodeID string, outChainTX *pb.OutChainTransaction,
	params *OutChainMethod) (string, error) {
	switch params.Method {
	case "Keccak256HashPackedSig":
		var sigParams adaptor.Keccak256HashPackedSigParams
		err := json.Unmarshal(outChainTX.Params, &sigParams)
		if err != nil {
			return "", fmt.Errorf("Keccak256HashPackedSig params error : %s", err.Error())
		}
		prikey, err := GetJuryETHPrikeyTest(chaincodeID)
		if err != nil {
			return "", err
		}
		sigParams.PrivateKeyHex = prikey
		log.Debug(modName, "Keccak256HashPackedSig Privkeys ==== ==== ", sigParams.PrivateKeyHex)

		var ethAdaptor adaptoreth.AdaptorETH
		ethAdaptor.NetID = cfg.Ada.Eth.NetID
		return ethAdaptor.Keccak256HashPackedSig(&sigParams)
	case "RecoverAddr":
		var recoverParams adaptor.RecoverParams
		err := json.Unmarshal(outChainTX.Params, &recoverParams)
		if err != nil {
			return "", fmt.Errorf("RecoverAddr params error : %s", err.Error())
		}
		var ethAdaptor adaptoreth.AdaptorETH
		ethAdaptor.NetID = cfg.Ada.Eth.NetID
		return ethAdaptor.RecoverAddr(&recoverParams)
	case "GetEventByAddress":
		var getEventByAddressParams adaptor.GetEventByAddressParams
		err := json.Unmarshal(outChainTX.Params, &getEventByAddressParams)
		if err != nil {
			return "", fmt.Errorf("GetEventByAddress params error : %s", err.Error())
		}

		var ethAdaptor adaptoreth.AdaptorETH
		ethAdaptor.NetID = cfg.Ada.Eth.NetID
		ethAdaptor.Rawurl = cfg.Ada.Eth.Rawurl
		return ethAdaptor.GetEventByAddress(&getEventByAddressParams)
	}

	return "", errors.New("Unspport out chain Transaction method.")
}
