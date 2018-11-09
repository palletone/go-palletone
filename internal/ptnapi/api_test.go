/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package ptnapi

import (
	"fmt"
	//"bytes"
	"encoding/json"
	"github.com/palletone/go-palletone/ptnjson"
	"strings"
	"testing"
	// "github.com/palletone/go-palletone/tokenengine/btcd/btcjson"
)

func TestRawTransactionGen(t *testing.T) {
	// txid from btc
	/*params := `{
	    "inputs": [
			{
	           "txid": "5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873",
	           "vout": 0,
	           "messageindex": 0
			}
	    ],
	    "outputs": [
			{
	           "address": "P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX",
	           "amount": 0.79
			}
	    ],
	    "locktime": 0
		}`*/
	params := `{
    "inputs": [
		{
           "txid": "b0bec28ef271525381d602b0b0035c27ec9896c3eda4a5ce58f33e94cd4da970",
           "vout": 0,
           "messageindex": 0
		}
    ],
    "outputs": [
		{
           "address": "P1KzS9JG7XCZvdKRNwL47mJWCGprCCt8j8D",
           "amount": 100000
		}
    ],
    "locktime": 0
	}`
	params = params
	testResult := "f89ea0019584e8148b779b2489296f98859681ea48a23e4734ce5b4a2a55a1d0304de7f87bf87980b876f874e7e6e3a0b0bec28ef271525381d602b0b0035c27ec9896c3eda4a5ce58f33e94cd4da97080808080f849f8478880000000000000009976a914d04ef6595ea6dd1cf512a5e9077a66f9b9fb422688ace3900000000000000000000000000000000090000000000000000000000000000000008080"
	var rawTransactionGenParams ptnjson.RawTransactionGenParams
	err := json.Unmarshal([]byte(params), &rawTransactionGenParams)
	if err != nil {
		return
	}
	//transaction inputs
	var inputs []ptnjson.TransactionInput
	for _, inputOne := range rawTransactionGenParams.Inputs {
		input := ptnjson.TransactionInput{inputOne.Txid, inputOne.Vout, inputOne.MessageIndex}
		inputs = append(inputs, input)
	}
	if len(inputs) == 0 {
		return
	}
	//realNet := &chaincfg.MainNetParams
	amounts := map[string]float64{}
	for _, outOne := range rawTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount <= 0 {
			continue
		}
		amounts[outOne.Address] = float64(outOne.Amount * 1e8)
	}
	if len(amounts) == 0 {
		return
	}

	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &rawTransactionGenParams.Locktime)

	result, _ := CreateRawTransaction(arg)
	if !strings.Contains(result, testResult) {
		t.Errorf("unexpected result - got: %v, "+"want: %v", result, testResult)
	}
	fmt.Println(result)
	return
}

// func TestVoteTransactionGen(t *testing.T) {
// 	params := `{
//     "inputs": [
// 		{
//            "txid": "b0bec28ef271525381d602b0b0035c27ec9896c3eda4a5ce58f33e94cd4da970",
//            "vout": 0,
//            "messageindex": 0
// 		}
//     ],
//     "outputs": [
// 		{
//            "address": "P1KzS9JG7XCZvdKRNwL47mJWCGprCCt8j8D",
//            "amount": 100000
// 		}
//     ],
//     "locktime": 0,
// 	"expiredterm": 0,
// 	"mediatoraddress": "P1KzS9JG7XCZvdKRNwL47mJWCGprCCt8j8D"

// 	}`
// 	params = params
// 	testResult := "f8dda01166fd7e9ccbf92da6351e07e6dea0f2a8774694ab1b5468dce73baccb4a06c1f8baf88f80b88cf88ae7e6e3a0b0bec28ef271525381d602b0b0035c27ec9896c3eda4a5ce58f33e94cd4da97080808080f85ff85d8880000000000000009976a914d04ef6595ea6dd1cf512a5e9077a66f9b9fb422688ace390000000000000000000000000000000009000000000000000000000000000000000809500000000000000000000000000000000000000000080e806a6e5a350314b7a53394a473758435a76644b524e774c34376d4a5743477072434374386a384480"
// 	var voteTransactionGenParams ptnjson.VoteTransactionGenParams
// 	err := json.Unmarshal([]byte(params), &voteTransactionGenParams)
// 	if err != nil {
// 		return
// 	}
// 	//fmt.Println("voteTransactionGenParams:",voteTransactionGenParams)
// 	//transaction inputs
// 	var inputs []ptnjson.TransactionInput
// 	for _, inputOne := range voteTransactionGenParams.Inputs {
// 		input := ptnjson.TransactionInput{inputOne.Txid, inputOne.Vout, inputOne.MessageIndex}
// 		inputs = append(inputs, input)
// 	}
// 	if len(inputs) == 0 {
// 		return
// 	}
// 	//realNet := &chaincfg.MainNetParams
// 	amounts := map[string]float64{}
// 	for _, outOne := range voteTransactionGenParams.Outputs {
// 		if len(outOne.Address) == 0 || outOne.Amount <= 0 {
// 			continue
// 		}
// 		amounts[outOne.Address] = float64(outOne.Amount * 1e8)
// 	}
// 	if len(amounts) == 0 {
// 		return
// 	}

// 	MediatorAddress := voteTransactionGenParams.MediatorAddress
// 	ExpiredTerm := voteTransactionGenParams.ExpiredTerm
// 	arg := ptnjson.NewCreateVoteTransactionCmd(inputs, amounts, &voteTransactionGenParams.Locktime, MediatorAddress, ExpiredTerm)

// 	result, _ := CreateVoteTransaction(arg)
// 	if !strings.Contains(result, testResult) {
// 		t.Errorf("unexpected result - got: %v, "+"want: %v", result, testResult)
// 	}
// 	fmt.Println(result)
// 	return
// }

/*
func TestDecodeRawTransaction(t *testing.T) {

	rpcParams := RPCParams{
		Host:      "localhost:18332",
		RPCUser:   "zxl",
		RPCPasswd: "123456",
		CertPath:  "C:/Users/zxl/AppData/Local/Btcwallet/rpc.cert",
	}

	testResult := `{"hex":"","txid":"0bf2bbdabd7561fe035eb383d14e376f04690c62301cc78d89dd189f7e6c3a72","version":1,"locktime":0,"vin":[{"txid":"132154398e312b69b62973f8f6a91797bba9996bc60dc1d7b1f8697df196088d","vout":0,"scriptSig":{"asm":"","hex":""},"sequence":4294967295}],"vout":[{"value":0.98811339,"n":0,"scriptPubKey":{"asm":"OP_DUP OP_HASH160 bddc9a62e9b7c3cfdbe1c817520e24e32c339f32 OP_EQUALVERIFY OP_CHECKSIG","hex":"76a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac","reqSigs":1,"type":"pubkeyhash","addresses":["mxprH5bkXtn9tTTAxdQGPXrvruCUvsBNKt"]}}]}`

	parms := ` {
		    "rawtx": "01000000018d0896f17d69f8b1d7c10dc66b99a9bb9717a9f6f87329b6692b318e395421130000000000ffffffff01cbbde305000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000"
		  	}`
	result := DecodeRawTransaction(parms, &rpcParams)
	fmt.Println(result)
	if !strings.Contains(result, testResult) {
		t.Errorf("unexpected result - got: %v, "+"want: %v", result, testResult)
	}
}*/

func TestSignTransaction(t *testing.T) {
	//from TestRawTransactionGen A --> B C
	//参数格式错误
	params := `{
	"rawtx":"f8b4a01085560d869f0f2f3eb696d068713adf20879a333f5e11672ebda8f0fad8f82bf891f88f01b88cf88ae7e6e3a0b0bec28ef271525381d602b0b0035c27ec9896c3eda4a5ce58f33e94cd4da97080808080f85ff85d8880000000000000009976a914d04ef6595ea6dd1cf512a5e9077a66f9b9fb422688ace390000000000000000000000000000000009000000000000000000000000000000000809500000000000000000000000000000000000000000080",
    "rawtxinput": [
		{
           "txid": "b0bec28ef271525381d602b0b0035c27ec9896c3eda4a5ce58f33e94cd4da970",
           "vout": 0,
           "messageindex": 0,
           "scriptPubKey":"76a9146b9fc0dd015308b8a29c460256150500cbae86cc88ac",
           "redeemScript":""
		}
    ],
    "privkeys":["L3XNk86G6LXpFr9y7JSYDztgtmcC7xdcVoUkiF8dNCPRroDSyBHe"]
	}`
	/*params := `{
	  "transactionhex": "010000000236045404e65bd741109db92227ca0dc9274ef717a6612c96cd77b24a17d1bcd70000000000ffffffff7c1f7d5407b41abf29d41cf6f122ef2d40f76d956900d2c89314970951ef5b940000000000ffffffff014431d309000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000",
	  "redeemhex": "522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553ae",
	      "privkeys": ["cUakDAWEeNeXTo3B93WBs9HRMfaFDegXcbEGooLz8BSxRBfmpYcX"]
	      }`*/
	/*params := `{
	  "transactionhex": "010000000236045404e65bd741109db92227ca0dc9274ef717a6612c96cd77b24a17d1bcd700000000b400473044022024e6a6ca006f25ccd3ebf5dadf21397a6d7266536cd336061cd17cff189d95e402205af143f6726d75ac77bc8c80edcb6c56579053d2aa31601b23bc8da41385dd86014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff7c1f7d5407b41abf29d41cf6f122ef2d40f76d956900d2c89314970951ef5b9400000000b40047304402206a1d7a2ae07840957bee708b6d3e1fbe7858760ac378b1e21209b348c1e2a5c402204255cd4cd4e5b5805d44bbebe7464aa021377dca5fc6bf4a5632eb2d8bc9f9e4014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff014431d309000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000",
	  "redeemhex": "522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553ae",
	      "privkeys": ["cQJB6w8SxVNoprVwp2xyxUFxvExMbpR2qj3banXYYXmhtTc1WxC8"]
	      }`*/

	var signTransactionParams SignTransactionParams
	err := json.Unmarshal([]byte(params), &signTransactionParams)
	if err != nil {
		return
	}

	//transaction inputs
	var rawinputs []ptnjson.RawTxInput
	for _, inputOne := range signTransactionParams.Inputs {
		input := ptnjson.RawTxInput{inputOne.Txid, inputOne.Vout, inputOne.MessageIndex, inputOne.ScriptPubKey, inputOne.RedeemScript}
		rawinputs = append(rawinputs, input)
	}
	if len(rawinputs) == 0 {
		return
	}
	var keys []string
	for _, key := range signTransactionParams.PrivKeys {
		key = strings.TrimSpace(key) //Trim whitespace
		if len(key) == 0 {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return
	}

	newsign := ptnjson.NewSignRawTransactionCmd(signTransactionParams.RawTx, &rawinputs, &keys, ptnjson.String("ALL"))
	//fmt.Println("newsign:", newsign)
	//pk, _ := crypto.FromWIF("L3XNk86G6LXpFr9y7JSYDztgtmcC7xdcVoUkiF8dNCPRroDSyBHe")
	//addr := crypto.PubkeyToAddress(&pk.PublicKey)
	//script := tokenengine.GenerateLockScript(addr)
	//fmt.Printf("Script: %x\n", script)

	/*newsign := ptnjson.SignRawTransactionCmd{
		RawTx: "f8b7a00000000000000000000000000000000000000000000000000000000000000000f894f89201b88ff88de7e6e3a07348b3dba6d43e10d7140e737a05bed70a1d17220a96bd6d3794c8a80a87111180808080f862f860019976a914ff9452a168a8d8f1a5a1c6912acc4a3ff465e3e988acf843a00000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000008080",
		Inputs: &[]ptnjson.RawTxInput{
			{
				Txid:         "1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873",
				Vout:         0,
				MessageIndex: 0,
				ScriptPubKey: "a914c5d9f1e6ded8332e8114fcdbbaf47c612129719187",
				RedeemScript: "522102f9286c44fe7ebff9788425d5025ad764cdf5aec3daef862d143c5f09134d75b02102a2ba6f2a6e1467334d032ec54ac862c655d7e8bd6bbbce36c771fcdc0ddfb01f21020945d0c9ed05cf5ca9fe38dde799d7b73f4a5bfb71fc1a3c1dca79e2d86462a753ae",
			},
		},
		PrivKeys: &[]string{"KwN8TdhAMeU8b9UrEYTNTVEvDsy9CSyepwRVNEy2Fc9nbGqDZw4J"},
		Flags:    ptnjson.String("ALL"),
	}*/
	/*newsign := ptnjson.SignRawTransactionCmd{
	RawTx: "f9016ba00000000000000000000000000000000000000000000000000000000000000000f90147f9014401b90140f9013df8d6f8d4e3a07348b3dba6d43e10d7140e737a05bed70a1d17220a96bd6d3794c8a80a8711118080b8ad0040bd1e91f7c54de07fff977217c6edba88bf7f8eb046ab0bbf2e444a423a2e5c48069483b82d00f0baa5484aeaf45f88e5b9b80b7aaca7d19697eb5503c2a1d20e4c69522102f9286c44fe7ebff9788425d5025ad764cdf5aec3daef862d143c5f09134d75b02102a2ba6f2a6e1467334d032ec54ac862c655d7e8bd6bbbce36c771fcdc0ddfb01f21020945d0c9ed05cf5ca9fe38dde799d7b73f4a5bfb71fc1a3c1dca79e2d86462a753ae80f862f860019976a914ff9452a168a8d8f1a5a1c6912acc4a3ff465e3e988acf843a00000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000008080",
	Inputs: &[]ptnjson.RawTxInput{
		{
			Txid:         "1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873",
			Vout:         0,
			MessageIndex: 0,
			ScriptPubKey: "a914c5d9f1e6ded8332e8114fcdbbaf47c612129719187",
			RedeemScript: "522102f9286c44fe7ebff9788425d5025ad764cdf5aec3daef862d143c5f09134d75b02102a2ba6f2a6e1467334d032ec54ac862c655d7e8bd6bbbce36c771fcdc0ddfb01f21020945d0c9ed05cf5ca9fe38dde799d7b73f4a5bfb71fc1a3c1dca79e2d86462a753ae",
		},
	},
		PrivKeys: &[]string{"Ky7gQF2rxXLjGSymCtCMa67N2YMt98fRgxyy5WfH92FpbWDxWVRM"},
		Flags:    ptnjson.String("ALL"),
	}*/
	//check empty string
	if "" == newsign.RawTx {
		return
	}

	resultTransToMultsigAddr, _ := SignRawTransaction(newsign)
	//	if !strings.Contains(resultTransToMultsigAddr, theComplete) {
	//		t.Errorf("complete - got: false, want: true")
	//	}

	fmt.Println("Signed Result is:", resultTransToMultsigAddr)
	return
}
