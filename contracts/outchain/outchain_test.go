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

	key, err := GetJuryKeyInfo("PCYL82nJX3rKjxMUTHxWA4wswzWyn1EvpYx", "eth", inputJSON, GetETHAdaptor())
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("prikey: %x\n", key)
	}
}
func TestGetJuryPubkey(t *testing.T) {
	outChainCall := &pb.OutChainCall{OutChainName: "eth", Method: "GetJuryPubkey", Params: []byte("")}
	result, err := ProcessOutChainCall("PCYL82nJX3rKjxMUTHxWA4wswzWyn1EvpYx", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
		return
	} else {
		fmt.Println(result)
	}
	output := adaptor.GetPublicKeyOutput{}
	json.Unmarshal([]byte(result), &output)
	fmt.Printf("%x\n", output.PublicKey)
}
func TestGetJuryAddr(t *testing.T) {
	outChainCall := &pb.OutChainCall{OutChainName: "eth", Method: "GetJuryAddr", Params: []byte("")}
	result, err := ProcessOutChainCall("PCYL82nJX3rKjxMUTHxWA4wswzWyn1EvpYx", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
		return
	} else {
		fmt.Println(result)
	}
}
func TestSignMessage(t *testing.T) {
	msg, _ := hex.DecodeString("c47369fa0759702c6e36bc265eae32dccdd13a0e7d7116a8706ae08baa7f4909e26728fa7a5f03650000000000000000000000000000000000000000000000000000000002fac97091450669af63f6d2a87054ef72c9b24e78e5c48ea2653e3b391559c4b12e798b")
	input := adaptor.SignMessageInput{Message: msg}
	reqBytes, err := json.Marshal(input)
	if err != nil {
		return
	}

	outChainCall := &pb.OutChainCall{OutChainName: "eth", Method: "SignMessage", Params: reqBytes}
	result, err := ProcessOutChainCall("PCYL82nJX3rKjxMUTHxWA4wswzWyn1EvpYx", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
		return
	} else {
		fmt.Println(result)
	}

	output := adaptor.SignMessageOutput{}
	json.Unmarshal([]byte(result), &output)
	fmt.Printf("%x\n", output.Signature)
}

func TestVerifySignature(t *testing.T) {
	msg, _ := hex.DecodeString("c47369fa0759702c6e36bc265eae32dccdd13a0e7d7116a8706ae08baa7f4909e26728fa7a5f03650000000000000000000000000000000000000000000000000000000002fac97091450669af63f6d2a87054ef72c9b24e78e5c48ea2653e3b391559c4b12e798b")
	pubkey, _ := hex.DecodeString("0381361a6c37353f068a234a70ab761420b35a9dfc369c8e5a193afd080a5404c6")
	sig, _ := hex.DecodeString("c3e56e6968e507447da66da70394bab74780bb67e97a08f4097d5d952375d05101d144c8d06677750d63965cb2a9a14cb04937ef6b3d47558208d5bf192c926400")
	input := adaptor.VerifySignatureInput{Message: msg, Signature: sig, PublicKey: pubkey}
	reqBytes, err := json.Marshal(input)
	if err != nil {
		return
	}

	outChainCall := &pb.OutChainCall{OutChainName: "eth", Method: "VerifySignature", Params: reqBytes}
	result, err := ProcessOutChainCall("PCYL82nJX3rKjxMUTHxWA4wswzWyn1EvpYx", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
		return
	} else {
		fmt.Println(result)
	}

	type pubkeyAddr struct {
		Addr   string
		Pubkey []byte
	}

	pubkeyJSON := "[{\"Addr\":\"0x085170BcBd6D9Bb0824592377a43373024A2770F\",\"Pubkey\":\"Ak6WRbraNDCOjfqwZeAd91Joiex5WT0ZGDh7Rlsjl2O0\"},{\"Addr\":\"0x4125cc53BD98242DAD705036BF5AF6EdA96ac0E8\",\"Pubkey\":\"A4E2Gmw3NT8GiiNKcKt2FCCzWp38NpyOWhk6/QgKVATG\"},{\"Addr\":\"0xdd65409A78795a96724800160C539c10640519F6\",\"Pubkey\":\"AzjQhRHkNuHTs3os7DMNdnY72R1yx0yKq6SJp+gCL4ra\"},{\"Addr\":\"0xf7B9D545fD51732FD81eD10426D329c48B20d57A\",\"Pubkey\":\"A2Yk9xTrOgdEjl2aCwHS9XfpJmNLY5JgvyrbZ646XCxI\"}]"
	pubkeyAddrObj := make([]pubkeyAddr, 0)
	json.Unmarshal([]byte(pubkeyJSON), &pubkeyAddrObj)
	for i := range pubkeyAddrObj {
		fmt.Printf("%x\n", pubkeyAddrObj[i].Pubkey)
	}
	for i := range pubkeyAddrObj {
		eth := GetETHAdaptor()
		out, _ := eth.GetAddress(&adaptor.GetAddressInput{Key: pubkeyAddrObj[i].Pubkey})
		addr := out.Address
		fmt.Println("addr : 	", addr)
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
	//txID, _ := hex.DecodeString("51121d1124fb844132f994ef5067ec73f9bbe92b41c12720ae073401f746dc99") //eth transfer
	//txID, _ := hex.DecodeString("498b634c39fbd19af340d66c8866623c124eb0e2160a45aa433644adc636bedb") //eth pending transfer
	//txID, _ := hex.DecodeString("61cded704bd23d8ff7cbe0ac4b62b940bd76f3709f784db695c95efa8074b7df") //panz transfer
	txID, _ := hex.DecodeString("6b2c4379b326757dd5b847f3c584170c5fe2649e6e33f962cf7e9826f77f07b6") //btc op_return
	//txID, _ := hex.DecodeString("4ef356ce0fc244ffb43cc0a941ca293c5b80e91254ad70474ba27acb9eb7b8fd")//panz approve
	input := adaptor.GetTransferTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(string(inputBytes))
	}

	outChainCall := &pb.OutChainCall{OutChainName: "btc", Method: "GetTransferTx", Params: inputBytes}
	result, err := ProcessOutChainCall("sample_syscc", outChainCall)
	if err != nil {
		fmt.Println("err: ", err)
		return
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
