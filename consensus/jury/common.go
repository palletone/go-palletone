package jury

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts"
	//"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
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

			if len(sigs) < CONTRACT_SIG_NUM {
				return false
			}
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

//执行合约命令:install、deploy、invoke、stop，同时只支持一种类型
func runContractCmd(dag iDag, contract *contracts.Contract, trs *modules.Transaction) (modules.MessageType, []*modules.Message, error) {
	if trs == nil || len(trs.TxMessages) <= 0 {
		return 0, nil, errors.New("runContractCmd transaction or msg is nil")
	}
	for _, msg := range trs.TxMessages {
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
				}
				installResult, err := ContractProcess(contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err.Error())
					return msg.App, nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_TPL_REQUEST txid(%s) err:%s", req.ccName, err))
				}
				payload := installResult.(*modules.ContractTplPayload)
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_TPL, payload))
				return modules.APP_CONTRACT_TPL, msgs, nil
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractDeployRequestPayload)
				req := ContractDeployReq{
					chainID:    "palletone",
					templateId: reqPay.TplId,
					txid:       reqPay.TxId,
					args:       reqPay.Args,
					timeout:    reqPay.Timeout,
				}
				deployResult, err := ContractProcess(contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err.Error())
					return msg.App, nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_DEPLOY_REQUEST TplId(%s) err:%s", req.templateId, err))
				}
				payload := deployResult.(*modules.ContractDeployPayload)
				msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_DEPLOY, payload))
				return modules.APP_CONTRACT_DEPLOY, msgs, nil
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractInvokeRequestPayload)
				req := ContractInvokeReq{
					chainID:  "palletone",
					deployId: reqPay.ContractId,
					args:     reqPay.Args,
					txid:     trs.RequestHash().String(),
				}
				//对msg0进行修改
				fullArgs, err := handleMsg0(trs, dag, req.args)
				if err != nil {
					return modules.APP_CONTRACT_INVOKE, nil, err
				}
				req.args = fullArgs
				invokeResult, err := ContractProcess(contract, req)
				if err != nil {
					log.Error("runContractCmd ContractProcess", "ContractProcess error", err.Error())
					return msg.App, nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_INVOKE txid(%s) rans err:%s", req.txid, err))
				}
				result := invokeResult.(*modules.ContractInvokeResult)
				payload := modules.NewContractInvokePayload(result.ContractId, result.FunctionName, result.Args, 0 /*result.ExecutionTime*/, result.ReadSet, result.WriteSet, result.Payload)

				if payload != nil {
					msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_INVOKE, payload))
				}
				toContractPayments, err := resultToContractPayments(dag, result)
				if err != nil {
					return modules.APP_CONTRACT_INVOKE, nil, err
				}
				if toContractPayments != nil && len(toContractPayments) > 0 {
					for _, contractPayment := range toContractPayments {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, contractPayment))
					}
				}
				cs, err := resultToCoinbase(result)
				if err != nil {
					return modules.APP_CONTRACT_INVOKE, nil, err
				}
				if cs != nil && len(cs) > 0 {
					for _, coinbase := range cs {
						msgs = append(msgs, modules.NewMessage(modules.APP_PAYMENT, coinbase))
					}
				}
				return modules.APP_CONTRACT_INVOKE, msgs, nil
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			{
				msgs := []*modules.Message{}
				reqPay := msg.Payload.(*modules.ContractStopRequestPayload)
				req := ContractStopReq{
					chainID:     "palletone",
					deployId:    reqPay.ContractId,
					txid:        reqPay.Txid,
					deleteImage: reqPay.DeleteImage,
				}
				_, err := ContractProcess(contract, req) //todo
				if err != nil {
					log.Error("runContractCmd ContractProcess ", "error", err.Error())
					return msg.App, nil, errors.New(fmt.Sprintf("runContractCmd APP_CONTRACT_STOP_REQUEST contractId(%s) err:%s", req.deployId, err))
				}
				//payload := stopResult.(*modules.ContractStopPayload)
				//msgs = append(msgs, modules.NewMessage(modules.APP_CONTRACT_STOP, payload))
				return modules.APP_CONTRACT_STOP, msgs, nil
			}
		}
	}

	return 0, nil, errors.New(fmt.Sprintf("runContractCmd err, txid=%s", trs.RequestHash().String()))
}

func handleMsg0(tx *modules.Transaction, dag iDag, reqArgs [][]byte) ([][]byte, error) {
	var txArgs [][]byte
	invokeInfo := modules.InvokeInfo{}
	if len(tx.TxMessages) > 0 {
		msg0 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
		invokeAddr, err := dag.GetAddrByOutPoint(msg0.Inputs[0].PreviousOutPoint)
		if err != nil {
			return nil, err
		}
		//如果是交付保证金
		//if string(reqArgs[0]) == "DepositWitnessPay" {
		invokeTokens := &modules.InvokeTokens{}
		outputs := msg0.Outputs
		invokeTokens.Asset = outputs[0].Asset
		for _, output := range outputs {
			addr, err := tokenengine.GetAddressFromScript(output.PkScript)
			if err != nil {
				return nil, err
			}
			contractAddr, err := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM")
			if err != nil {
				return nil, err
			}
			if addr.Equal(contractAddr) {
				invokeTokens.Amount += output.Value
			}
		}
		invokeInfo.InvokeTokens = invokeTokens
		//}
		invokeFees, err := dag.GetTxFee(tx)
		if err != nil {
			return nil, err
		}

		invokeInfo.InvokeAddress = invokeAddr.String()
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

func checkAndAddTxData(local *modules.Transaction, recv *modules.Transaction) (bool, error) {
	var recvSigMsg *modules.Message

	if local == nil || recv == nil {
		return false, errors.New("checkAndAddTxData param is nil")
	}
	if len(local.TxMessages) != len(recv.TxMessages) {
		return false, errors.New("checkAndAddTxData tx msg is invalid")
	}
	for i := 0; i < len(local.TxMessages); i++ {
		if recv.TxMessages[i].App == modules.APP_SIGNATURE {
			recvSigMsg = recv.TxMessages[i]
		} else if !local.TxMessages[i].CompareMessages(recv.TxMessages[i]) {
			return false, errors.New("checkAndAddTxData tx msg is not equal")
		}
	}

	if recvSigMsg == nil {
		return false, errors.New("checkAndAddTxData not find recv sig msg")
	}
	for i, msg := range local.TxMessages {
		if msg.App == modules.APP_SIGNATURE {
			sigPayload := msg.Payload.(*modules.SignaturePayload)
			sigs := sigPayload.Signatures
			for _, sig := range sigs {
				if true == bytes.Equal(sig.PubKey, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].PubKey) &&
					true == bytes.Equal(sig.Signature, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0].Signature) {
					log.Info("checkAndAddTxData tx  already recv:", recv.RequestHash().String())
					return false, nil
				}
			}
			//直接将签名添加到msg中
			if len(recvSigMsg.Payload.(*modules.SignaturePayload).Signatures) > 0 {
				sigPayload.Signatures = append(sigs, recvSigMsg.Payload.(*modules.SignaturePayload).Signatures[0])
			}
			local.TxMessages[i].Payload = sigPayload
			log.Info("checkAndAddTxData", "add sig payload:", sigPayload.Signatures)
			return true, nil
		}
	}

	return false, errors.New("checkAndAddTxData fail")
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

func checkTxValid(tx *modules.Transaction) bool {
	if tx == nil {
		return false
	}
	//var sigs []modules.SignatureSet
	//tmpTx := &modules.Transaction{}
	////todo 检查msg的有效性
	//
	//for _, msg := range tx.TxMessages {
	//	if msg.App == modules.APP_SIGNATURE {
	//		sigs = msg.Payload.(*modules.SignaturePayload).Signatures
	//	} else {
	//		tmpTx.TxMessages = append(tmpTx.TxMessages, msg)
	//	}
	//}
	//printTxInfo(tmpTx)
	//TODO VerifyTXWithPK has a bug, fix it soon...
	//if len(sigs) > 0 {
	//	for i := 0; i < len(sigs); i++ {
	//
	//		if !keystore.VerifyTXWithPK(sigs[i].Signature, tmpTx, sigs[i].PubKey) {
	//			log.Error("ValidateTxSig", "VerifyTXWithPK sig fail!!!!", tmpTx.RequestHash().String())
	//			//log.Debug("--ValidateTxSig", "tx info:", tmpTx)
	//			//log.Debug("--ValidateTxSig", "sigSet info:", sigs[i])
	//			//return false
	//			return true
	//		}
	//	}
	//}

	return true
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

func isSystemContract(tx *modules.Transaction) bool {
	for _, msg := range tx.TxMessages {
		if msg.App == modules.APP_CONTRACT_INVOKE_REQUEST {
			contractId := msg.Payload.(*modules.ContractInvokeRequestPayload).ContractId
			log.Debug("isSystemContract", "contract id", contractId, "len", len(contractId))
			contractAddr := common.NewAddress(contractId, common.ContractHash)
			return contractAddr.IsSystemContractAddress() //, nil

		} else if msg.App == modules.APP_CONTRACT_TPL_REQUEST {
			return true //todo  先期将install作为系统合约处理，只有Mediator可以安装，后期在扩展到所有节点
		} else if msg.App >= modules.APP_CONTRACT_DEPLOY_REQUEST {
			return false //, nil
		}
	}
	return true //, errors.New("isSystemContract not find contract type")
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
		return modules.APP_UNKNOW, errors.New("getContractTxType get param is nil")
	}
	for _, msg := range tx.TxMessages {
		if msg.App == msgType {
			return msg.Payload, nil
		}
	}
	return modules.APP_UNKNOW, errors.New("getContractTxContractInfo not find")
}
