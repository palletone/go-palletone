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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/contracts/syscontract/sysconfigcc"
	"github.com/palletone/go-palletone/contracts/ucc"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/consensus/jury"
)

var (
	defaultMsg0 = []byte("query has no msg0")
	defaultMsg1 = []byte("query has no msg1")
)

type PublicContractAPI struct {
	b Backend
}

func NewPublicContractAPI(b Backend) *PublicContractAPI {
	//go synDag(b)
	return &PublicContractAPI{b}
}

type PrivateContractAPI struct {
	b Backend
}

func NewPrivateContractAPI(b Backend) *PrivateContractAPI {
	return &PrivateContractAPI{b}
}

//contract command
//install
func (s *PrivateContractAPI) Ccinstall(
	ccname, ccpath, ccversion, ccdescription, ccabi, cclanguage string) (hexutil.Bytes, error) {
	log.Info("CcInstall:", "ccname", ccname, "ccpath", ccpath, "ccversion", ccversion)
	templateId, err := s.b.ContractInstall(ccname, ccpath, ccversion, ccdescription, ccabi, cclanguage)
	return hexutil.Bytes(templateId), err
}

func (s *PrivateContractAPI) Ccdeploy(templateId string, param []string) (*ContractDeployRsp, error) {
	tempId, err := hex.DecodeString(templateId)
	if err != nil {
		return nil, err
	}
	rd, err := crypto.GetRandomBytes(32)
	if err != nil {
		return nil, err
	}
	txid := util.RlpHash(rd)

	log.Info("Ccdeploy:", "templateId", templateId, "txid", txid.String())
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Info("Ccdeploy", "param index:", i, "arg", arg)
	}
	//参数前面加入msg0和msg1,这里为空
	fullArgs := [][]byte{defaultMsg0, defaultMsg1}
	fullArgs = append(fullArgs, args...)
	_, err = s.b.ContractDeploy(tempId, txid.String(), fullArgs, time.Duration(30)*time.Second)
	if err != nil {
		log.Debug("Ccdeploy", "ContractDeploy err", err)
	}
	contractAddr := crypto.RequestIdToContractAddress(txid)
	log.Debug("-----Ccdeploy:", "txid", txid, "contractAddr", contractAddr.String())
	rsp := &ContractDeployRsp{
		ReqId:      "",
		ContractId: contractAddr.String(),
	}
	return rsp, nil
}

func (s *PrivateContractAPI) Ccinvoke(contractAddr string, param []string) (string, error) {
	contractId, _ := common.StringToAddress(contractAddr)
	//contractId, _ := hex.DecodeString(contractAddr)
	rd, _ := crypto.GetRandomBytes(32)
	txid := util.RlpHash(rd)
	log.Info("Ccinvoke", "contractId", contractId, "txid", txid.String())

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Info("Ccinvoke", "param index:", i, "arg", arg)
	}
	//参数前面加入msg0和msg1,这里为空
	fullArgs := [][]byte{defaultMsg0, defaultMsg1}
	fullArgs = append(fullArgs, args...)
	rsp, err := s.b.ContractInvoke(contractId.Bytes(), txid.String(), fullArgs, time.Duration(30)*time.Second)
	log.Info("Ccinvoke", "rsp", rsp)
	return string(rsp), err
}

func (s *PublicContractAPI) Ccquery(id string, param []string, timeout *Int) (string, error) {
	var idByte []byte

	log.Debugf("Ccquery, id len:%d, id[%s]", len(id), id)
	//if len(id) > 35 {
	//	idByte, _ = hex.DecodeString(id)
	//}else{
	//	idByte = []byte(id)
	//}
	idByte = []byte(id)
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Debug("Ccquery", "param index:", i, "arg", arg)
	}
	//参数前面加入msg0和msg1,这里为空
	fullArgs := [][]byte{defaultMsg0, defaultMsg1}
	fullArgs = append(fullArgs, args...)
	duration := time.Duration(0)
	if timeout != nil {
		duration = time.Duration(timeout.Uint32()) * time.Second
	}
	rsp, err := s.b.ContractQuery(idByte, fullArgs, duration)
	if err != nil {
		return "", err
	}
	return string(rsp), nil
}

func (s *PrivateContractAPI) Ccstop(contractAddr string) error {
	contractId, _ := common.StringToAddress(contractAddr)
	//contractId, _ := hex.DecodeString(contractAddr)
	txid := "123"
	log.Info("Ccstop:", "contractId", contractId, "txid", txid)
	err := s.b.ContractStop(contractId.Bytes(), txid, false)
	return err
}

//将Install包装成对系统合约的ccinvoke
func (s *PrivateContractAPI) Ccinstalltx(from, to string, amount, fee decimal.Decimal,
	tplName, path, version, ccdescription, ccabi, cclanguage string, addrs []string, password *string, timeout *Int) (*ContractInstallRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)

	log.Info("Ccinstalltx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   tplName[%s], path[%s],version[%s]", tplName, path, version)
	log.Infof("   description[%s], abi[%s],language[%s]", ccdescription, ccabi, cclanguage)
	log.Infof("   addrs len[%d]", len(addrs))
	if strings.ToLower(cclanguage) == GO {
		cclanguage = GOLANG
	}
	language := strings.ToUpper(cclanguage)
	if _, ok := peer.PtnChaincodeSpec_Type_value[language]; !ok {
		return nil, errors.New(cclanguage + " language is not supported")
	}

	for i, s := range addrs {
		_, err := common.StringToAddress(s)
		if err != nil {
			return nil, err
		}
		//addrs = append(addrs, a)
		log.Infof("    index[%d],addr[%s]", i, s)
	}
	//参数检查
	if fromAddr == (common.Address{}) || toAddr == (common.Address{}) || tplName == "" || path == "" || version == "" {
		log.Error("Ccinstalltx, param is error")
		return  nil, errors.New("Ccinstalltx, request param is error")
	}
	if len(tplName) > jury.MaxLengthTplName || len(path) > jury.MaxLengthTplPath || len(version) > jury.MaxLengthTplVersion ||
		len(ccdescription) > jury.MaxLengthDescription || len(ccabi) > jury.MaxLengthAbi || len(language) > jury.MaxLengthLanguage ||
		len(addrs) > jury.MaxNumberTplEleAddrHash {
		log.Error("Ccinstalltx", "request param len overflow，len(tplName)",
			len(tplName), "len(path)", len(path), "len(version)", len(version), "len(description)", len(ccdescription),
			"len(abi)", len(ccabi), "len(language)", len(language), "len(addrs)", len(addrs))
		return nil, errors.New("Ccinstalltx, request param len overflow")
	}

	usrcc := &ucc.UserChaincode{
		Name:     tplName,
		Path:     path,
		Version:  version,
		Language: language,
		Enabled:  true,
	}
	//将合约代码文件打包成 tar 文件
	byteCode, err := ucc.GetUserCCPayload(usrcc)
	if err != nil {
		log.Error("Ccinstalltx, getUserCCPayload err:", "error", err)
		return nil, err
	}
	juryAddr, _ := json.Marshal(addrs)
	contractAddr := syscontract.InstallContractAddress.String()
	result, err := s.Ccinvoketx(from, to, amount, fee, contractAddr, []string{
		"installByteCode",
		tplName,
		ccdescription,
		path,
		base64.StdEncoding.EncodeToString(byteCode),
		version,
		ccabi,
		language,
		string(juryAddr),
	}, password, timeout)
	if err != nil {
		return nil, err
	}
	tplId := getTemplateId(tplName, path, version)
	sTplId := hex.EncodeToString(tplId)

	rsp := &ContractInstallRsp{
		ReqId: result.ReqId,
		TplId: sTplId,
	}
	return rsp, err
}

func (s *PrivateContractAPI) Ccdeploytx(from, to string, amount, fee decimal.Decimal,
	tplId string, param []string, extData string) (*ContractDeployRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	templateId, _ := hex.DecodeString(tplId)
	extendData, _ := hex.DecodeString(extData)

	log.Info("Ccdeploytx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, ptnjson.Ptn2Dao(fee))
	log.Infof("   templateId[%s], extData[%s]", tplId, extData)

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	s.b.Lock()
	defer s.b.Unlock()

	//1.参数检查
	if fromAddr == (common.Address{}) || toAddr == (common.Address{}) || templateId == nil {
		log.Error("Ccdeploytx, param is error")
		return nil, errors.New("Ccdeploytx request param is error")
	}
	if len(templateId) > jury.MaxLengthTplId || len(args) > jury.MaxNumberArgs || len(extData) > jury.MaxLengthExtData {
		log.Error("Ccdeploytx", "request param len overflow, len(templateId)",
			len(templateId), "len(args)", len(args), "len(extData)", len(extData))
		return nil, errors.New("Ccdeploytx request param len overflow")
	}
	for _, arg := range args {
		if len(arg) > jury.MaxLengthArgs {
			log.Error("Ccdeploytx", "request param len overflow,len(arg)", len(arg))
			return nil, errors.New("Ccdeploytx request param len overflow")
		}
	}
	//2.费用检查
	ctx := &buildContractContext{
		tokenId:    dagconfig.DagConfig.GasToken,
		fromAddr:   fromAddr,
		toAddr:     toAddr,
		amount:     amount,
		gasFee:     fee,
		args:       args,
		password:   "",
		exeTimeout: Int{0},
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_DEPLOY_REQUEST,
		Payload: &modules.ContractDeployRequestPayload{
			TemplateId: templateId,
			Args:       args,
			ExtData:    extendData,
			Timeout:    0,
		},
	}

	daoFee, err := s.contractFeeCheck(s.b.EnableGasFee(), ctx, msgReq)
	if err != nil {
		log.Errorf("Ccdeploytx, contractFeeCheck err:%s", err.Error())
		return nil, err
	}
	ctx.gasFee = daoFee

	//3.构建合约请求交易
	tx, err := s.buildContractReqTx(ctx, msgReq)

	//4.只广播交易事件
	reqId := tx.RequestHash()
	go s.b.ContractEventBroadcast(jury.ContractEvent{Ele: nil, CType: jury.CONTRACT_EVENT_ELE, Tx: tx}, true)

	//5.执行结果返回
	contractAddr := crypto.RequestIdToContractAddress(reqId)
	sReqId := hex.EncodeToString(reqId[:])
	log.Debug("-----Ccdeploytx:", "reqId", sReqId, "depId", contractAddr.String())
	rsp := &ContractDeployRsp{
		ReqId:      sReqId,
		ContractId: contractAddr.String(),
	}
	return rsp, err
}

func (s *PrivateContractAPI) Ccinvoketx(from, to string, amount, fee decimal.Decimal,
	contractAddress string, param []string, password *string, timeout *Int) (*ContractInvokeRsp, error) {
	return s.CcinvokeToken(from, to, dagconfig.DagConfig.GasToken, amount, fee, contractAddress, param, password, timeout)
}

func (s *PrivateContractAPI) CcinvokeToken(from, to, token string, amountToken, fee decimal.Decimal,
	contractAddress string, param []string, pwd *string, timeout *Int) (*ContractInvokeRsp, error) {
	password := ""
	if pwd != nil {
		password = *pwd
	}
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	contractAddr, _ := common.StringToAddress(contractAddress)

	log.Info("Ccinvoketx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   token[%s], amountToken[%d], fee[%d]", token, amountToken, fee)
	log.Infof("   contractAddr[%s], is systemContract[%v]", contractAddr.String(), contractAddr.IsSystemContractAddress())
	log.Infof("   param len[%d]", len(param))
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Infof("      index[%d], value[%s]\n", i, arg)
	}

	s.b.Lock()
	defer s.b.Unlock()
	//1.参数检查
	if fromAddr == (common.Address{}) || toAddr == (common.Address{}) || contractAddr == (common.Address{}) || args == nil {
		log.Error("Ccinvoketx, param is error")
		return nil, errors.New("Ccinvoketx request param is error")
	}
	if len(args) > jury.MaxNumberArgs {
		log.Error("Ccinvoketx", "len(args)", len(args))
		return nil, errors.New("Ccinvoketx request param len overflow")
	}
	for _, arg := range args {
		if len(arg) > jury.MaxLengthArgs {
			log.Error("Ccinvoketx", "request param len overflow,len(arg)", len(arg))
			return nil, errors.New("Ccinvoketx request param args len overflow")
		}
	}
	//2.费用检查
	ctx := &buildContractContext{
		tokenId:    dagconfig.DagConfig.GasToken,
		fromAddr:   fromAddr,
		toAddr:     toAddr,
		amount:     amountToken,
		gasFee:     fee,
		args:       args,
		password:   password,
		exeTimeout: *timeout,
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: contractAddr.Bytes(),
			Args:       args,
			Timeout:    timeout.Uint32(),
		},
	}

	daoFee, err := s.contractFeeCheck(s.b.EnableGasFee(), ctx, msgReq)
	if err != nil {
		log.Errorf("Ccdeploytx, contractFeeCheck err:%s", err.Error())
		return nil, err
	}
	ctx.gasFee = daoFee

	//3.构建请求交易
	//如没有GasFee，而且to address不是合约地址，则不构建Payment，直接InvokeRequest+Signature
	//if s.b.EnableGasFee() || toAddr == contractAddr || fromAddr != toAddr {
	tx, err := s.buildContractReqTx(ctx, msgReq)
	if err != nil {
		log.Errorf("Ccdeploytx, buildContractReqTx err:%s", err.Error())
		return nil, err
	}

	//4. 广播交易
	reqId, err := submitTransaction(s.b, tx)
	if err != nil {
		log.Errorf("CcinvokeToken, submitTransaction err:%s", err.Error())
		return nil, err
	}

	log.Infof("   reqId[%s]", hex.EncodeToString(reqId[:]))
	rsp1 := &ContractInvokeRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: contractAddress,
	}
	return rsp1, err
}

func (s *PrivateContractAPI) Ccstoptx(from, to string, amount, fee decimal.Decimal, contractId string) (*ContractStopRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	contractAddr, _ := common.StringToAddress(contractId)

	log.Info("Ccstoptx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%s]", daoAmount, fee.String())
	log.Infof("   contractId[%s]", contractAddr.String())

	s.b.Lock()
	defer s.b.Unlock()
	//1.参数检查
	if fromAddr == (common.Address{}) || toAddr == (common.Address{}) || contractAddr == (common.Address{}) {
		log.Error("Ccstoptx, param is error")
		return nil, errors.New("Ccstoptx request param is error")
	}
	//2.费用检查
	//daoFee := fee
	//var err error
	ctx := &buildContractContext{
		tokenId:    dagconfig.DagConfig.GasToken,
		password:   "",
		fromAddr:   fromAddr,
		toAddr:     toAddr,
		amount:     amount,
		gasFee:     fee,
		exeTimeout: Int{0},
	}
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		log.Errorf("Ccstoptx, GetRandomNonce err:%s", err.Error())
		return nil, err
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_STOP_REQUEST,
		Payload: &modules.ContractStopRequestPayload{
			ContractId:  contractAddr.Bytes(),
			Txid:        hex.EncodeToString(randNum),
			DeleteImage: false,
		},
	}
	daoFee, err := s.contractFeeCheck(s.b.EnableGasFee(), ctx, msgReq)
	if err != nil {
		log.Errorf("Ccstoptx, contractFeeCheck err:%s", err.Error())
		return nil, err
	}

	//3.构建请求交易
	ctx.gasFee = daoFee
	if err != nil {
		return nil, errors.New("Ccstoptx, GetRandomNonce error")
	}
	tx, err := s.buildContractReqTx(ctx, msgReq)
	if err != nil {
		log.Errorf("Ccstoptx, buildContractReqTx err:%s", err.Error())
		return nil, err
	}

	//4.广播
	reqId, err := submitTransaction(s.b, tx)
	if err != nil {
		log.Errorf("Ccstoptx, submitTransaction err:%s", err.Error())
		return nil, err
	}
	go s.b.ContractEventBroadcast(jury.ContractEvent{Ele: nil, CType: jury.CONTRACT_EVENT_EXEC, Tx: tx}, true)
	//go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_EXEC, Ele: p.mtx[reqId].eleNode, Tx: tx}, true)
	log.Infof("   reqId[%s]", hex.EncodeToString(reqId[:]))
	rsp := &ContractStopRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: contractId,
	}
	return rsp, err
}

func (s *PrivateContractAPI) Ccinstalltxfee(from, to string, amount, fee decimal.Decimal,
	tplName, path, version, ccdescription, ccabi, cclanguage string, addr []string) (*ContractFeeRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)

	log.Info("CcInstallTxFee info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   tplName[%s], path[%s],version[%s]", tplName, path, version)
	log.Infof("   description[%s], abi[%s],language[%s]", ccdescription, ccabi, cclanguage)
	log.Infof("   addrs len[%d]", len(addr))
	if strings.ToLower(cclanguage) == GO {
		cclanguage = GOLANG
	}
	language := strings.ToUpper(cclanguage)
	if _, ok := peer.PtnChaincodeSpec_Type_value[language]; !ok {
		return nil, errors.New(cclanguage + " language is not supported")
	}
	addrs := make([]common.Address, 0)
	for i, s := range addr {
		a, _ := common.StringToAddress(s)
		addrs = append(addrs, a)
		log.Infof("    index[%d],addr[%s]", i, s)
	}
	afee, sz, tm, err := s.b.ContractInstallReqTxFee(fromAddr, toAddr, daoAmount, daoFee,
		tplName, path, version, ccdescription, ccabi, language, addrs)
	if err != nil {
		return nil, err
	}
	rsp := &ContractFeeRsp{
		TxSize:         sz,
		TimeOut:        tm,
		ApproximateFee: afee,
	}
	log.Infof("   fee[%f]", afee)
	return rsp, nil
}

func (s *PrivateContractAPI) Ccdeploytxfee(from, to string, amount, fee decimal.Decimal,
	tplId string, param []string, extData string) (*ContractFeeRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	templateId, _ := hex.DecodeString(tplId)
	extendData, _ := hex.DecodeString(extData)

	log.Info("CcDeployTxFee info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   templateId[%s], extData[%s]", tplId, extData)

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	fullArgs := [][]byte{defaultMsg0}
	fullArgs = append(fullArgs, args...)
	afee, sz, tm, err := s.b.ContractDeployReqTxFee(fromAddr, toAddr, daoAmount, daoFee, templateId, fullArgs, extendData, 0)
	if err != nil {
		return nil, err
	}
	rsp := &ContractFeeRsp{
		TxSize:         sz,
		TimeOut:        tm,
		ApproximateFee: afee,
	}
	log.Infof("   fee[%f]", afee)
	return rsp, nil
}

func (s *PrivateContractAPI) Ccinvoketxfee(from, to string, amount, fee decimal.Decimal,
	deployId string, param []string, certID string, timeout string) (*ContractFeeRsp, error) {
	contractAddr, _ := common.StringToAddress(deployId)
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	timeout64, _ := strconv.ParseUint(timeout, 10, 64)

	log.Info("CcInvokeTxFee info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   contractId[%s], certID[%s], timeout[%s]", contractAddr.String(), certID, timeout)

	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return nil, fmt.Errorf("certid is invalid")
		}
	}

	log.Infof("   param len[%d]", len(param))
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Infof("      index[%d], value[%s]\n", i, arg)
	}
	afee, sz, tm, err := s.b.ContractInvokeReqTxFee(fromAddr, toAddr, daoAmount, daoFee, intCertID, contractAddr, args, uint32(timeout64))
	if err != nil {
		return nil, err
	}
	rsp := &ContractFeeRsp{
		TxSize:         sz,
		TimeOut:        tm,
		ApproximateFee: afee,
	}
	log.Infof("   fee[%f]", afee)
	return rsp, nil
}

func (s *PrivateContractAPI) Ccstoptxfee(from, to string, amount, fee decimal.Decimal,
	contractId string) (*ContractFeeRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	contractAddr, _ := common.StringToAddress(contractId)

	log.Info("CcStopTx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   contractId[%s]", contractAddr.String())

	afee, sz, tm, err := s.b.ContractStopReqTxFee(fromAddr, toAddr, daoAmount, daoFee, contractAddr, false)
	if err != nil {
		return nil, err
	}
	rsp := &ContractFeeRsp{
		TxSize:         sz,
		TimeOut:        tm,
		ApproximateFee: afee,
	}
	log.Infof("   fee[%f]", afee)
	return rsp, nil
}

//  TODO
func (s *PublicContractAPI) ListAllContractTemplates() ([]*ptnjson.ContractTemplateJson, error) {
	return s.b.GetAllContractTpl()
}

//  TODO
func (s *PublicContractAPI) ListAllContracts() ([]*ptnjson.ContractJson, error) {
	return s.b.GetAllContracts()
}

//  通过合约模板id获取模板信息
func (s *PublicContractAPI) GetContractTemplateInfoById(contractTplId string) (*modules.ContractTemplate, error) {
	templateId, err := hex.DecodeString(contractTplId)
	if err != nil {
		return nil, err
	}
	return s.b.GetContractTpl(templateId)
}

//  查看某个模板id对应着多个合约实例的合约信息
func (s *PublicContractAPI) GetAllContractsUsedTemplateId(tplId string) ([]*ptnjson.ContractJson, error) {
	id, err := hex.DecodeString(tplId)
	if err != nil {
		return nil, err
	}
	return s.b.GetContractsByTpl(id)
}

//  通过合约Id，获取合约的详细信息
func (s *PublicContractAPI) GetContractInfoById(contractId string) (*ptnjson.ContractJson, error) {
	id, _ := hex.DecodeString(contractId)
	addr := common.NewAddress(id, common.ContractHash)
	contract, err := s.b.GetContract(addr)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

//  通过合约地址，获取合约的详细信息
func (s *PublicContractAPI) GetContractInfoByAddr(contractAddr string) (*ptnjson.ContractJson, error) {
	addr, _ := common.StringToAddress(contractAddr)
	contract, err := s.b.GetContract(addr)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

func (s *PrivateContractAPI) DepositContractInvoke(from, to string, amount, fee decimal.Decimal,
	param []string) (string, error) {
	log.Debug("---enter DepositContractInvoke---")
	// append by albert·gou
	if param[0] == modules.ApplyMediator {
		args := modules.NewMediatorCreateArgs()
		err := json.Unmarshal([]byte(param[1]), &args)
		if err != nil {
			return "", fmt.Errorf("param error(%v), please use mediator.apply()", err.Error())
		} else {
			// 参数验证
			_, _, err := args.Validate()
			if err != nil {
				return "", fmt.Errorf("error(%v), please use mediator.apply()", err.Error())
			}

			if args.MediatorInfoBase == nil || args.MediatorApplyInfo == nil {
				return "", fmt.Errorf("invalid args, is null")
			}

			if from != args.AddStr {
				return "", fmt.Errorf("the calling account(%v) is not applying account(%v), "+
					"please use mediator.apply()", from, args.AddStr)
			}

			// 参数序列化
			argsB, err := json.Marshal(args)
			if err != nil {
				return "", fmt.Errorf("error(%v), please use mediator.apply()", err.Error())
			}

			param[1] = string(argsB)
		}
	}

	rsp, err := s.Ccinvoketx(from, to, amount, fee, syscontract.DepositContractAddress.String(),
		param, nil, nil)
	if err != nil {
		return "", err
	}
	return rsp.ReqId, err
}

func (s *PublicContractAPI) DepositContractQuery(param []string) (string, error) {
	log.Debug("---enter DepositContractQuery---")
	return s.Ccquery(syscontract.DepositContractAddress.String(), param, nil)
}

func (s *PublicContractAPI) SysConfigContractQuery(param []string) (string, error) {
	log.Debugf("---enter SysConfigContractQuery---")
	return s.Ccquery(syscontract.SysConfigContractAddress.String(), param, nil)
}

func (s *PrivateContractAPI) SysConfigContractInvoke(from, to string, amount, fee decimal.Decimal,
	param []string) (string, error) {
	log.Debugf("---enter SysConfigContractInvoke---")
	if len(param) == 0 {
		err := "args is nil"
		log.Debugf(err)
		return "", fmt.Errorf(err)
	}

	// 检查参数
	if param[0] == sysconfigcc.UpdateSysParamWithoutVote {
		if len(param) != 3 {
			err := "args len not equal 3"
			log.Debugf(err)
			return "", fmt.Errorf(err)
		}

		field, value := param[1], param[2]
		err := core.CheckSysConfigArgType(field, value)
		if err != nil {
			log.Debugf(err.Error())
			return "", err
		}

		dag := s.b.Dag()
		gp := s.b.Dag().GetGlobalProp()
		err = core.CheckChainParameterValue(field, value, &gp.ImmutableParameters, &gp.ChainParameters,
			dag.GetMediatorCount)
		if err != nil {
			log.Debugf(err.Error())
			return "", err
		}
	} else if param[0] == sysconfigcc.CreateVotesTokens {
		if len(param) != 6 {
			err := "args len not equal 6"
			log.Debugf(err)
			return "", fmt.Errorf(err)
		}

		var voteTopics []sysconfigcc.SysVoteTopic
		err := json.Unmarshal([]byte(param[5]), &voteTopics)
		if err != nil {
			log.Debugf(err.Error())
			return "", err
		}

		for _, oneTopic := range voteTopics {
			for _, oneOption := range oneTopic.SelectOptions {
				err := core.CheckSysConfigArgType(oneTopic.TopicTitle, oneOption)
				if err != nil {
					log.Debugf(err.Error())
					return "", err
				}

				dag := s.b.Dag()
				gp := s.b.Dag().GetGlobalProp()
				err = core.CheckChainParameterValue(oneTopic.TopicTitle, oneOption, &gp.ImmutableParameters,
					&gp.ChainParameters, dag.GetMediatorCount)
				if err != nil {
					log.Debugf(err.Error())
					return "", err
				}
			}
		}
	}

	rsp, err := s.Ccinvoketx(from, to, amount, fee, syscontract.SysConfigContractAddress.String(),
		param, nil, nil)
	if err != nil {
		return "", err
	}
	return rsp.ReqId, err
}

func (s *PublicContractAPI) GetContractState(contractAddr, prefix string) (string, error) {
	addr, err := common.StringToAddress(contractAddr)
	if err != nil {
		return "", err
	}
	mvalue, err := s.b.GetContractStateJsonByPrefix(addr.Bytes(), prefix)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(mvalue)
	return string(data), nil
}
func (s *PublicContractAPI) GetContractFeeLevel() (*ContractFeeLevelRsp, error) {
	cp := s.b.Dag().GetChainParameters()
	feeLevel := &ContractFeeLevelRsp{
		ContractTxTimeoutUnitFee:  cp.ContractTxTimeoutUnitFee,
		ContractTxSizeUnitFee:     cp.ContractTxSizeUnitFee,
		ContractTxInstallFeeLevel: cp.ContractTxInstallFeeLevel,
		ContractTxDeployFeeLevel:  cp.ContractTxDeployFeeLevel,
		ContractTxInvokeFeeLevel:  cp.ContractTxInvokeFeeLevel,
		ContractTxStopFeeLevel:    cp.ContractTxStopFeeLevel,
	}
	return feeLevel, nil
}

//获取所担任的用户合约相关信息
func (s *PublicContractAPI) GetContractsWithJuryAddress(addr string) ([]*ptnjson.ContractJson, error) {
	a, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	log.Debugf("jury address %s", a.String())
	cs := s.b.GetContractsWithJuryAddr(util.RlpHash(a))
	cj := make([]*ptnjson.ContractJson, 0, len(cs))
	for _, c := range cs {
		cj = append(cj, ptnjson.ConvertContract2Json(c))
	}
	return cj, nil
}
