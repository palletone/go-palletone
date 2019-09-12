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
package jury

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/tokenengine"
)

const (
	ContractFeeTypeTimeOut = 1 //deploy during time, other is timeout
	ContractFeeTypeTxSize  = 2
)

func localIsMinSignature(tx *modules.Transaction) bool {
	if tx == nil || len(tx.TxMessages) < 3 {
		return false
	}
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(*modules.SignaturePayload)
			sigs := sigPayload.Signatures
			localSig := sigs[0].Signature

			for i := 1; i < len(sigs); i++ {
				if sigs[i].Signature == nil {
					return false
				}
				if bytes.Compare(localSig, sigs[i].Signature) >= 1 {
					return false
				}
			}
			//log.Debug("localIsMinSignature", "local sig", localSig)
			return true
		}
	}
	return false
}
func generateJuryRedeemScript(jury []modules.ElectionInf) []byte {
	count := len(jury)
	needed := byte(math.Ceil((float64(count)*2 + 1) / 3))
	pubKeys := [][]byte{}
	for _, jurior := range jury {
		pubKeys = append(pubKeys, jurior.PublicKey)
	}
	return tokenengine.Instance.GenerateRedeemScript(needed, pubKeys)
}

//对于Contract Payout的情况，将SignatureSet转移到Payment的解锁脚本中
func processContractPayout(tx *modules.Transaction, ele *modules.ElectionNode) {
	if tx == nil || ele == nil {
		log.Error("processContractPayout param is nil")
		return
	}
	reqId := tx.RequestHash()
	if has, payout := tx.HasContractPayoutMsg(); has {
		pubkeys, signs := getSignature(tx)
		redeem := generateJuryRedeemScript(ele.EleList)

		signsOrder := SortSigs(pubkeys, signs, redeem)
		unlock := tokenengine.Instance.MergeContractUnlockScript(signsOrder, redeem)
		log.DebugDynamic(func() string {
			unlockStr, _ := tokenengine.Instance.DisasmString(unlock)
			return fmt.Sprintf("[%s]processContractPayout, Move sign payload to contract payout unlock script:%s",
				shortId(reqId.String()), unlockStr)
		})
		for _, input := range payout.Inputs {
			input.SignatureScript = unlock
		}
	}
	//remove signature payload
	msgs := []*modules.Message{}
	for _, msg := range tx.TxMessages {
		if msg.App != modules.APP_SIGNATURE {
			msgs = append(msgs, msg)
		}
	}
	log.Debugf("[%s]processContractPayout, Remove SignaturePayload from req[%s]", shortId(reqId.String()), reqId.String())
	tx.TxMessages = msgs
}

func DeleOneMax(signs [][]byte) [][]byte {
	n := len(signs)
	maxSig := signs[0]
	max := 0
	for i := 1; i < n; i++ {
		if bytes.Compare(maxSig, signs[i]) < 0 {
			max = i
		}
	}
	var signsNew [][]byte
	for i := 0; i < max; i++ {
		signsNew = append(signsNew, signs[i])
	}
	for i := max + 1; i < n; i++ {
		signsNew = append(signsNew, signs[i])
	}
	return signsNew
}

func SortSigs(pubkeys [][]byte, signs [][]byte, redeem []byte) [][]byte {
	//get all pubkey of redeem
	redeemStr, _ := tokenengine.Instance.DisasmString(redeem)
	pubkeyStrs := strings.Split(redeemStr, " ")
	if len(pubkeyStrs) < 3 {
		log.Debugf("invalid redeemStr %s", redeemStr)
	}
	pubkeyBytes := [][]byte{}
	for i := 1; i < len(pubkeyStrs)-2; i++ {
		//log.Debugf("%d %s", i, pubkeyStrs[i])//the order of redeem's Pubkey
		pubkeyBytes = append(pubkeyBytes, common.Hex2Bytes(pubkeyStrs[i]))
	}

	//select sign by public key
	signsNew := make([][]byte, len(pubkeyBytes))
	for i := range pubkeys {
		for j := range pubkeyBytes {
			if bytes.Equal(pubkeys[i], pubkeyBytes[j]) {
				signsNew[j] = signs[i]
				break
			}
		}
	}
	//get order signs by public key
	signsOrder := [][]byte{}
	for i := range signsNew {
		if len(signsNew[i]) > 0 {
			signsOrder = append(signsOrder, signsNew[i])
		}
	}

	//delete max for leave needed min
	needed, _ := strconv.Atoi(pubkeyStrs[0])
	delNum := len(signsOrder) - needed
	for delNum > 0 {
		signsOrder = DeleOneMax(signsOrder)
		delNum--
	}

	return signsOrder
}

func getSignature(tx *modules.Transaction) ([][]byte, [][]byte) {
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sig := msg.Payload.(*modules.SignaturePayload)
			pubKeys := [][]byte{}
			signs := [][]byte{}
			for _, s := range sig.Signatures {
				pubKeys = append(pubKeys, s.PubKey)
				signs = append(signs, s.Signature)
			}
			return pubKeys, signs
		}
	}
	return nil, nil
}

func genContractErrorMsg(dag iDag, tx *modules.Transaction, addr []byte,
	errIn error, errMsgEnable bool) ([]*modules.Message, error) {
	reqType, _ := getContractTxType(tx)
	errString := fmt.Sprintf("[%s]genContractErrorMsg, reqType:%d,err:%s",
		shortId(tx.RequestHash().String()), reqType, errIn.Error())
	log.Error(errString)
	if !errMsgEnable {
		return nil, errors.New(errString)
	}
	msgs := make([]*modules.Message, 0)
	errMsg := createContractErrorPayloadMsg(tx, errIn)
	msgs = append(msgs, errMsg)

	//合约发生错误，检查有没有支付到合约的Token，有则原路返回
	paybacks := contractPayBack(tx, addr, dag.GetUtxoEntry)
	msgs = append(msgs, paybacks...)

	return msgs, nil
}

func createContractErrorPayloadMsg(tx *modules.Transaction, errIn error) *modules.Message {
	contractErr := modules.ContractError{
		Code:    500, //todo
		Message: errIn.Error(),
	}
	reqType, _ := getContractTxType(tx)
	contractReq, _ := getContractTxContractInfo(tx, reqType)
	switch reqType {
	case modules.APP_CONTRACT_TPL_REQUEST:
		//req := contractReq.(*modules.ContractInstallRequestPayload)
		payload := modules.NewContractTplPayload(nil, 0, nil, contractErr)
		return modules.NewMessage(modules.APP_CONTRACT_TPL, payload)
	case modules.APP_CONTRACT_DEPLOY_REQUEST:
		req := contractReq.(*modules.ContractDeployRequestPayload)
		payload := modules.NewContractDeployPayload(req.TemplateId, nil, "", req.Args, nil, nil, nil, contractErr)
		return modules.NewMessage(modules.APP_CONTRACT_DEPLOY, payload)
	case modules.APP_CONTRACT_INVOKE_REQUEST:
		req := contractReq.(*modules.ContractInvokeRequestPayload)
		payload := modules.NewContractInvokePayload(req.ContractId, nil, nil, nil, contractErr)
		return modules.NewMessage(modules.APP_CONTRACT_INVOKE, payload)
	case modules.APP_CONTRACT_STOP_REQUEST:
		req := contractReq.(*modules.ContractStopRequestPayload)
		payload := modules.NewContractStopPayload(req.ContractId, nil, nil, contractErr)
		return modules.NewMessage(modules.APP_CONTRACT_STOP, payload)
	}

	return nil
}

//执行合约命令:install、deploy、invoke、stop，同时只支持一种类型
func runContractCmd(rwM rwset.TxManager, dag iDag, contract *contracts.Contract, tx *modules.Transaction,
	ele *modules.ElectionNode, errMsgEnable bool) ([]*modules.Message, error) {
	if tx == nil || len(tx.TxMessages) <= 0 {
		return nil, errors.New("runContractCmd transaction or msg is nil")
	}
	for _, msg := range tx.TxMessages {
		switch msg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractInstallRequestPayload)
				req := ContractInstallReq{
					chainID:       "palletone",
					ccName:        reqPay.TplName,
					ccPath:        reqPay.Path,
					ccVersion:     reqPay.Version,
					addrHash:      reqPay.AddrHash,
					ccDescription: reqPay.TplDescription,
					ccAbi:         reqPay.Abi,
					ccLanguage:    reqPay.Language,
				}
				installResult, err := ContractProcess(rwM, contract, req)
				if err != nil {
					return genContractErrorMsg(dag, tx, nil, err, errMsgEnable)
				}
				payload := installResult.(*modules.ContractTplPayload)
				//payload.AddrHash = req.addrHash
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_TPL, payload))
				return msgs, nil
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractDeployRequestPayload)
				req := ContractDeployReq{
					chainID:    "palletone",
					templateId: reqPay.TemplateId,
					txid:       tx.RequestHash().String(),
					args:       reqPay.Args,
					timeout:    time.Duration(reqPay.Timeout) * time.Second,
				}
				fullArgs, err := handleMsg0(tx, dag, req.args)
				if err != nil {
					return nil, err
				}
				req.args = fullArgs
				deployResult, err := ContractProcess(rwM, contract, req)
				if err != nil {
					return genContractErrorMsg(dag, tx, nil, err, errMsgEnable)
				}
				payload := deployResult.(*modules.ContractDeployPayload)
				if ele != nil {
					payload.EleNode = *ele
				}
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_DEPLOY, payload))
				return msgs, nil
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractInvokeRequestPayload)
				req := ContractInvokeReq{
					chainID:  "palletone",
					deployId: reqPay.ContractId,
					args:     reqPay.Args,
					txid:     tx.RequestHash().String(),
					timeout:  time.Duration(reqPay.Timeout) * time.Second,
				}

				fullArgs, err := handleMsg0(tx, dag, req.args)
				if err != nil {
					return nil, err
				}
				// add cert id to args
				newFullArgs, err := handleArg1(tx, fullArgs)
				if err != nil {
					return nil, err
				}
				req.args = newFullArgs
				invokeResult, err := ContractProcess(rwM, contract, req)
				if err != nil {
					return genContractErrorMsg(dag, tx, reqPay.ContractId, err, errMsgEnable)
				}
				result := invokeResult.(*modules.ContractInvokeResult)
				payload := modules.NewContractInvokePayload(result.ContractId, result.ReadSet, result.WriteSet,
					result.Payload, modules.ContractError{})
				if payload != nil {
					msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_INVOKE, payload))
				}
				toContractPayments, err := resultToContractPayments(dag, result)
				if err != nil {
					return genContractErrorMsg(dag, tx, reqPay.ContractId, err, errMsgEnable)
				}
				if len(toContractPayments) > 0 {
					for _, contractPayment := range toContractPayments {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, contractPayment))
					}
				}
				cs, err := resultToCoinbase(result)
				if err != nil {
					return genContractErrorMsg(dag, tx, reqPay.ContractId, err, errMsgEnable)
				}
				if len(cs) > 0 {
					for _, coinbase := range cs {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, coinbase))
					}
				}
				return msgs, nil
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractStopRequestPayload)
				req := ContractStopReq{
					chainID:     "palletone",
					deployId:    reqPay.ContractId,
					txid:        tx.RequestHash().String(),
					deleteImage: reqPay.DeleteImage,
				}
				stopResult, err := ContractProcess(rwM, contract, req)
				if err != nil {
					return genContractErrorMsg(dag, tx, reqPay.ContractId, err, errMsgEnable)
				}
				payload := stopResult.(*modules.ContractStopPayload)
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_STOP, payload))
				return msgs, nil
			}
		}
	}

	return nil, errors.New(fmt.Sprintf("runContractCmd err, txid=%s", tx.RequestHash().String()))
}

func contractPayBack(tx *modules.Transaction, addr []byte, queryUtxoFunc modules.QueryUtxoFunc) []*modules.Message {
	messages := []*modules.Message{}
	for msgIdx, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			payment := msg.Payload.(*modules.PaymentPayload)
			for outIdx, out := range payment.Outputs {
				toAddr, _ := tokenengine.Instance.GetAddressFromScript(out.PkScript)
				if addr != nil && bytes.Equal(toAddr.Bytes(), addr) {
					input := modules.NewTxIn(modules.NewOutPoint(common.NewSelfHash(), uint32(msgIdx), uint32(outIdx)), nil)
					inputUtxo, _ := queryUtxoFunc(payment.Inputs[0].PreviousOutPoint)
					fromAddr, _ := tokenengine.Instance.GetAddressFromScript(inputUtxo.PkScript)
					output := modules.NewTxOut(out.Value, tokenengine.Instance.GenerateLockScript(fromAddr), out.Asset)
					payback := modules.NewPaymentPayload([]*modules.Input{input}, []*modules.Output{output})
					messages = append(messages, modules.NewMessage(modules.APP_PAYMENT, payback))
				}
			}
		}
	}
	return messages
}
func handleMsg0(tx *modules.Transaction, dag iDag, reqArgs [][]byte) ([][]byte, error) {
	var txArgs [][]byte
	invokeInfo := modules.InvokeInfo{}
	lenTxMsgs := len(tx.TxMessages)
	if lenTxMsgs > 0 {
		msg0 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
		invokeAddr, err := dag.GetAddrByOutPoint(msg0.Inputs[0].PreviousOutPoint)
		if err != nil {
			return nil, err
		}
		var invokeTokensAll []*modules.InvokeTokens
		for i := 0; i < lenTxMsgs; i++ {
			msg, ok := tx.TxMessages[i].Payload.(*modules.PaymentPayload)
			if !ok {
				continue
			}
			for _, output := range msg.Outputs {
				addr, err := tokenengine.Instance.GetAddressFromScript(output.PkScript)
				if err != nil {
					return nil, err
				}
				if !addr.Equal(invokeAddr) { //note : only return not invokeAddr
					invokeTokens := &modules.InvokeTokens{}
					invokeTokens.Asset = output.Asset
					invokeTokens.Amount += output.Value
					invokeTokens.Address = addr.String()
					invokeTokensAll = append(invokeTokensAll, invokeTokens)
				}
			}
		}
		invokeInfo.InvokeTokens = invokeTokensAll
		invokeFees, err := dag.GetTxFee(tx)
		if err != nil {
			return nil, err
		}

		invokeInfo.InvokeAddress = invokeAddr
		invokeInfo.InvokeFees = invokeFees

		invokeInfoBytes, err := json.Marshal(invokeInfo)
		if err != nil {
			return nil, err
		}
		txArgs = append(txArgs, invokeInfoBytes)

	} else {
		invokeInfoBytes, err := json.Marshal(invokeInfo)
		if err != nil {
			return nil, err
		}
		txArgs = append(txArgs, invokeInfoBytes)
	}
	txArgs = append(txArgs, reqArgs...)
	return txArgs, nil
}

func handleArg1(tx *modules.Transaction, reqArgs [][]byte) ([][]byte, error) {
	if len(reqArgs) <= 1 {
		return nil, fmt.Errorf("[%s]handlemsg1 req args error", shortId(tx.RequestHash().String()))
	}
	newReqArgs := [][]byte{}
	newReqArgs = append(newReqArgs, reqArgs[0])
	newReqArgs = append(newReqArgs, tx.CertId)
	newReqArgs = append(newReqArgs, reqArgs[1:]...)
	return newReqArgs, nil
}

func checkAndAddTxSigMsgData(local *modules.Transaction, recv *modules.Transaction) (bool, error) {
	var recvSigMsg *modules.Message
	if local == nil {
		log.Info("checkAndAddTxSigMsgData, local sig msg not exist")
		return false, nil
	}
	if recv == nil {
		return false, errors.New("checkAndAddTxSigMsgData param is nil")
	}
	reqId := local.RequestHash()

	if len(local.TxMessages) != len(recv.TxMessages) {
		return false, fmt.Errorf("[%s]checkAndAddTxSigMsgData tx msg is invalid,local msg len[%d],recv msg len[%d]",
			shortId(reqId.String()), len(local.TxMessages), len(recv.TxMessages))
	}
	for i := 0; i < len(local.TxMessages); i++ {
		if recv.TxMessages[i].App == modules.APP_SIGNATURE {
			recvSigMsg = recv.TxMessages[i]
		} else if !local.TxMessages[i].CompareMessages(recv.TxMessages[i]) {
			log.Info("checkAndAddTxSigMsgData", "reqId", shortId(reqId.String()), "local:", local.TxMessages[i],
				"recv:", recv.TxMessages[i])
			return false, fmt.Errorf("[%s]checkAndAddTxSigMsgData tx msg[%d] is not equal", shortId(reqId.String()), i)
		}
	}
	if recvSigMsg == nil {
		return false, fmt.Errorf("[%s]checkAndAddTxSigMsgData not find recv sig msg", shortId(reqId.String()))
	}
	for i, msg := range local.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(*modules.SignaturePayload)
			sigs := sigPayload.Signatures
			for _, sig := range sigs {
				if bytes.Equal(sig.PubKey, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].PubKey) &&
					bytes.Equal(sig.Signature, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].Signature) {
					log.Infof("[%s]checkAndAddTxSigMsgData tx  already receive, tx[%s]", shortId(reqId.String()), recv.Hash().String())
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(*modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0])
			}
			local.TxMessages[i].Payload = sigPayload
			log.Infof("[%s]checkAndAddTxSigMsgData, add sig payload, len(Signatures)=%d, Signatures[%s]",
				shortId(reqId.String()), len(sigPayload.Signatures), sigPayload.Signatures)
			return true, nil
		}
	}
	return false, fmt.Errorf("[%s]checkAndAddTxSigMsgData fail", shortId(reqId.String()))
}

func getTxSigNum(tx *modules.Transaction) int {
	if tx != nil {
		for _, msg := range tx.TxMessages {
			if msg.App == modules.APP_SIGNATURE {
				return len(msg.Payload.(*modules.SignaturePayload).Signatures)
			}
		}
	}
	return 0
}

func (p *Processor) checkTxIsExist(tx *modules.Transaction) bool {
	return p.validator.CheckTxIsExist(tx)
}

func (p *Processor) checkTxReqIdIsExist(reqId common.Hash) bool {
	id, err := p.dag.GetTxHashByReqId(reqId)
	if err == nil && id != (common.Hash{}) {
		return true
	}
	return false
}

func (p *Processor) checkTxValid(tx *modules.Transaction) bool {
	reqId := tx.RequestHash()
	txHash := tx.Hash()
	_, _, err := p.validator.ValidateTx(tx, false)
	if err != nil {
		log.Debugf("[%s]checkTxValid, Validate fail, txHash[%s], err:%s",
			shortId(reqId.String()), txHash.String(), err.Error())
		return false
	}
	if !checkContractTxFeeValid(p.dag, tx) {
		log.Debugf("[%s]checkTxValid, checkContractTxFeeValid fail, txHash[%s]",
			shortId(reqId.String()), txHash.String())
		return false
	}
	return true
}

func (p *Processor) checkTxAddrValid(tx *modules.Transaction) bool {
	reqId := tx.RequestHash()
	cType, err := getContractTxType(tx)
	if err != nil {
		log.Infof("[%s]checkTxAddrValid, getContractTxType fail", shortId(reqId.String()))
		return false
	}
	reqAddr, err := p.dag.GetTxRequesterAddress(tx)
	if err != nil {
		log.Infof("[%s]checkTxAddrValid, GetTxRequesterAddress fail", shortId(reqId.String()))
		return false
	}
	switch cType {
	case modules.APP_CONTRACT_TPL_REQUEST:
		return p.dag.IsContractDeveloper(reqAddr)
	case modules.APP_CONTRACT_DEPLOY_REQUEST:
	case modules.APP_CONTRACT_INVOKE_REQUEST:
	case modules.APP_CONTRACT_STOP_REQUEST:
		contractId := tx.ContractIdBytes()
		contract, err := p.dag.GetContract(contractId)
		if err != nil {
			log.Debugf("[%s]checkTxAddrValid, GetContract fail, contractId[%v]", shortId(reqId.String()), contractId)
			return false
		}
		reqAddr, err := p.dag.GetTxRequesterAddress(tx)
		if err != nil {
			return false
		}
		if !bytes.Equal(contract.Creator, reqAddr.Bytes()) {
			log.Debugf("[%s]checkTxAddrValid, addr is not equal, Creator[%v], reqAddr[%v]",
				shortId(reqId.String()), contract.Creator, reqAddr.Bytes())
			return false
		}
	}
	return true
}

func checkTxReceived(all []*modules.Transaction, tx *modules.Transaction) bool {
	if len(all) < 1 {
		return false
	}
	if tx == nil {
		return true
	}
	inHash := tx.Hash()
	for _, local := range all {
		if bytes.Equal(inHash.Bytes(), local.Hash().Bytes()) {
			return true
		}
	}
	return false
}

func msgsCompare(msgsA []*modules.Message, msgsB []*modules.Message, msgType modules.MessageType) bool {
	if msgsA == nil || msgsB == nil {
		log.Error("msgsCompare", "param is nil")
		return false
	}
	var msg1, msg2 *modules.Message
	for _, v := range msgsA {
		if v.App == msgType {
			msg1 = v
		}
	}
	for _, v := range msgsB {
		if v.App == msgType {
			msg2 = v
		}
	}
	if msg1 != nil && msg2 != nil {
		if msg1.CompareMessages(msg2) {
			log.Debug("msgsCompare,msg is equal.", "type", msgType)
			return true
		}
	}
	log.Debug("msgsCompare,msg is not equal", "msg1", msg1.Payload, "msg2", msg2.Payload) //todo del
	return false
}

//
//func printTxInfo(tx *modules.Transaction) {
//	if tx == nil {
//		return
//	}
//
//	log.Debug("=========tx info============hash:", tx.Hash().String())
//	for i := 0; i < len(tx.TxMessages); i++ {
//		log.Debug("---------")
//		app := tx.TxMessages[i].App
//		pay := tx.TxMessages[i].Payload
//		log.Debug("", "app:", app)
//		if app == modules.APP_PAYMENT {
//			p := pay.(*modules.PaymentPayload)
//			log.Debugf("%d", p.LockTime)
//		} else if app == modules.APP_CONTRACT_INVOKE_REQUEST {
//			p := pay.(*modules.ContractInvokeRequestPayload)
//			log.Debugf("%x", p.ContractId)
//		} else if app == modules.APP_CONTRACT_INVOKE {
//			p := pay.(*modules.ContractInvokePayload)
//			log.Debugf("%x", p.Args)
//			for idx, v := range p.WriteSet {
//				log.Debugf("WriteSet:idx[%d], k[%v]-v[%v]\n", idx, v.Key, v.Value)
//			}
//			for idx, v := range p.ReadSet {
//				log.Debugf("ReadSet:idx[%d], k[%v]-v[%v]\n", idx, v.Key, v.ContractId)
//			}
//		} else if app == modules.APP_SIGNATURE {
//			p := pay.(*modules.SignaturePayload)
//			log.Debugf("Signatures:[%v]", p.Signatures)
//		} else if app == modules.APP_DATA {
//			p := pay.(*modules.DataPayload)
//			log.Debugf("Text:[%v]", p.MainData)
//		}
//	}
//}

func getContractTxType(tx *modules.Transaction) (modules.MessageType, error) {
	if tx == nil {
		return modules.APP_UNKNOW, errors.New("getContractTxType get param is nil")
	}
	for _, msg := range tx.TxMessages {
		if msg.App >= modules.APP_CONTRACT_TPL_REQUEST && msg.App <= modules.APP_CONTRACT_STOP_REQUEST {
			return msg.App, nil
		}
	}
	return modules.APP_UNKNOW, fmt.Errorf("getContractTxType not contract Tx, txHash[%s]", shortId(tx.Hash().String()))
}

func getContractTxContractInfo(tx *modules.Transaction, msgType modules.MessageType) (interface{}, error) {
	if tx == nil {
		return nil, errors.New("getContractTxType get param is nil")
	}
	for _, msg := range tx.TxMessages {
		if msg.App == msgType {
			return msg.Payload, nil
		}
	}
	log.Debugf("[%s]getContractTxContractInfo,  not find msgType[%v]", shortId(tx.RequestHash().String()), msgType)
	return nil, nil
}

func getElectionSeedData(in common.Hash) []byte {
	addr := crypto.RequestIdToContractAddress(in)
	return addr.Bytes()
}

func conversionElectionSeedData(in []byte) []byte {
	return in
}

/*
Preliminary test conclusion
expectNum:4
use:
total    weight    num
20         4        5
50         7        7
100        8        5
200        15       5
500        17       6
1000       18       5
*/
func electionWeightValue(total uint64) (val uint64) {
	if total <= 20 {
		return 4
	} else if total > 20 && total <= 50 {
		return 7
	} else if total > 50 && total <= 100 {
		return 8
	} else if total > 100 && total <= 200 {
		return 15
	} else if total > 200 && total <= 500 {
		return 17
	} else if total > 500 {
		return 20
	}
	return 4
}

func shortId(id string) string {
	if len(id) < 8 {
		return id
	}
	return id[0:8]
}

func getValidAddress(addrs []common.Address) []common.Address {
	result := make([]common.Address, 0)
	for _, a := range addrs {
		find := false
		for _, b := range result {
			if a.Equal(b) {
				find = true
			}
		}
		if !find {
			result = append(result, a)
		}
	}
	return result
}

func checkJuryCountValid(numIn, numLocal uint64) bool {
	if numLocal == 0 {
		log.Error("checkJuryCountValid, numLocal is 0")
		return false
	}
	if int(math.Abs(float64(numIn-numLocal))*100/float64(numLocal)) <= 10 {
		return true
	}
	log.Error("checkJuryCountValid", "numIn", numIn, "numLocal", numLocal)
	return false
}

//根据交易和费用类型，获取费用基数(dao),其中计算单位
//timeout:秒
//size:字节
func getContractFeeLevel(dag iDag, msg modules.MessageType, feeType int) (level float64) {
	level = 1 //todo
	cp := dag.GetChainParameters()
	var opFee float64
	timeFee := float64(cp.ContractTxTimeoutUnitFee)
	sizeFee := float64(cp.ContractTxSizeUnitFee)
	switch msg {
	case modules.APP_CONTRACT_TPL_REQUEST:
		opFee = cp.ContractTxInstallFeeLevel
	case modules.APP_CONTRACT_DEPLOY_REQUEST:
		opFee = cp.ContractTxDeployFeeLevel
	case modules.APP_CONTRACT_INVOKE_REQUEST:
		opFee = cp.ContractTxInvokeFeeLevel
	case modules.APP_CONTRACT_STOP_REQUEST:
		opFee = cp.ContractTxStopFeeLevel
	}
	if feeType == ContractFeeTypeTimeOut {
		level = opFee * timeFee
	} else if feeType == ContractFeeTypeTxSize {
		level = opFee * sizeFee
	}
	return level
}

func getContractTxNeedFee(dag iDag, msgType modules.MessageType, timeout float64, txSize float64) (timeFee float64, sizeFee float64) {
	timeoutLevel := getContractFeeLevel(dag, msgType, ContractFeeTypeTimeOut)
	sizeLevel := getContractFeeLevel(dag, msgType, ContractFeeTypeTxSize)

	timeFee = timeoutLevel * timeout
	sizeFee = sizeLevel * txSize
	return timeFee, sizeFee
}

func checkContractTxFeeValid(dag iDag, tx *modules.Transaction) bool {
	if tx == nil {
		return false
	}
	var timeout uint32
	reqId := tx.RequestHash()
	txSize := tx.Size().Float64()
	fees, err := dag.GetTxFee(tx)
	if err != nil {
		log.Errorf("[%s]checkContractTxFeeValid, GetTxFee fail", shortId(reqId.String()))
		return false
	}
	txType, err := getContractTxType(tx)
	if err != nil {
		log.Errorf("[%s]checkContractTxFeeValid, getContractTxType fail", shortId(reqId.String()))
		return false
	}
	switch txType {
	case modules.APP_CONTRACT_TPL_REQUEST: //todo
	case modules.APP_CONTRACT_DEPLOY_REQUEST: //todo
	case modules.APP_CONTRACT_INVOKE_REQUEST:
		payload, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_INVOKE_REQUEST)
		if err != nil {
			log.Errorf("[%s]checkContractTxFeeValid, getContractTxContractInfo fail", shortId(reqId.String()))
			return false
		}
		timeout = payload.(*modules.ContractInvokeRequestPayload).Timeout
	case modules.APP_CONTRACT_STOP_REQUEST: //todo
	}
	timeFee, sizeFee := getContractTxNeedFee(dag, txType, float64(timeout), txSize)
	// TODO
	//if err != nil {
	//	log.Errorf("[%s]checkContractTxFeeValid, getContractTxNeedFee fail", shortId(reqId.String()))
	//	return false
	//}
	val := math.Max(float64(fees.Amount), timeFee+sizeFee) == float64(fees.Amount)
	if !val {
		log.Errorf("[%s]checkContractTxFeeValid invalid, fee amount[%f]-fees[%f](%f + %f), txSize[%f], timeout[%d]",
			shortId(reqId.String()), float64(fees.Amount), timeFee+sizeFee, timeFee, sizeFee, txSize, timeout)
	}
	return val
}

func calculateContractDeployDuringTime(dag iDag, tx *modules.Transaction) (uint64, error) {
	if tx == nil {
		return 0, errors.New("calculateContractDeployDuringTime, param is nil")
	}
	txSize := tx.Size()
	fees, err := dag.GetTxFee(tx)
	if err != nil {
		return 0, errors.New("calculateContractDeployDuringTime, GetTxFee fail")
	}
	sizeLevel := getContractFeeLevel(dag, modules.APP_CONTRACT_DEPLOY_REQUEST, ContractFeeTypeTxSize)
	timeLevel := getContractFeeLevel(dag, modules.APP_CONTRACT_DEPLOY_REQUEST, ContractFeeTypeTimeOut)

	sizeFee := sizeLevel * float64(txSize)
	timeFee := float64(fees.Amount) - sizeFee

	if timeLevel == 0 {
		//default
		timeLevel = 10
	}
	duringTime := timeFee / timeLevel
	log.Debug("calculateContractDeployDuringTime", "sizeLevel", sizeLevel, "timeLevel", timeLevel, "sizeFee", sizeFee, "timeFee", timeFee, "duringTime", duringTime)

	return uint64(duringTime), nil
}

func addContractDeployDuringTime(dag iDag, tx *modules.Transaction) error {
	if tx == nil {
		return errors.New("calculateContractDeployDuringTime, param is nil")
	}
	txType, _ := getContractTxType(tx)
	if txType != modules.APP_CONTRACT_DEPLOY_REQUEST {
		return nil
	}
	duringTime, err := calculateContractDeployDuringTime(dag, tx)
	if err != nil {
		return fmt.Errorf("addContractDeployDuringTime, err:%s", err)
	}
	payload, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_DEPLOY)
	if err != nil {
		return errors.New("addContractDeployDuringTime, getContractTxContractInfo fail")
	}
	payload.(*modules.ContractDeployPayload).DuringTime = duringTime

	return nil
}
