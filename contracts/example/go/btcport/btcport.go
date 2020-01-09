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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	dm "github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"

	"github.com/palletone/adaptor"
)

type BTCPort struct {
}

func (p *BTCPort) Init(stub shim.ChaincodeStubInterface) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(symbolsOwner, []byte(invokeAddr.String()))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}

	return shim.Success(nil)
}

func (p *BTCPort) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "initDepositAddr":
		return p.InitDepositAddr(stub)
	case "setBTCTokenAsset":
		if len(args) < 1 {
			return shim.Error("need 1 args (AssetStr)")
		}
		return p.SetBTCTokenAsset(args[0], stub)
	case "getDepositAddr":
		return p.GetDepositAddr(stub)

	case "setOwner":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNAddr)")
		}
		return p.SetOwner(args[0], stub)

	case "payoutBTCTokenByTxID":
		if len(args) < 1 {
			return shim.Error("need 1 args (btcTxHash)")
		}
		return p.PayoutBTCTokenByTxID(args[0], stub)

	case "withdrawBTC":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNTransferTxID)")
		}
		btcAddrInput := ""
		if len(args) > 1 {
			btcAddrInput = args[1]
		}
		return p.WithdrawBTC(args[0], btcAddrInput, stub)

	case "withdrawSubmit":
		if len(args) < 1 {
			return shim.Error("need 1 args (BTCTxID)")
		}
		return p.WithdrawSubmit(args[0], stub)

	case "Set":
		if len(args) < 2 {
			return shim.Error("need 2 args (Key, Value)")
		}
		return p.Set(stub, args[0], args[1])
	case "get":
		if len(args) < 1 {
			return shim.Error("need 1 args (Key)")
		}
		return p.Get(stub, args[0])

	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

type JuryMsgAddr struct {
	Address string
	Answer  []byte
}

func creatMulti(juryMsg []JuryMsgAddr, stub shim.ChaincodeStubInterface) (*adaptor.CreateMultiSigAddressOutput, error) {
	//
	answers := make([]string, 0, len(juryMsg))
	for i := range juryMsg {
		answers = append(answers, hex.EncodeToString(juryMsg[i].Answer))
	}
	a := sort.StringSlice(answers[0:])
	sort.Sort(a)
	//
	var input adaptor.CreateMultiSigAddressInput
	input.SignCount = consultM
	for i := range answers {
		pubkey, _ := hex.DecodeString(answers[i])
		input.Keys = append(input.Keys, pubkey)
	}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	log.Debugf("inputBytes : %s", string(inputBytes))
	result, err := stub.OutChainCall("btc", "CreateMultiSigAddress", inputBytes)
	if err != nil {
		return nil, errors.New("OutChainCall CreateMultiSigAddress failed: " + err.Error())
	}
	log.Debugf("OutChainCall CreateMultiSigAddress ==== ===== %s", result)

	var output adaptor.CreateMultiSigAddressOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

const symbolsDepositAddr = "btc_multsigAddr"
const symbolsDepositRedeem = "btc_redeem"

const symbolsBTCAsset = "btc_asset"

const symbolsOwner = "owner_"

const symbolsDeposit = "deposit_"

const symbolsWithdraw = "withdraw_"
const symbolsUTXO = "utxo_"

const consultM = 3
const consultN = 4

func (p *BTCPort) InitDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	//
	saveResult, _ := stub.GetState(symbolsDepositAddr)
	if len(saveResult) != 0 {
		return shim.Success([]byte("DepositAddr has been init"))
	}

	result, err := stub.OutChainCall("btc", "GetJuryPubkey", []byte(""))
	if err != nil {
		log.Debugf("OutChainCall GetJuryPubkey err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryPubkey failed " + err.Error())
	}
	var juryPubkey adaptor.GetPublicKeyOutput
	err = json.Unmarshal(result, &juryPubkey)
	if err != nil {
		log.Debugf("OutChainCall GetJuryPubkey Unmarshal err: %s", err.Error())
		return shim.Error("OutChainCall GetJuryPubkey Unmarshal failed " + err.Error())
	}
	log.Debugf("juryPubkey.PublicKey: %x", juryPubkey.PublicKey)

	//
	recvResult, err := consult(stub, []byte("juryBTCPubkey"), juryPubkey.PublicKey)
	if err != nil {
		log.Debugf("consult juryBTCPubkey failed: " + err.Error())
		return shim.Error("consult juryBTCPubkey failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Error("Unmarshal result failed: " + err.Error())
	}
	if len(juryMsg) != consultN {
		return shim.Error("RecvJury result's len not enough")
	}
	log.Debugf("len(juryMsg) : %d", len(juryMsg))

	//
	createResult, err := creatMulti(juryMsg, stub)
	if err != nil {
		return shim.Success([]byte("creatMulti failed" + err.Error()))
	}

	// Write the state to the ledger
	err = stub.PutState(symbolsDepositAddr, []byte(createResult.Address))
	if err != nil {
		return shim.Error("write " + symbolsDepositAddr + " failed: " + err.Error())
	}
	err = stub.PutState(symbolsDepositRedeem, []byte(hex.EncodeToString(createResult.Extra)))
	if err != nil {
		return shim.Error("write " + symbolsDepositRedeem + " failed: " + err.Error())
	}
	return shim.Success([]byte(createResult.Address))
}

func (p *BTCPort) SetBTCTokenAsset(assetStr string, stub shim.ChaincodeStubInterface) pb.Response {
	err := stub.PutState(symbolsBTCAsset, []byte(assetStr))
	if err != nil {
		return shim.Error("write symbolsBTCAsset failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getBTCTokenAsset(stub shim.ChaincodeStubInterface) *dm.Asset {
	result, _ := stub.GetState(symbolsBTCAsset)
	if len(result) == 0 {
		return nil
	}
	asset, _ := dm.StringToAsset(string(result))
	log.Debugf("resultHex %s, asset: %s", common.Bytes2Hex(result), asset.String())

	return asset
}

func getDepositAddr(stub shim.ChaincodeStubInterface) string {
	result, _ := stub.GetState(symbolsDepositAddr)
	if len(result) == 0 {
		return ""
	}
	return string(result)
}

func (p *BTCPort) GetDepositAddr(stub shim.ChaincodeStubInterface) pb.Response {
	result := getDepositAddr(stub)
	if len(result) == 0 {
		return shim.Error("DepsoitAddr is empty")
	}
	return shim.Success([]byte(result))
}

func (p *BTCPort) SetOwner(ptnAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only owner can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can set")
	}
	err = stub.PutState(symbolsOwner, []byte(ptnAddr))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getOwner(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsOwner)
	if len(result) == 0 {
		return "", errors.New("Need set Owner")
	}

	return string(result), nil
}

func getBTCTx(txID []byte, stub shim.ChaincodeStubInterface) (*adaptor.GetTransferTxOutput, error) {
	//
	input := adaptor.GetTransferTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	result, err := stub.OutChainCall("btc", "GetTransferTx", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error: " + err.Error())
	}
	log.Debugf("result : %s", string(result))

	//
	var output adaptor.GetTransferTxOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (p *BTCPort) PayoutBTCTokenByTxID(btcTxHash string, stub shim.ChaincodeStubInterface) pb.Response {
	if "0x" == btcTxHash[0:2] || "0X" == btcTxHash[0:2] {
		btcTxHash = btcTxHash[2:]
	}
	//
	result, _ := stub.GetState(symbolsDeposit + btcTxHash)
	if len(result) != 0 {
		log.Debugf("The tx has been payout")
		return shim.Error("The tx has been payout")
	}

	txIDByte, err := hex.DecodeString(btcTxHash)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}

	depositAddr := getDepositAddr(stub)
	if "" == depositAddr {
		return shim.Error("need call InitDepositAddr")
	}
	//
	txResult, err := getBTCTx(txIDByte, stub)
	if err != nil {
		log.Debugf("getBTCTx failed : " + err.Error())
		return shim.Error("getBTCTx failed : " + err.Error())
	}
	if txResult.Tx.TargetAddress != depositAddr {
		log.Debugf("The tx is't transfer to btc port contract")
		return shim.Error("The tx is't transfer to btc port contract")
	}
	if !txResult.Tx.IsStable {
		log.Debugf("Need more confirms")
		return shim.Error("Need more confirms")
	}

	//bigIntAmount := txResult.Tx.Amount.Amount
	//bigIntAmount = bigIntAmount.Div(bigIntAmount, big.NewInt(1e10)) //btcToken in PTN is decimal is 8
	btcAmount := txResult.Tx.Amount.Amount.Uint64()
	if btcAmount == 0 {
		return shim.Error("need deposit bigger than 0")
	}

	//
	ptnAddr := string(txResult.Tx.AttachData)
	if "" == ptnAddr {
		return shim.Error("need set ptn address in op_return")
	}
	//
	err = stub.PutState(symbolsDeposit+btcTxHash, []byte(ptnAddr+"-"+txResult.Tx.Amount.Amount.String()))
	if err != nil {
		log.Debugf("PutState txhash failed err: %s", err.Error())
		return shim.Error("PutState txhash failed")
	}

	//
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return shim.Error("need call setBTCTokenAsset()")
	}
	invokeTokens := new(dm.AmountAsset)
	invokeTokens.Amount = btcAmount
	invokeTokens.Asset = btcTokenAsset
	err = stub.PayOutToken(ptnAddr, invokeTokens, 0)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.PayOutToken\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success([]byte("get success"))
}

func getUTXOAll(stub shim.ChaincodeStubInterface) []byte {
	KVs, _ := stub.GetStateByPrefix(symbolsUTXO)
	utxoAll := make([]byte, 0, len(KVs)*2*33)
	for _, oneKV := range KVs {
		utxoAll = append(utxoAll, oneKV.Value...)
	}
	return utxoAll
}
func genRawTx(prepare *WithdrawPrepare, depositAddr string, stub shim.ChaincodeStubInterface) (*adaptor.CreateMultiSigPayoutTxOutput, error) {
	utxoAllExcept := getUTXOAll(stub)
	//
	input := adaptor.CreateTransferTokenTxInput{FromAddress: depositAddr, ToAddress: prepare.BtcAddr}
	input.Amount = adaptor.NewAmountAssetUint64(prepare.BtcAmount-prepare.BtcFee, "BTC")
	input.Fee = adaptor.NewAmountAssetUint64(prepare.BtcFee, "BTC")
	input.Extra = utxoAllExcept

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	resultByte, err := stub.OutChainCall("btc", "CreateMultiSigPayoutTx", inputBytes)
	if err != nil {
		return nil, err
	}

	//
	var result adaptor.CreateMultiSigPayoutTxOutput
	err = json.Unmarshal(resultByte, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func signTx(tx []byte, redeemHex string, stub shim.ChaincodeStubInterface) (*adaptor.SignTransactionOutput, error) {
	//
	var input adaptor.SignTransactionInput
	input.Transaction = tx
	input.Extra = []byte(redeemHex)
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	resultByte, err := stub.OutChainCall("btc", "SignTransaction", inputBytes)
	if err != nil {
		return nil, err
	}

	//
	var output adaptor.SignTransactionOutput
	err = json.Unmarshal(resultByte, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func mergeTx(tx []byte, redeemHex string, juryMsg []JuryMsgAddr, stub shim.ChaincodeStubInterface) (*adaptor.BindTxAndSignatureOutput, error) {
	//
	var input adaptor.BindTxAndSignatureInput
	input.Transaction = tx
	input.Extra = []byte(redeemHex)

	//
	answers := make([]string, 0, len(juryMsg))
	for i := range juryMsg {
		answers = append(answers, string(juryMsg[i].Answer))
	}
	a := sort.StringSlice(answers[0:])
	sort.Sort(a)
	answersByte := make([][]byte, 0, len(answers))
	for i := range answers {
		signedHex, _ := hex.DecodeString(answers[i])
		answersByte = append(answersByte, signedHex)
	}

	//
	var output adaptor.BindTxAndSignatureOutput
	array := [][3]int{{1, 2, 3}, {1, 2, 4}, {1, 3, 4}, {2, 3, 4}}
	num := 4
	if len(answers) == 3 {
		num = 1
	}
	for i := 0; i < num; i++ {
		input.SignedTxs = [][]byte{}
		input.SignedTxs = append(input.SignedTxs, answersByte[array[i][0]-1])
		input.SignedTxs = append(input.SignedTxs, answersByte[array[i][1]-1])
		input.SignedTxs = append(input.SignedTxs, answersByte[array[i][2]-1])
		//
		reqBytes, err := json.Marshal(input)
		if err != nil {
			continue
		}
		resultByte, err := stub.OutChainCall("btc", "BindTxAndSignature", reqBytes)
		if err != nil {
			continue
		}
		err = json.Unmarshal(resultByte, &output)
		if err != nil {
			continue
		}
		if 0 != len(output.SignedTx) {
			break
		}
	} //for

	if 0 != len(output.SignedTx) {
		return &output, nil
	}
	return nil, fmt.Errorf("not complete")
}

func sendTx(tx []byte, stub shim.ChaincodeStubInterface) (*adaptor.SendTransactionOutput, error) {
	//
	input := adaptor.SendTransactionInput{Transaction: tx}

	//
	reqBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	resultByte, err := stub.OutChainCall("btc", "SendTransactionHttp", reqBytes)
	if err != nil {
		return nil, err
	}

	//
	var output adaptor.SendTransactionOutput
	err = json.Unmarshal(resultByte, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

type WithdrawPrepare struct {
	BtcAddr   string
	BtcAmount uint64
	BtcFee    uint64
}

func getWithdrawPrepare(txID, btcAddrInput string, stub shim.ChaincodeStubInterface) (*WithdrawPrepare, error) {
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return nil, fmt.Errorf("need call setBTCTokenAsset()")
	}

	// 1 get this tx
	tx, err := stub.GetStableTransactionByHash(txID)
	if nil != err {
		return nil, fmt.Errorf("GetStableTransactionByHash failed " + err.Error())
	}

	//2 sender address, get it from txPre output
	txMsgs := tx.TxMessages()
	payment := txMsgs[0].Payload.(*dm.PaymentPayload)
	outPoint := payment.Inputs[0].PreviousOutPoint
	txPre, err := stub.GetStableTransactionByHash(outPoint.TxHash.String())
	if nil != err {
		return nil, fmt.Errorf("GetStableTransactionByHash txPre failed " + err.Error())
	}
	txPreMsgs := txPre.TxMessages()
	paymentPre := txPreMsgs[outPoint.MessageIndex].Payload.(*dm.PaymentPayload)
	outputPre := paymentPre.Outputs[outPoint.OutIndex]
	toAddr, _ := tokenengine.Instance.GetAddressFromScript(outputPre.PkScript)

	//3 amount, get it from tx output
	_, contractAddr := stub.GetContractID()
	amount := uint64(0)
	newOutpointMapUTXOs := tx.GetNewUtxos()
	for _, utxo := range newOutpointMapUTXOs {
		recvAddr, _ := tokenengine.Instance.GetAddressFromScript(utxo.PkScript)
		if recvAddr.String() != contractAddr {
			continue
		}
		if !utxo.Asset.IsSameAssetId(btcTokenAsset) {
			continue
		}
		amount += utxo.Amount
	}

	//4 op_return
	btcAddr := ""
	for _, msg := range txMsgs {
		if msg.App == dm.APP_DATA {
			text := msg.Payload.(*dm.DataPayload)
			btcAddr = string(text.MainData)
			break
		}
	}
	if btcAddr == "" {
		if btcAddrInput == "" {
			return nil, fmt.Errorf("Get btcAddr failed")
		}
		btcAddr = btcAddrInput
	}

	//get all result
	var prepare WithdrawPrepare
	prepare.BtcFee = 10000 // 0.0001 btc token
	prepare.BtcAddr = btcAddr
	prepare.BtcAmount = amount

	log.Debugf("%s-%d-%s", toAddr.String(), amount, btcAddr)
	return &prepare, nil
}

type Withdraw struct {
	BtcAddr   string
	BtcAmount uint64
	BtcFee    uint64
	SignedTx  []byte
}

func consult(stub shim.ChaincodeStubInterface, content []byte, myAnswer []byte) ([]byte, error) {
	sendResult, err := stub.SendJury(2, content, myAnswer)
	if err != nil {
		log.Debugf("SendJury err: %s", err.Error())
		return nil, errors.New("SendJury failed")
	}
	log.Debugf("sendResult: %s", common.Bytes2Hex(sendResult))
	recvResult, err := stub.RecvJury(2, content, 2)
	if err != nil {
		recvResult, err = stub.RecvJury(2, content, 2)
		if err != nil {
			log.Debugf("RecvJury err: %s", err.Error())
			return nil, errors.New("RecvJury failed")
		}
	}
	log.Debugf("recvResult: %s", string(recvResult))
	return recvResult, nil
}

func (p *BTCPort) WithdrawBTC(txID, btcAddrInput string, stub shim.ChaincodeStubInterface) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can withdraw")
	}

	resultWithdraw, _ := stub.GetState(symbolsWithdraw + txID)
	if len(resultWithdraw) != 0 {
		return shim.Error("The txID has been withdraw")
	}

	prepare, err := getWithdrawPrepare(txID, btcAddrInput, stub)
	if nil != err {
		log.Debugf("getWithdrawPrepare failed : " + err.Error())
		return shim.Error("getWithdrawPrepare failed : " + err.Error())
	}
	if prepare.BtcAmount <= prepare.BtcFee {
		var withdraw Withdraw
		withdraw.BtcAddr = prepare.BtcAddr
		withdraw.BtcAmount = prepare.BtcAmount
		withdraw.BtcFee = prepare.BtcFee
		withdrawBytes, err := json.Marshal(withdraw)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(symbolsWithdraw+txID, withdrawBytes)
		if err != nil {
			log.Debugf("save withdraw failed: " + err.Error())
			return shim.Error("save withdraw failed: " + err.Error())
		}
		return shim.Success(withdrawBytes)
	}
	//
	btcTokenAsset := getBTCTokenAsset(stub)
	if btcTokenAsset == nil {
		return shim.Error("need call setBTCTokenAsset()")
	}
	depositAddr := getDepositAddr(stub)
	if "" == depositAddr {
		return shim.Error("need call InitDepositAddr")
	}

	// 产生交易
	rawTx, err := genRawTx(prepare, depositAddr, stub)
	if nil != err {
		log.Debugf("genRawTx failed : " + err.Error())
		return shim.Error("genRawTx failed : " + err.Error())
	}

	//
	result, _ := stub.GetState(symbolsDepositRedeem)
	if len(result) == 0 {
		return shim.Error("DepsoitRedeem is empty")
	}
	redeemHex := string(result)

	// 签名交易
	rawTxSign, err := signTx(rawTx.Transaction, redeemHex, stub)
	if err != nil {
		log.Debugf("signTx rawTxSign failed: " + err.Error())
		return shim.Error("signTx rawTxSign failed: " + err.Error())
	}
	tempHash := crypto.Keccak256([]byte(rawTx.Extra))
	tempHashHex := fmt.Sprintf("%x", tempHash)

	//用交易哈希协商交易签名，作适当安全防护
	recvResult, err := consult(stub, []byte(tempHashHex), []byte(hex.EncodeToString(rawTxSign.SignedTx)))
	if err != nil {
		log.Debugf("consult tempHashHex failed: " + err.Error())
		return shim.Error("consult tempHashHex failed: " + err.Error())
	}
	var juryMsg []JuryMsgAddr
	err = json.Unmarshal(recvResult, &juryMsg)
	if err != nil {
		return shim.Error("Unmarshal result failed: " + err.Error())
	}
	if len(juryMsg) < consultM {
		return shim.Error("RecvJury result's len not enough")
	}

	// 合并交易
	txMerged, err := mergeTx(rawTx.Transaction, redeemHex, juryMsg, stub)
	if err != nil {
		return shim.Error("mergeTx failed: " + err.Error())
	}

	var withdraw Withdraw
	withdraw.BtcAddr = prepare.BtcAddr
	withdraw.BtcAmount = prepare.BtcAmount
	withdraw.BtcFee = prepare.BtcFee
	withdraw.SignedTx = txMerged.SignedTx
	withdrawBytes, err := json.Marshal(withdraw)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(symbolsWithdraw+txID, withdrawBytes)
	if err != nil {
		log.Debugf("save withdraw failed: " + err.Error())
		return shim.Error("save withdraw failed: " + err.Error())
	}

	txHash := crypto.Keccak256(txMerged.SignedTx)
	txHashHex := fmt.Sprintf("%x", txHash)

	err = stub.PutState(symbolsUTXO+txHashHex, rawTx.Extra)
	if err != nil {
		log.Debugf("save utxo failed: " + err.Error())
		return shim.Error("save utxo failed: " + err.Error())
	}

	return shim.Success(withdrawBytes)
}

func (p *BTCPort) Get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	result, _ := stub.GetState(key)
	return shim.Success(result)
}
func (p *BTCPort) Set(stub shim.ChaincodeStubInterface, key string, value string) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can withdraw")
	}

	err = stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(fmt.Sprintf("PutState failed: %s", err.Error()))
	}
	return shim.Success([]byte("Success"))
}

func (p *BTCPort) WithdrawSubmit(btcTxID string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	if "0x" == btcTxID[0:2] || "0X" == btcTxID[0:2] {
		btcTxID = btcTxID[2:]
	}

	//get sender receiver amount
	txIDByte, err := hex.DecodeString(btcTxID)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}

	txResult, err := getBTCTx(txIDByte, stub)
	if err != nil {
		log.Debugf("getBTCTx failed : " + err.Error())
		return shim.Error("getBTCTx failed : " + err.Error())
	}
	//check tx status
	if !txResult.Tx.IsStable {
		log.Debugf("The tx is not Stable")
		return shim.Error("The tx is not Stable")
	}

	depositAddr := getDepositAddr(stub)
	if "" == depositAddr {
		return shim.Error("need call InitDepositAddr")
	}
	if txResult.Tx.FromAddress != depositAddr {
		log.Debugf("The tx is't payout from btc port contract")
		return shim.Error("The tx is't payout from btc port contract")
	}

	txHash := crypto.Keccak256(txResult.Tx.TxRawData)
	txHashHex := fmt.Sprintf("%x", txHash)
	//
	err = stub.DelState(symbolsUTXO + txHashHex)
	if err != nil {
		log.Debugf("DelState symbolsUTXO failed err: %s", err.Error())
		return shim.Error("DelState symbolsUTXO failed " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func main() {
	err := shim.Start(new(BTCPort))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
