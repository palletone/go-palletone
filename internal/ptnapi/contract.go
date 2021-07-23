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
	"bytes"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/dag/dagconfig"
)

const GOLANG = "golang"
const GO = "go"

type buildContractContext struct {
	tokenId    string
	password   string
	fromAddr   common.Address
	toAddr     common.Address
	ccAddr     common.Address
	amount     decimal.Decimal
	gasFee     decimal.Decimal
	args       [][]byte
	exeTimeout *Int
}
type buildMutiContractContext struct {
	tokenId1   string
	tokenId2   string
	password   string
	fromAddr   common.Address
	toAddr     common.Address
	ccAddr     common.Address
	amount1    decimal.Decimal
	amount2    decimal.Decimal
	gasFee     decimal.Decimal
	args       [][]byte
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
	if ctx == nil || msgReq == nil {
		return nil, errors.New("buildContractReqTx, param is nil")
	}

	//如没有GasFee，而且to address不是合约地址，则不构建Payment，直接InvokeRequest+Signature
	if s.b.EnableGasFee() || ctx.toAddr == ctx.ccAddr || ctx.fromAddr != ctx.toAddr {
		var usedUtxo []*modules.UtxoWithOutPoint
		//费用检查
		var fee decimal.Decimal
		// 当支付给合约，但是没有交易费
		if s.b.EnableGasFee() {
			var err error
			fee, err = s.contractFeeCheck(ctx, msgReq)
			if err != nil {
				log.Errorf("buildContractReqTx, contractFeeCheck err:%s", err.Error())
				return nil, err
			}
		}
		//fee, err := s.contractFeeCheck(ctx, msgReq)
		//if err != nil {
		//	log.Errorf("buildContractReqTx, contractFeeCheck err:%s", err.Error())
		//	return nil, err
		//}
		//build raw tx
		tx, usedUtxo, err = buildRawTransferTx(s.b, ctx.tokenId, ctx.fromAddr.String(), ctx.toAddr.String(), ctx.amount, fee, ctx.password)
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
func (s *PrivateContractAPI) buildMutiContractReqTx(ctx *buildMutiContractContext, msgReq *modules.Message) (*modules.Transaction, error) {
	var tx *modules.Transaction
	var tx2 *modules.Transaction
	var err error
	if ctx == nil || msgReq == nil {
		return nil, errors.New("buildContractReqTx, param is nil")
	}
	totalamount := ctx.amount1.Add(ctx.amount2)
	//如没有GasFee，而且to address不是合约地址，则不构建Payment，直接InvokeRequest+Signature
	if s.b.EnableGasFee() || ctx.toAddr == ctx.ccAddr || ctx.fromAddr != ctx.toAddr {
		var usedUtxo []*modules.UtxoWithOutPoint
		var usedUtxo2 []*modules.UtxoWithOutPoint
		//费用检查
		ctx4check := &buildContractContext{
			tokenId:    ctx.tokenId1,
			fromAddr:   ctx.fromAddr,
			toAddr:     ctx.toAddr,
			ccAddr:     ctx.ccAddr,
			amount:     totalamount,
			gasFee:     ctx.gasFee,
			args:       ctx.args,
			password:   ctx.password,
			exeTimeout: ctx.exeTimeout,
		}
		var fee decimal.Decimal
		// 当支付给合约，但是没有交易费
		if s.b.EnableGasFee() {
			var err error
			fee, err = s.contractFeeCheck(ctx4check, msgReq)
			if err != nil {
				log.Errorf("buildContractReqTx, contractFeeCheck err:%s", err.Error())
				return nil, err
			}
		}
		//fee, err := s.contractFeeCheck(ctx4check, msgReq)
		//if err != nil {
		//	log.Errorf("buildContractReqTx, contractFeeCheck err:%s", err.Error())
		//	return nil, err
		//}
		//build raw tx
		tx, usedUtxo, err = buildRawTransferTx(s.b, ctx.tokenId1, ctx.fromAddr.String(), ctx.toAddr.String(), ctx.amount1, fee, ctx.password)
		if err != nil {
			return nil, err
		}
		tx2, usedUtxo2, err = buildRawTransferTx(s.b, ctx.tokenId2, ctx.fromAddr.String(), ctx.toAddr.String(), ctx.amount2, fee, ctx.password)
		if err != nil {
			return nil, err
		}
		usedUtxo = append(usedUtxo, usedUtxo2...)
		tx.AddMessage(tx2.TxMessages()[1])

		tx.AddMessage(msgReq)
		//sign
		err = signRawTransaction(s.b, tx, ctx.fromAddr.String(), ctx.password, ctx.exeTimeout, 1, usedUtxo)
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("buildContractReqTx, disabled gas fee, to address[%s],amount[%s] and fee[%s] will ignore.", ctx.toAddr.String(), totalamount.String(), ctx.gasFee.String())
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

func (s *PrivateContractAPI) contractFeeCheck(ctx *buildContractContext, reqMsg *modules.Message) (decimal.Decimal, error) {
	if ctx == nil || reqMsg == nil {
		return decimal.NewFromFloat(0), fmt.Errorf("contractFeeCheck param ctx is nil")
	}
	//return ctx.gasFee, nil

	var err error
	fee := ctx.gasFee

	//assetId, _, err := modules.String2AssetId(ctx.tokenId)
	assetId := dagconfig.DagConfig.GetGasToken()
	//if err != nil {
	//	log.Errorf("contractFeeCheck,  String2AssetId err:%s:%s", err.Error(), ctx.tokenId)
	//	return fee, fmt.Errorf("contractFeeCheck, String2AssetId err:%s", err.Error())
	//}
	//baseFee := decimal.NewFromFloat(float64(s.b.Dag().GetChainParameters().TransferPtnBaseFee))
	//if ctx.gasFee.Cmp(baseFee) < 0 { //ctx.gasFee < s.b.Dag().GetChainParameters().TransferPtnBaseFee

	if !ctx.gasFee.GreaterThan(decimal.Zero) {
		var needFee float64
		switch reqMsg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			payload := reqMsg.Payload.(*modules.ContractInstallRequestPayload)
			needFee, _, _, err = s.b.ContractInstallReqTxFee(ctx.fromAddr, ctx.toAddr, assetId.Uint64Amount(ctx.amount), assetId.Uint64Amount(ctx.gasFee),
				payload.TplName, payload.Path, payload.Version, payload.TplDescription, payload.Abi, payload.Language, nil)
			needFee = needFee * 2
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			payload := reqMsg.Payload.(*modules.ContractDeployRequestPayload)
			needFee, _, _, err = s.b.ContractDeployReqTxFee(ctx.fromAddr, ctx.toAddr,
				assetId.Uint64Amount(ctx.amount), 0, payload.TemplateId, ctx.args, payload.ExtData, 0)
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			payload := reqMsg.Payload.(*modules.ContractInvokeRequestPayload)
			needFee, _, _, err = s.b.ContractInvokeReqTxFee(ctx.fromAddr, ctx.toAddr, assetId.Uint64Amount(ctx.amount), assetId.Uint64Amount(ctx.gasFee),
				nil, common.NewAddress(payload.ContractId, common.ContractHash), ctx.args, 0)
		case modules.APP_CONTRACT_STOP_REQUEST:
			payload := reqMsg.Payload.(*modules.ContractStopRequestPayload)
			needFee, _, _, err = s.b.ContractStopReqTxFee(ctx.fromAddr, ctx.toAddr, assetId.Uint64Amount(ctx.amount), assetId.Uint64Amount(ctx.gasFee),
				common.NewAddress(payload.ContractId, common.ContractHash), false)
		}
		if err != nil {
			return fee, fmt.Errorf("contractFeeCheck, contract fee get err:%s", err.Error())
		}
		fee = assetId.DisplayAmount(uint64(needFee))
	}

	//dNeedFee := assetId.DisplayAmount(uint64(needFee)) //保证交易费用的充足，所以设定值为需要费用的2倍
	////dNeedFee := ptnjson.Dao2Ptn(uint64(needFee))
	////如果设定费用<=0，则由程序计算费用。如果设定>0，则进行费用比较，不足则直接返回错误，费用够则使用用户设置费用
	//if ctx.gasFee.GreaterThan(decimal.Zero) { // gasFee> 0
	//	if ctx.gasFee.LessThan(dNeedFee) {
	//		log.Warnf("contractFeeCheck, fee not enough, fee[%s], need fee[%s]",
	//			ctx.gasFee.String(), dNeedFee.String())
	//		fee = dNeedFee
	//	}
	//} else { // gasFee<=0
	//	fee = dNeedFee
	//}

	log.Debug("contractFeeCheck", "dynamic calculation fee:", fee.String(), "origin fee:", ctx.gasFee.String())
	return fee, nil
}
