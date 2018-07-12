package txscript

import (
	"regexp"
	"fmt"
	"github.com/palletone/go-palletone/common"
)

/**
scriptPubKey could be like
"scriptPubKey":{
            "asm":"OP_DUP OP_HASH160 aba7915d5964406e8a02c3202f1f8a4a63e95c13 OP_EQUALVERIFY OP_CHECKSIG",
            "hex":"76a914aba7915d5964406e8a02c3202f1f8a4a63e95c1388ac",
            "reqSigs":1,
            "type":"pubkeyhash",
            "addresses":[
               "1GedHcxdxq2tab98hqAmREUK9BBYHKznof"
            ]
         }
*/

var (
	T_P2PKH = "pubkeyhash"
	T_P2SH = "scripthash"
)

type ScriptPubKey struct {
	Asm     []byte
	Hex     []byte //could be like 50 bytes
	RegSigs uint16 // signature num
	Type    string
	Address common.Address // utxo holder address
}

/**
从P2SH（多签）或者P2PKH（单签）
P2PKH Example: OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
P2SH Example: OP_HASH160 <scripthash> OP_EQUAL
 */
func ExtractPkScriptAddrs(pkScript []byte) (*ScriptPubKey, error) {
	spk := ScriptPubKey{}

	reg := regexp.MustCompile(`OP_DUP OP_HASH160 ([\w]+) OP_EQUALVERIFY OP_CHECKSIG`)
	shReq := regexp.MustCompile(`OP_HASH160 ([\w]+) OP_EQUAL`)
	// now just decode P2PKH
	results := reg.FindStringSubmatch(string(pkScript))
	if len(results)==2 {
		spk.Type = T_P2PKH
	} else {
		results = shReq.FindStringSubmatch(string(pkScript))
		if len(results)==2 {
			spk.Type = T_P2SH
		} else {
			return &spk, fmt.Errorf("Extract PkScript Addrs error")
		}
	}

	spk.Address.SetBytes([]byte(results[1]))

	return &spk, nil
}
