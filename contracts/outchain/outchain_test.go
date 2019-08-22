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
		fmt.Println("err1: ", err)
	} else {
		fmt.Println(result)
	}
}
