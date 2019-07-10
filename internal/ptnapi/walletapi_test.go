package ptnapi

import (
	"fmt"
	//        "strings"
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/ptnjson/walletjson"
	"github.com/palletone/go-palletone/tokenengine"
	"testing"
)

func TestSimpleSignHash(t *testing.T) {
	text := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	hash := common.HexToHash(text)

	privateKey := "L3vhqkbATXGc4o7VTG1mT7z1gDEFn1QmVwtxd4kWL9mMJevWnzwo"
	// privateKeyBytes := hexutil.MustDecode(privateKey)
	prvKey, _ := crypto.FromWIF(privateKey)

	//signB, _ := hexutil.Decode(sign)
	signature, _ := crypto.Sign(hash.Bytes(), prvKey)
	t.Log("Signature is: " + hexutil.Encode(signature))
	pubKey := crypto.FromECDSAPub(&prvKey.PublicKey)
	//pubKey1 := prvKey.PublicKey
	//pubKeyBytes := crypto.CompressPubkey(&pubKey)
	//sign := tokenengine.GenerateP2PKHUnlockScript(signature[0:64],pubKeyBytes)
	pass := crypto.VerifySignature(pubKey, hash.Bytes(), signature)
	if pass {
		t.Log("Pass")
	} else {
		t.Error("No Pass")
	}
}
func TestSignHash(t *testing.T) {
	text := "{\"payload\":[{\"inputs\":[{\"txid\":\"0x6cbab00351b1dcd4833242247cc8058a45af3a36eedd30196941e65dd507ea4e\",\"message_index\":0,\"out_index\":0,\"hash\":\"0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470\",\"signature\":\"\"}],\"outputs\":[{\"amount\":10000000000,\"asset\":\"PTN+8000000000000\",\"to_address\":\"P1sn5uKz2SBvhcRKtQGEqrpGB7mkf73btd\"},{\"amount\":99999989000000000,\"asset\":\"PTN+8000000000000\",\"to_address\":\"P1H4uUec5di1wCm8pKGLPxhXM6s7xVutKs9\"}]}],\"invoke_request\":{\"ContractAddress\":\"\",\"Args\":null}}"
	var RawTxjsonGenParams walletjson.TxJson
	err := json.Unmarshal([]byte(text), &RawTxjsonGenParams)
	if err != nil {
		t.Error("No Pass")
	}

	for index, input := range RawTxjsonGenParams.Payload[0].Inputs {

		hash := common.HexToHash(input.HashForSign)

		privateKey := "L3vhqkbATXGc4o7VTG1mT7z1gDEFn1QmVwtxd4kWL9mMJevWnzwo"
		prvKey, _ := crypto.FromWIF(privateKey)
		signature, _ := crypto.Sign(hash.Bytes(), prvKey)

		t.Log("Signature is: " + hexutil.Encode(signature))
		pubKey := crypto.FromECDSAPub(&prvKey.PublicKey)
		pass := crypto.VerifySignature(pubKey, hash.Bytes(), signature)
		if pass {
			t.Log("Pass")
			pubKey := prvKey.PublicKey
			pubKeyBytes := crypto.CompressPubkey(&pubKey)
			sign := tokenengine.GenerateP2PKHUnlockScript(signature, pubKeyBytes)
			hs := hexutil.Encode(sign)
			RawTxjsonGenParams.Payload[0].Inputs[index].Signature = hs
			//dc,err:=hexutil.Decode(hs)
			//err=err
		} else {
			t.Error("No Pass")
		}
	}
	jsonTx, err := json.Marshal(RawTxjsonGenParams)
	if err != nil {
		t.Log("生成json字符串错误")
	}
	fmt.Println(string(jsonTx))
}

func TestJsSign(t *testing.T) {
	privateKey := "L3KxwagZok1yvVaNEg3dkhx2Wft8zoszxgDz8JyvvQDzH2y53ryL"
	h := hexutil.Encode([]byte("L3KxwagZok1yvVaNEg3dkhx2Wft8zoszxgDz8JyvvQDzH2y53ryL"))
	t.Log(h)
	signature, _ := hexutil.Decode("0x11a566e95a3e38d9e0b7115c11513a9b6e4b0ea5989cd1edd549fae11e8422053683ea3e5bce5d8415476ea94d2e1dbaee7f9e4f2166ff5e17601409a7d61a491c")
	prvKey, _ := crypto.FromWIF(privateKey)
	pubKey := prvKey.PublicKey
	pubKeyBytes := crypto.CompressPubkey(&pubKey)
	//pubKeyBytes,_ := hexutil2.Decode("072aa614647a979f360cba4e5f1f825a71779ba954e15dc50a765c40ec11807f1e0ada7de14d8cccf9f3364ca7698ce9d1ccf5f57c59b49cc37bf1925a436ed0")
	t.Log(pubKeyBytes)
	sign := tokenengine.GenerateP2PKHUnlockScript(signature, pubKeyBytes)
	hs := hexutil.Encode(sign)
	t.Log(hs)
}
