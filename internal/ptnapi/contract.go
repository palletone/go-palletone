/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
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

const GOLANG = "golang"
const GO = "go"

type buildContractContext struct {
	msgType  modules.MessageType
	tokenId  string
	password string
	fromAddr common.Address
	toAddr   common.Address
	ccAddr   common.Address
	amount   decimal.Decimal
	gasFee   decimal.Decimal
	args     [][]byte

	exeTimeout *Int
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

	//如没有GasFee，而且to address不是合约地址，则不构建Payment，直接InvokeRequest+Signature
	if s.b.EnableGasFee() || ctx.toAddr == ctx.ccAddr || ctx.fromAddr != ctx.toAddr {
		var usedUtxo []*modules.UtxoWithOutPoint
		//build raw tx
		tx, usedUtxo, err = buildRawTransferTx(s.b, ctx.tokenId, ctx.fromAddr.String(), ctx.toAddr.String(), ctx.amount, ctx.gasFee, ctx.password)
		if err != nil {
			return nil, err
		}
		tx.AddMessage(msgReq)

		//sign
		err = signRawTransaction(s.b, tx, ctx.fromAddr.String(), ctx.password, ctx.exeTimeout, 1, usedUtxo)
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("buildContractReqTx, disabled gas fee, to address[%s],amount[%s] and fee[%s] will ignore.", ctx.toAddr.String(), ctx.amount.String(), ctx.gasFee.String())
		tx, err = s.buildContractReqTxWithoutGasFee(s.b, ctx.fromAddr, ctx.password, msgReq)
		if err != nil {
			return nil, err
		}
	}
	return tx, err
}

//创建没有Payment的合约请求交易
func (s *PrivateContractAPI) buildContractReqTxWithoutGasFee(b Backend, from common.Address,
	pwd string, msgReq *modules.Message) (*modules.Transaction, error) {
	tx := modules.NewTransaction([]*modules.Message{msgReq})
	return signRawNoGasTx(b, tx, from, pwd)
}

func (s *PrivateContractAPI) contractFeeCheck(enableGasFee bool, ctx *buildContractContext, reqMsg *modules.Message) (decimal.Decimal, error) {
	if ctx == nil {
		return decimal.NewFromFloat(0), fmt.Errorf("contractFeeCheck param ctx is nil")
	}
	var err error
	fee := ctx.gasFee
	if enableGasFee {
		baseFee := decimal.NewFromFloat(float64(s.b.Dag().GetChainParameters().TransferPtnBaseFee))
		if ctx.gasFee.Cmp(baseFee) < 0 { //ctx.gasFee < s.b.Dag().GetChainParameters().TransferPtnBaseFee
			var needFee float64
			switch ctx.msgType {
			case modules.APP_CONTRACT_TPL_REQUEST:
				payload := reqMsg.Payload.(*modules.ContractInstallRequestPayload)
				needFee, _, _, err = s.b.ContractInstallReqTxFee(ctx.fromAddr, ctx.toAddr, ptnjson.Ptn2Dao(ctx.amount), ptnjson.Ptn2Dao(ctx.gasFee),
					payload.TplName, payload.Path, payload.Version, payload.TplDescription, payload.Abi, payload.Language, nil)
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
				return fee, fmt.Errorf("Ccdeploytx, ContractDeployReqFee err:%s", err.Error())
			}

			fee = decimal.NewFromFloat(needFee + 1)
			log.Debug("Ccdeploytx", "dynamic calculation fee:", fee.String())
		}
	}
	return fee, nil
}
