package outchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"

	"github.com/palletone/adaptor"
)

func TestGetJuryKeyInfo(t *testing.T) {
	input := adaptor.NewPrivateKeyInput{}
	inputJSON, err := json.Marshal(&input)
	fmt.Println(string(inputJSON))

	key, err := GetJuryKeyInfo("sample_syscc", "erc20", inputJSON, GetERC20Adaptor())
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("prikey: %x\n", key)
	}
}

func TestUnmarshal(t *testing.T) {
	outputBytes := `{"transaction":{"tx_id":"86c4920a8698a5aadaf9f5eedd45efdedbb924cb59dab4a46231a2d8286039c6","tx_raw":"a9059cbb000000000000000000000000a840d94b1ef4c326c370e84d108d539d31d52e840000000000000000000000000000000000000000000000056bc75e2d63100000","creator_address":"0x588eB98f8814aedB056D549C0bafD5Ef4963069C","target_address":"0xa54880Da9A63cDD2DdAcF25aF68daF31a1bcC0C9","is_in_block":true,"is_success":true,"is_stable":true,"block_id":"b551b9a7c0f168d7509f67b39a955a6f41ab32ba1d394651bd8e6f460dc70062","block_height":6234506,"tx_index":0,"timestamp":0,"from_address":"0x588eB98f8814aedB056D549C0bafD5Ef4963069C","to_address":"0xa840d94B1ef4c326C370e84D108D539d31D52e84","amount":{"amount":100000000000000000000,"asset":""},"fee":{"amount":54606,"asset":""},"attach_data":"\ufffd\u0005\ufffd\ufffd\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\ufffd@\ufffdK\u001e\ufffd\ufffd\u0026\ufffdp\ufffdM\u0010\ufffdS\ufffd1\ufffd.\ufffd\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0005k\ufffd^-c\u0010\u0000\u0000"},"extra":null}`

	//outputBytes := {"transaction":{"tx_id":"86c4920a8698a5aadaf9f5eedd45efdedbb924cb59dab4a46231a2d8286039c6","tx_raw":"a9059cbb000000000000000000000000a840d94b1ef4c326c370e84d108d539d31d52e840000000000000000000000000000000000000000000000056bc75e2d63100000","creator_address":"0x588eB98f8814aedB056D549C0bafD5Ef4963069C","target_address":"0xa54880Da9A63cDD2DdAcF25aF68daF31a1bcC0C9","is_in_block":true,"is_success":true,"is_stable":true,"block_id":"b551b9a7c0f168d7509f67b39a955a6f41ab32ba1d394651bd8e6f460dc70062","block_height":6234506,"tx_index":0,"timestamp":0,"from_address":"0x588eB98f8814aedB056D549C0bafD5Ef4963069C","to_address":"0xa840d94B1ef4c326C370e84D108D539d31D52e84","amount":{"amount":100000000000000000000,"asset":""},"fee":{"amount":54606,"asset":""},"attach_data":"\ufffd\u0005\ufffd\ufffd\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\ufffd@\ufffdK\u001e\ufffd\ufffd\u0026\ufffdp\ufffdM\u0010\ufffdS\ufffd1\ufffd.\ufffd\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0005k\ufffd^-c\u0010\u0000\u0000"},"extra":null}
	var output adaptor.GetTransferTxOutput
	err := json.Unmarshal([]byte(outputBytes), &output)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println("OK")
	}
}

func TestGetTransferTx(t *testing.T) {
	txID, _ := hex.DecodeString("86c4920a8698a5aadaf9f5eedd45efdedbb924cb59dab4a46231a2d8286039c6")
	input := adaptor.GetTransferTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(string(inputBytes))
	}

	outChainCall := &pb.OutChainCall{OutChainName: "erc20", Method: "GetTransferTx", Params: inputBytes}
	result, err := ProcessOutChainCall("sample_syscc", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(result)
	}
	outputBytes := []byte(result)

	var output adaptor.GetTransferTxOutput
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println("OK")
	}
}

func TestGetPalletOneMappingAddress(t *testing.T) {
	input := adaptor.GetPalletOneMappingAddressInput{ChainAddress: "0x588eb98f8814aedb056d549c0bafd5ef4963069c", MappingDataSource: "0xa840d94b1ef4c326c370e84d108d539d31d52e84"}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(string(inputBytes))
	}

	outChainCall := &pb.OutChainCall{OutChainName: "erc20", Method: "GetPalletOneMappingAddress", Params: inputBytes}
	result, err := ProcessOutChainCall("sample_syscc", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(result)
		var output adaptor.GetPalletOneMappingAddressOutput
		err = json.Unmarshal([]byte(result), &output)
		if err != nil {
			fmt.Println("err: ", err)
			return
		}
		if output.PalletOneAddress == "" {
			fmt.Println("PalletOneAddress empty")
			return
		}

		fmt.Println(output.PalletOneAddress)
	}
}

func TestGetBlockInfo(t *testing.T) {
	//
	input := adaptor.GetBlockInfoInput{Latest: true} //get best hight
	//
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return
	}

	outChainCall := &pb.OutChainCall{OutChainName: "erc20", Method: "GetBlockInfo", Params: inputBytes}
	result, err := ProcessOutChainCall("sample_syscc", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(result)
	}
}

func TestGetAddrTxHistory(t *testing.T) {
	//
	input := adaptor.GetAddrTxHistoryInput{FromAddress: "0x588eb98f8814aedb056d549c0bafd5ef4963069c",
		ToAddress: "0xa840d94b1ef4c326c370e84d108d539d31d52e84", Asset: "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9",
		AddressLogicAndOr: true} //
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return
	}

	outChainCall := &pb.OutChainCall{OutChainName: "erc20", Method: "GetAddrTxHistory", Params: inputBytes}
	result, err := ProcessOutChainCall("sample_syscc", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(result)
		var output adaptor.GetAddrTxHistoryOutput
		err = json.Unmarshal([]byte(result), &output)
		if err != nil {
			fmt.Println("err: ", err)
			return
		}
		for _, txResult := range output.Txs {
			txIDHex := hex.EncodeToString(txResult.TxID)
			fmt.Println(txIDHex)
		}
	}
}
