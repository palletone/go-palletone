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
	"time"

	"encoding/hex"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/tokenengine"
)

const (
//VrfElectionNum = 4
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
			log.Debug("localIsMinSignature", "local sig", localSig)
			return true
		}
	}
	return false
}

func checkAndAddSigSet(local *modules.Transaction, recv *modules.Transaction) error {
	if local == nil || recv == nil {
		return errors.New("checkAndAddSigSet param is nil")
	}
	var app modules.MessageType
	for _, msg := range local.TxMessages {
		if msg.App >= modules.APP_CONTRACT_TPL && msg.App <= modules.APP_SIGNATURE {
			app = msg.App
			break
		}
	}
	if app <= 0 {
		return errors.New("checkAndAddSigSet not find contract app type")
	}
	if msgsCompare(local.TxMessages, recv.TxMessages, app) {
		getSigPay := func(mesgs []*modules.Message) *modules.SignaturePayload {
			for _, v := range mesgs {
				if v.App == modules.APP_SIGNATURE {
					return v.Payload.(*modules.SignaturePayload)
				}
			}
			return nil
		}
		localSigPay := getSigPay(local.TxMessages)
		recvSigPay := getSigPay(recv.TxMessages)
		if localSigPay != nil && recvSigPay != nil {
			localSigPay.Signatures = append(localSigPay.Signatures, recvSigPay.Signatures[0])
			log.Debug("checkAndAddSigSet", "local transaction", local.RequestHash(), "recv transaction", recv.RequestHash())
			return nil
		}
	}

	return errors.New("checkAndAddSigSet add sig fail")
}

func createContractErrorPayloadMsg(reqType modules.MessageType, contractReq interface{}, errMsg string) *modules.Message {
	err := modules.ContractError{
		Code:    500, //todo
		Message: errMsg,
	}
	switch reqType {
	case modules.APP_CONTRACT_TPL_REQUEST:
		req := contractReq.(ContractInstallReq)
		payload := modules.NewContractTplPayload(nil, req.ccName, req.ccPath, req.ccVersion, 0, nil, err)
		return modules.NewMessage(modules.APP_CONTRACT_TPL, payload)
	case modules.APP_CONTRACT_DEPLOY_REQUEST:
		req := contractReq.(ContractDeployReq)
		payload := modules.NewContractDeployPayload(req.templateId, nil, "", req.args, nil, nil, nil, err)
		return modules.NewMessage(modules.APP_CONTRACT_DEPLOY, payload)
	case modules.APP_CONTRACT_INVOKE_REQUEST:
		req := contractReq.(ContractInvokeReq)
		payload := modules.NewContractInvokePayload(req.deployId, "", req.args, 0, nil, nil, nil, err)
		return modules.NewMessage(modules.APP_CONTRACT_INVOKE, payload)
	case modules.APP_CONTRACT_STOP_REQUEST:
		req := contractReq.(ContractStopReq)
		payload := modules.NewContractStopPayload(req.deployId, nil, nil, err)
		return modules.NewMessage(modules.APP_CONTRACT_STOP, payload)
	}

	return nil
}

//执行合约命令:install、deploy、invoke、stop，同时只支持一种类型
func runContractCmd(rwM rwset.TxManager, dag iDag, contract *contracts.Contract, tx *modules.Transaction, elf []modules.ElectionInf, errMsgEnable bool) ([]*modules.Message, error) {
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
					chainID:   "palletone",
					ccName:    reqPay.TplName,
					ccPath:    reqPay.Path,
					ccVersion: reqPay.Version,
					addrHash:  reqPay.AddrHash,
				}
				installResult, err := ContractProcess(rwM, contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err.Error())
					if errMsgEnable {
						errMsg := createContractErrorPayloadMsg(modules.APP_CONTRACT_TPL_REQUEST, req, err.Error())
						msgs = append(msgs, errMsg)
						return msgs, nil
					}
					return nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_TPL_REQUEST txid(%s) err:%s", req.ccName, err))
				}
				payload := installResult.(*modules.ContractTplPayload)
				payload.AddrHash = req.addrHash
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_TPL, payload))
				return msgs, nil
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractDeployRequestPayload)
				req := ContractDeployReq{
					chainID:    "palletone",
					templateId: reqPay.TplId,
					txid:       hex.EncodeToString(common.BytesToAddress(tx.RequestHash().Bytes()).Bytes21()),
					args:       reqPay.Args,
					timeout:    time.Duration(reqPay.Timeout),
				}
				deployResult, err := ContractProcess(rwM, contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err.Error())
					if errMsgEnable {
						errMsg := createContractErrorPayloadMsg(modules.APP_CONTRACT_DEPLOY_REQUEST, req, err.Error())
						msgs = append(msgs, errMsg)
						return msgs, nil
					}
					return nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_DEPLOY_REQUEST TplId(%s) err:%s", req.templateId, err))
				}
				payload := deployResult.(*modules.ContractDeployPayload)
				if len(elf) > 0 {
					payload.EleList = elf
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
					log.Error("runContractCmd ContractProcess", "ContractProcess error", err.Error())
					if errMsgEnable {
						errMsg := createContractErrorPayloadMsg(modules.APP_CONTRACT_INVOKE_REQUEST, req, err.Error())
						msgs = append(msgs, errMsg)
						return msgs, nil
					}
					return nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_INVOKE txid(%s) rans err:%s", req.txid, err))
				}
				result := invokeResult.(*modules.ContractInvokeResult)
				payload := modules.NewContractInvokePayload(result.ContractId, result.FunctionName, result.Args, 0 /*result.ExecutionTime*/ , result.ReadSet, result.WriteSet, result.Payload, modules.ContractError{})
				if payload != nil {
					msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_INVOKE, payload))
				}
				toContractPayments, err := resultToContractPayments(dag, result)
				if err != nil {
					return nil, err
				}
				if toContractPayments != nil && len(toContractPayments) > 0 {
					for _, contractPayment := range toContractPayments {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, contractPayment))
					}
				}
				cs, err := resultToCoinbase(result)
				if err != nil {
					return nil, err
				}
				if cs != nil && len(cs) > 0 {
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
					log.Error("runContractCmd ContractProcess ", "error", err.Error())
					if errMsgEnable {
						errMsg := createContractErrorPayloadMsg(modules.APP_CONTRACT_STOP_REQUEST, req, err.Error())
						msgs = append(msgs, errMsg)
						return msgs, nil
					}
					return nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_STOP_REQUEST contractId(%s) err:%s", req.deployId, err))
				}
				payload := stopResult.(*modules.ContractStopPayload)
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_STOP, payload))
				return msgs, nil
			}
		}
	}

	return nil, errors.New(fmt.Sprintf("runContractCmd err, txid=%s", tx.RequestHash().String()))
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
				addr, err := tokenengine.GetAddressFromScript(output.PkScript)
				if err != nil {
					return nil, err
				}
				if !addr.Equal(invokeAddr) { //note : only return not invokeAddr
					//
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
	//reqArgs = append(reqArgs, txArgs...)
	return txArgs, nil
}

func handleArg1(tx *modules.Transaction, reqArgs [][]byte) ([][]byte, error) {
	if len(reqArgs) <= 1 {
		return nil, fmt.Errorf("handlemsg1 req args error")
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
	if len(local.TxMessages) != len(recv.TxMessages) {
		return false, errors.New("checkAndAddTxSigMsgData tx msg is invalid")
	}
	for i := 0; i < len(local.TxMessages); i++ {
		if recv.TxMessages[i].App == modules.APP_SIGNATURE {
			recvSigMsg = recv.TxMessages[i]
		} else if !local.TxMessages[i].CompareMessages(recv.TxMessages[i]) {
			log.Info("checkAndAddTxSigMsgData", "local", local.TxMessages[i], "recv", recv.TxMessages[i])
			return false, errors.New("checkAndAddTxSigMsgData tx msg is not equal")
		}
	}

	if recvSigMsg == nil {
		return false, errors.New("checkAndAddTxSigMsgData not find recv sig msg")
	}
	for i, msg := range local.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(*modules.SignaturePayload)
			sigs := sigPayload.Signatures
			for _, sig := range sigs {
				if true == bytes.Equal(sig.PubKey, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].PubKey) &&
					true == bytes.Equal(sig.Signature, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].Signature) {
					log.Info("checkAndAddTxSigMsgData tx  already recv:", recv.RequestHash().String())
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(*modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0])
			}
			local.TxMessages[i].Payload = sigPayload
			log.Info("checkAndAddTxSigMsgData", "add sig payload:", sigPayload.Signatures)
			return true, nil
		}
	}

	return false, errors.New("checkAndAddTxSigMsgData fail")
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
	if p.validator.CheckTxIsExist(tx) {
		return false
	}
	return true
}

func (p *Processor) checkTxReqIdIsExist(reqId common.Hash) bool {
	id, err := p.dag.GetTxHashByReqId(reqId)
	if err == nil && id != (common.Hash{}) {
		return true
	}
	return false
}

func (p *Processor) checkTxValid(tx *modules.Transaction) bool {
	err := p.validator.ValidateTx(tx, false)
	if err != nil {
		log.Errorf("Validate tx[%s] throw an error:%s", tx.Hash().String(), err.Error())
	}

	return err == nil
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
			log.Debug("msgsCompare", "msg is equal, type", msgType)
			return true
		}
	}
	log.Debug("msgsCompare", "msg is not equal") //todo del
	return false
}

func printTxInfo(tx *modules.Transaction) {
	if tx == nil {
		return
	}

	log.Info("=========tx info============hash:", tx.Hash().String())
	for i := 0; i < len(tx.TxMessages); i++ {
		log.Info("---------")
		app := tx.TxMessages[i].App
		pay := tx.TxMessages[i].Payload
		log.Info("", "app:", app)
		if app == modules.APP_PAYMENT {
			p := pay.(*modules.PaymentPayload)
			fmt.Println(p.LockTime)
		} else if app == modules.APP_CONTRACT_INVOKE_REQUEST {
			p := pay.(*modules.ContractInvokeRequestPayload)
			fmt.Println(p.ContractId)
		} else if app == modules.APP_CONTRACT_INVOKE {
			p := pay.(*modules.ContractInvokePayload)
			fmt.Println(p.Args)
			for idx, v := range p.WriteSet {
				fmt.Printf("WriteSet:idx[%d], k[%v]-v[%v]\n", idx, v.Key, v.Value)
			}
			for idx, v := range p.ReadSet {
				fmt.Printf("ReadSet:idx[%d], k[%v]-v[%v]\n", idx, v.Key, v.Value)
			}
		} else if app == modules.APP_SIGNATURE {
			p := pay.(*modules.SignaturePayload)
			fmt.Printf("Signatures:[%v]", p.Signatures)
		} else if app == modules.APP_DATA {
			p := pay.(*modules.DataPayload)
			fmt.Printf("Text:[%v]", p.MainData)
		}
	}
}

func getFileHash(tx *modules.Transaction) []byte {
	if tx != nil {
		for _, msg := range tx.TxMessages {
			if msg.App == modules.APP_DATA {
				return msg.Payload.(*modules.DataPayload).MainData
			}
		}
	}

	return nil
}

func getContractTxType(tx *modules.Transaction) (modules.MessageType, error) {
	if tx == nil {
		return modules.APP_UNKNOW, errors.New("getContractTxType get param is nil")
	}
	for _, msg := range tx.TxMessages {
		if msg.App >= modules.APP_CONTRACT_TPL_REQUEST && msg.App <= modules.APP_CONTRACT_STOP_REQUEST {
			return msg.App, nil
		}
	}
	return modules.APP_UNKNOW, errors.New("getContractTxType not contract Tx")
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
	log.Debug("getContractTxContractInfo", "reqId", tx.RequestHash(), " not find msgType", msgType)
	return nil, nil
}

func getElectionSeedData(in common.Hash) ([]byte, error) {
	//rd := make([]byte, 20)
	//_, err := rand.Read(rd)
	//if err != nil {
	//	return nil, errors.New("getElectionData, rand fail")
	//}
	//seedData := make([]byte, len(in)+len(rd))
	//copy(seedData, in[:])
	//copy(seedData[len(in):], rd)
	//return seedData, nil

	return in.Bytes(), nil
}

func conversionElectionSeedData(in []byte) []byte {
	tmp := common.BytesToAddress(in)
	out := common.NewAddress(tmp.Bytes(), common.ContractHash)
	return out[:]
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
