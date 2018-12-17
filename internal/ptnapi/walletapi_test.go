package ptnapi
import (
        "fmt"
//        "strings"
        "encoding/json"
        "testing"
        "github.com/palletone/go-palletone/tokenengine"
        "github.com/palletone/go-palletone/ptnjson/walletjson"
        "github.com/palletone/go-palletone/common"
        "github.com/palletone/go-palletone/common/crypto"
        "github.com/palletone/go-palletone/common/hexutil"
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
      //	pubKeyBytes := crypto.CompressPubkey(&pubKey)
        //sign := tokenengine.GenerateP2PKHUnlockScript(signature[0:64],pubKeyBytes)
        pass := crypto.VerifySignature(pubKey, hash.Bytes(), signature[0:64])
        if pass {
                t.Log("Pass")
        } else {
                t.Error("No Pass")
        }
}
 func TestSignHash(t *testing.T) {
        text := "{\"payload\":[{\"inputs\":[{\"txid\":\"0x6cbab00351b1dcd4833242247cc8058a45af3a36eedd30196941e65dd507ea4e\",\"message_index\":0,\"out_index\":0,\"hash\":\"0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470\",\"signature\":\"\"}],\"outputs\":[{\"amount\":10000000000,\"asset\":\"PTN+8000000000000\",\"to_address\":\"P1sn5uKz2SBvhcRKtQGEqrpGB7mkf73btd\"},{\"amount\":99999989000000000,\"asset\":\"PTN+8000000000000\",\"to_address\":\"P1H4uUec5di1wCm8pKGLPxhXM6s7xVutKs9\"}]}],\"invoke_request\":{\"ContractAddress\":\"\",\"FunctionName\":\"\",\"Args\":null}}"
            var RawTxjsonGenParams walletjson.TxJson
	    err := json.Unmarshal([]byte(text), &RawTxjsonGenParams)
	    if err != nil {
		    t.Error("No Pass")
	    }
	
        for index,input:=range RawTxjsonGenParams.Payload[0].Inputs{
            
                  hash := common.HexToHash(input.HashForSign)

                  privateKey := "L3vhqkbATXGc4o7VTG1mT7z1gDEFn1QmVwtxd4kWL9mMJevWnzwo"
                  prvKey, _ := crypto.FromWIF(privateKey)
                  signature, _ := crypto.Sign(hash.Bytes(), prvKey)
    
                  t.Log("Signature is: " + hexutil.Encode(signature))
                  pubKey := crypto.FromECDSAPub(&prvKey.PublicKey)
                  pass := crypto.VerifySignature(pubKey, hash.Bytes(), signature[0:64])
                  if pass {
                      t.Log("Pass")
                      pubKey := prvKey.PublicKey
                      pubKeyBytes := crypto.CompressPubkey(&pubKey)
                      sign := tokenengine.GenerateP2PKHUnlockScript(signature[0:64],pubKeyBytes)
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
