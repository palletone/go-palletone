package ptnapi
import (
//      "fmt"
        "testing"
        "github.com/palletone/go-palletone/common"
        "github.com/palletone/go-palletone/common/crypto"
        "github.com/palletone/go-palletone/common/hexutil"
)

func TestSignHash(t *testing.T) {
        text := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
        hash := common.HexToHash(text)

        privateKey := "L3vhqkbATXGc4o7VTG1mT7z1gDEFn1QmVwtxd4kWL9mMJevWnzwo"
        // privateKeyBytes := hexutil.MustDecode(privateKey)
        prvKey, _ := crypto.FromWIF(privateKey)

        //signB, _ := hexutil.Decode(sign)
        signature, _ := crypto.Sign(hash.Bytes(), prvKey)
        t.Log("Signature is: " + hexutil.Encode(signature))
        pubKey := crypto.FromECDSAPub(&prvKey.PublicKey)
        pass := crypto.VerifySignature(pubKey, hash.Bytes(), signature[0:64])
        if pass {
                t.Log("Pass")
        } else {
                t.Error("No Pass")
        }
}