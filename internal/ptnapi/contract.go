package ptnapi

import (
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/ptnjson"
	"fmt"
	"bytes"
	"github.com/palletone/go-palletone/common/crypto"
)

type buildContractContext struct {
	msgType  modules.MessageType
	tokenId  string
	password string
	fromAddr common.Address
	toAddr   common.Address
	amount   decimal.Decimal
	gasFee   decimal.Decimal
	args     [][]byte

	//install

	//deploy

	//invoke

	//stop

	exeTimeout Int
}

func getTemplateId(ccName, ccPath, ccVersion string) []byte {
	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := crypto.Keccak256Hash(buffer.Bytes())
	return tpid[:]
}

func (s *PrivateContractAPI) buildContractReqTx(ctx *buildContractContext, msgReq *modules.Message) (*modules.Transaction, error) {
	var tx *modules.Transaction
	var err error
	if s.b.EnableGasFee() {
		var usedUtxo []*modules.UtxoWithOutPoint
		//build raw tx
		tx, usedUtxo, err = buildRawTransferTx(s.b, ctx.tokenId, ctx.fromAddr.String(), ctx.toAddr.String(), ctx.amount, ctx.gasFee, ctx.password)
		if err != nil {
			return nil, err
		}

		tx.AddMessage(msgReq)
		//sign
		err = signRawTransaction(s.b, tx, ctx.fromAddr.String(), ctx.password, &ctx.exeTimeout, 1, usedUtxo)
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("buildContractReqTx, disabled gas fee, to address[%s],amount[%s] and fee[%s] will ignore.", ctx.toAddr.String(), ctx.amount.String(), ctx.gasFee.String())
		tx, err = s.buildContractReqTxWithoutGasFee(s.b, ctx.fromAddr, ctx.password, msgReq)
		//tx, err = s.buildCcinvokeTxWithoutGasFee(s.b, ctx.fromAddr, contractAddr, ctx.args, ctx.password, ctx.exeTimeout.Uint32())
		if err != nil {
			return nil, err
		}
	}

	return tx, err
}

//创建没有Payment的ccinvoketx
func (s *PrivateContractAPI) buildCcinvokeTxWithoutGasFee(b Backend, from,
contractAddr common.Address, args [][]byte, pwd string, exeTimeout uint32) (*modules.Transaction, error) {
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: contractAddr.Bytes(),
			Args:       args,
			Timeout:    exeTimeout,
		},
	}
	tx := modules.NewTransaction([]*modules.Message{msgReq})

	return signRawNoGasTx(b, tx, from, pwd)
}

//创建没有Payment的合约氢气交易
func (s *PrivateContractAPI) buildContractReqTxWithoutGasFee(b Backend, from common.Address,
	pwd string, msgReq *modules.Message) (*modules.Transaction, error) {
	tx := modules.NewTransaction([]*modules.Message{msgReq})
	return signRawNoGasTx(b, tx, from, pwd)
}

func (s *PrivateContractAPI) contractFeeCheck(enableGasFee bool, ctx *buildContractContext, reqMsg *modules.Message) (needFee decimal.Decimal, err error) {
	if ctx == nil {
		return decimal.NewFromFloat(0), fmt.Errorf("contractFeeCheck param ctx is nil")
	}
	daoFee := ctx.gasFee
	if enableGasFee {
		if ctx.gasFee.IsZero() { //dynamic calculation fee
			var needFee float64
			switch ctx.msgType {
			case modules.APP_CONTRACT_TPL_REQUEST:
				payload := reqMsg.Payload.(*modules.ContractInstallRequestPayload)
				needFee, _, _, err = s.b.ContractInstallReqTxFee(ctx.fromAddr, ctx.toAddr, ptnjson.Ptn2Dao(ctx.amount), ptnjson.Ptn2Dao(ctx.gasFee),
					payload.TplName, payload.Path, payload.Version,payload.TplDescription, payload.Abi, payload.Language, nil)
			case modules.APP_CONTRACT_DEPLOY_REQUEST:
				payload := reqMsg.Payload.(*modules.ContractDeployRequestPayload)
				needFee, _, _, err = s.b.ContractDeployReqTxFee(ctx.fromAddr, ctx.toAddr,
					ptnjson.Ptn2Dao(ctx.amount), 0, payload.TemplateId, ctx.args, payload.ExtData, 0)
			case modules.APP_CONTRACT_INVOKE_REQUEST:
				payload := reqMsg.Payload.(*modules.ContractInvokeRequestPayload)
				needFee, _, _, err = s.b.ContractInvokeReqTxFee(ctx.fromAddr, ctx.toAddr, ptnjson.Ptn2Dao(ctx.amount), ptnjson.Ptn2Dao(ctx.gasFee),
					nil, common.NewAddress(payload.ContractId, common.ContractHash), ctx.args, 0)
			case modules.APP_CONTRACT_STOP_REQUEST:
				payload := reqMsg.Payload.(*modules.ContractStopRequestPayload)
				needFee, _, _, err = s.b.ContractStopReqTxFee(ctx.fromAddr, ctx.toAddr, ptnjson.Ptn2Dao(ctx.amount), ptnjson.Ptn2Dao(ctx.gasFee),
					common.NewAddress(payload.ContractId, common.ContractHash), false)
			}
			if err != nil {
				return daoFee, fmt.Errorf("Ccdeploytx, ContractDeployReqFee err:%s", err.Error())
			}

			daoFee = decimal.NewFromFloat(needFee + 1)
			log.Debug("Ccdeploytx", "dynamic calculation fee:", daoFee.String())
		}
	}
	return daoFee, nil
}
