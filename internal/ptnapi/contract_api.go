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
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"github.com/shopspring/decimal"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/contracts/syscontract/sysconfigcc"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
)

var (
	defaultMsg0 = []byte("query has no msg0")
	defaultMsg1 = []byte("query has no msg1")
)

type PublicContractAPI struct {
	b Backend
}

func NewPublicContractAPI(b Backend) *PublicContractAPI {
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
func (s *PrivateContractAPI) Ccinstall(ctx context.Context,
	ccname, ccpath, ccversion, ccdescription, ccabi, cclanguage string) (hexutil.Bytes, error) {
	log.Info("CcInstall:", "ccname", ccname, "ccpath", ccpath, "ccversion", ccversion)
	templateId, err := s.b.ContractInstall(ccname, ccpath, ccversion, ccdescription, ccabi, cclanguage)
	return hexutil.Bytes(templateId), err
}

func (s *PrivateContractAPI) Ccdeploy(ctx context.Context, templateId string, param []string) (*ContractDeployRsp, error) {
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

func (s *PrivateContractAPI) Ccinvoke(ctx context.Context, contractAddr string, param []string) (string, error) {
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

func (s *PublicContractAPI) Ccquery(ctx context.Context, contractAddr string, param []string, timeout uint32) (string, error) {
	contractId, _ := common.StringToAddress(contractAddr)
	//contractId, _ := hex.DecodeString(contractAddr)
	rd, _ := crypto.GetRandomBytes(32)
	txid := util.RlpHash(rd)
	log.Info("Ccquery", "contractId", contractId, "txid", txid.String())
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Info("Ccquery", "param index:", i, "arg", arg)
	}

	//参数前面加入msg0和msg1,这里为空
	fullArgs := [][]byte{defaultMsg0, defaultMsg1}
	fullArgs = append(fullArgs, args...)
	rsp, err := s.b.ContractQuery(contractId.Bytes(), txid.String(), fullArgs, time.Duration(timeout)*time.Second)
	if err != nil {
		return "", err
	}
	return string(rsp), nil
}

func (s *PrivateContractAPI) Ccstop(ctx context.Context, contractAddr string) error {
	contractId, _ := common.StringToAddress(contractAddr)
	//contractId, _ := hex.DecodeString(contractAddr)
	txid := "123"
	log.Info("Ccstop:", "contractId", contractId, "txid", txid)
	err := s.b.ContractStop(contractId.Bytes(), txid, false)
	return err
}

//contract tx
func (s *PrivateContractAPI) Ccinstalltx(ctx context.Context, from, to string, amount, fee decimal.Decimal,
	tplName, path, version, ccdescription, ccabi, cclanguage string, addr []string) (*ContractInstallRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)

	log.Info("Ccinstalltx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   tplName[%s], path[%s],version[%s]", tplName, path, version)
	log.Infof("   description[%s], abi[%s],language[%s]", ccdescription, ccabi, cclanguage)
	log.Infof("   addrs len[%d]", len(addr))
	if strings.ToLower(cclanguage) == "go" {
		cclanguage = "golang"
	}
	language := strings.ToUpper(cclanguage)
	if _, ok := peer.ChaincodeSpec_Type_value[language]; !ok {
		return nil, errors.New(cclanguage + " language is not supported")
	}

	addrs := make([]common.Address, 0)
	for i, s := range addr {
		a, _ := common.StringToAddress(s)
		addrs = append(addrs, a)
		log.Infof("    index[%d],addr[%s]", i, s)
	}
	reqId, tplId, err := s.b.ContractInstallReqTx(fromAddr, toAddr, daoAmount, daoFee, tplName, path, version,
		ccdescription, ccabi, language, addrs)
	sReqId := hex.EncodeToString(reqId[:])
	sTplId := hex.EncodeToString(tplId)
	log.Info("Ccinstalltx:", "reqId", sReqId, "tplId", sTplId)

	rsp := &ContractInstallRsp{
		ReqId: sReqId,
		TplId: sTplId,
	}

	return rsp, err
}
func (s *PrivateContractAPI) Ccdeploytx(ctx context.Context, from, to string, amount, fee decimal.Decimal,
	tplId string, param []string, extData string) (*ContractDeployRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	templateId, _ := hex.DecodeString(tplId)
	extendData, _ := hex.DecodeString(extData)

	log.Info("Ccdeploytx info:")
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
	reqId, _, err := s.b.ContractDeployReqTx(fromAddr, toAddr, daoAmount, daoFee, templateId, fullArgs, extendData, 0)
	contractAddr := crypto.RequestIdToContractAddress(reqId)
	sReqId := hex.EncodeToString(reqId[:])
	log.Debug("-----Ccdeploytx:", "reqId", sReqId, "depId", contractAddr.String())
	rsp := &ContractDeployRsp{
		ReqId:      sReqId,
		ContractId: contractAddr.String(),
	}
	return rsp, err
}

func (s *PrivateContractAPI) Ccinvoketx(ctx context.Context, from, to string, amount, fee decimal.Decimal,
	deployId string, param []string, certID string, timeout string) (*ContractDeployRsp, error) {
	contractAddr, _ := common.StringToAddress(deployId)
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	timeout64, _ := strconv.ParseUint(timeout, 10, 64)

	log.Info("Ccinvoketx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   contractId[%s], certID[%s], timeout[%s]", contractAddr.String(), certID, timeout)

	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return &ContractDeployRsp{}, fmt.Errorf("certid is invalid")
		}
	}

	log.Infof("   param len[%d]", len(param))
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Infof("      index[%d], value[%s]\n", i, arg)
	}
	reqId, err := s.b.ContractInvokeReqTx(fromAddr, toAddr, daoAmount, daoFee, intCertID, contractAddr, args, uint32(timeout64))
	//log.Debug("-----ContractInvokeTxReq:" + hex.EncodeToString(reqId[:]))
	log.Infof("   reqId[%s]", hex.EncodeToString(reqId[:]))
	rsp1 := &ContractDeployRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: deployId,
	}
	return rsp1, err
}

func (s *PrivateContractAPI) CcinvokeToken(ctx context.Context, from, to, toToken string, amount, fee decimal.Decimal,
	assetToken, amountToken, deployId string, param []string) (*ContractDeployRsp, error) {
	contractAddr, _ := common.StringToAddress(deployId)
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	toAddrToken, _ := common.StringToAddress(toToken)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	amountOfToken, _ := strconv.ParseUint(amountToken, 10, 64)

	log.Info("CcinvokeToken info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   toAddrToken[%s], amountVote[%d]", toAddrToken.String(), amountOfToken)
	log.Infof("   contractId[%s]", contractAddr.String())

	log.Infof("   param len[%d]", len(param))
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Infof("      index[%d], value[%s]\n", i, arg)
	}
	reqId, err := s.b.ContractInvokeReqTokenTx(fromAddr, toAddr, toAddrToken, daoAmount, daoFee,
		amountOfToken, assetToken, contractAddr, args, 0)
	log.Infof("   reqId[%s]", hex.EncodeToString(reqId[:]))
	rsp1 := &ContractDeployRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: deployId,
	}
	return rsp1, err
}

func (s *PrivateContractAPI) CcinvoketxPass(ctx context.Context, from, to string, amount, fee decimal.Decimal,
	deployId string, param []string, password string, duration *uint64, certID string) (string, error) {
	contractAddr, _ := common.StringToAddress(deployId)
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)

	log.Info("CcinvoketxPass info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   contractId[%s], certID[%s], password[%s]", contractAddr.String(), certID, password)

	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return "", fmt.Errorf("certid is invalid")
		}
	}
	log.Infof("   param len[%d]", len(param))
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Infof("      index[%d], value[%s]\n", i, arg)
	}

	//2.
	err := s.unlockKS(fromAddr, password, duration)
	if err != nil {
		return "", err
	}

	reqId, err := s.b.ContractInvokeReqTx(fromAddr, toAddr, daoAmount, daoFee, intCertID, contractAddr, args, 0)
	log.Infof("   reqId[%s]", hex.EncodeToString(reqId[:]))

	return hex.EncodeToString(reqId[:]), err
}

func (s *PrivateContractAPI) Ccstoptx(ctx context.Context, from, to string, amount, fee decimal.Decimal, contractId string) (string, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	daoAmount := ptnjson.Ptn2Dao(amount)
	daoFee := ptnjson.Ptn2Dao(fee)
	contractAddr, _ := common.StringToAddress(contractId)

	log.Info("Ccstoptx info:")
	log.Infof("   fromAddr[%s], toAddr[%s]", fromAddr.String(), toAddr.String())
	log.Infof("   daoAmount[%d], daoFee[%d]", daoAmount, daoFee)
	log.Infof("   contractId[%s]", contractAddr.String())

	reqId, err := s.b.ContractStopReqTx(fromAddr, toAddr, daoAmount, daoFee, contractAddr, false)
	log.Infof("   reqId[%s]", hex.EncodeToString(reqId[:]))
	return hex.EncodeToString(reqId[:]), err
}

func (s *PrivateContractAPI) Ccinstalltxfee(ctx context.Context, from, to string, amount, fee decimal.Decimal,
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
	if strings.ToLower(cclanguage) == "go" {
		cclanguage = "golang"
	}
	language := strings.ToUpper(cclanguage)
	if _, ok := peer.ChaincodeSpec_Type_value[language]; !ok {
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

func (s *PrivateContractAPI) Ccdeploytxfee(ctx context.Context, from, to string, amount, fee decimal.Decimal,
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

func (s *PrivateContractAPI) Ccinvoketxfee(ctx context.Context, from, to string, amount, fee decimal.Decimal,
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

func (s *PrivateContractAPI) Ccstoptxfee(ctx context.Context, from, to string, amount, fee decimal.Decimal,
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

func (s *PrivateContractAPI) unlockKS(addr common.Address, password string, duration *uint64) error {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	ks := s.b.GetKeyStore()
	err := ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		return fmt.Errorf("get addr by outpoint is err: %v", err.Error())
	}
	return nil
}

//  TODO
func (s *PublicContractAPI) ListAllContractTemplates(ctx context.Context) ([]*ptnjson.ContractTemplateJson, error) {
	return s.b.GetAllContractTpl()
}

//  TODO
func (s *PublicContractAPI) ListAllContracts(ctx context.Context) ([]*ptnjson.ContractJson, error) {
	return s.b.GetAllContracts()
}

//  通过合约模板id获取模板信息
func (s *PublicContractAPI) GetContractTemplateInfoById(ctx context.Context, contractTplId string) (*modules.ContractTemplate, error) {
	templateId, err := hex.DecodeString(contractTplId)
	if err != nil {
		return nil, err
	}
	return s.b.GetContractTpl(templateId)
}

//  查看某个模板id对应着多个合约实例的合约信息
func (s *PublicContractAPI) GetAllContractsUsedTemplateId(ctx context.Context, tplId string) ([]*ptnjson.ContractJson, error) {
	id, err := hex.DecodeString(tplId)
	if err != nil {
		return nil, err
	}
	return s.b.GetContractsByTpl(id)
}

//  通过合约Id，获取合约的详细信息
func (s *PublicContractAPI) GetContractInfoById(ctx context.Context, contractId string) (*ptnjson.ContractJson, error) {
	id, _ := hex.DecodeString(contractId)
	addr := common.NewAddress(id, common.ContractHash)
	contract, err := s.b.GetContract(addr)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

//  通过合约地址，获取合约的详细信息
func (s *PublicContractAPI) GetContractInfoByAddr(ctx context.Context, contractAddr string) (*ptnjson.ContractJson, error) {
	addr, _ := common.StringToAddress(contractAddr)
	contract, err := s.b.GetContract(addr)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

func (s *PrivateContractAPI) DepositContractInvoke(ctx context.Context, from, to string, amount, fee decimal.Decimal,
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
			_, err := args.Validate()
			if err != nil {
				return "", fmt.Errorf("error(%v), please use mediator.apply()", err.Error())
			}

			if args.MediatorInfoBase == nil || args.MediatorApplyInfo == nil {
				return "", fmt.Errorf("invalid args, is null")
			}

			if from != args.AddStr{
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

	rsp, err := s.Ccinvoketx(ctx, from, to, amount, fee, syscontract.DepositContractAddress.String(),
		param, "", "0")

	return rsp.ReqId, err
}

func (s *PublicContractAPI) DepositContractQuery(ctx context.Context, param []string) (string, error) {
	log.Debug("---enter DepositContractQuery---")
	return s.Ccquery(ctx, syscontract.DepositContractAddress.String(), param, 0)
}

func (s *PublicContractAPI) SysConfigContractQuery(ctx context.Context, param []string) (string, error) {
	log.Debugf("---enter SysConfigContractQuery---")
	return s.Ccquery(ctx, syscontract.SysConfigContractAddress.String(), param, 0)
}

func (s *PrivateContractAPI) SysConfigContractInvoke(ctx context.Context, from, to string, amount, fee decimal.Decimal,
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
				err = core.CheckChainParameterValue(oneTopic.TopicTitle, oneOption, 	&gp.ImmutableParameters,
					&gp.ChainParameters, dag.GetMediatorCount)
				if err != nil {
					log.Debugf(err.Error())
					return "", err
				}
			}
		}
	}

	rsp, err := s.Ccinvoketx(ctx, from, to, amount, fee, syscontract.SysConfigContractAddress.String(),
		param, "", "0")

	return rsp.ReqId, err
}

//  TODO
func (s *PublicContractAPI) GetContractState(contractid []byte, key string) ([]byte, *modules.StateVersion, error) {
	return s.b.GetContractState(contractid, key)
}

func (s *PublicContractAPI) GetContractFeeLevel(ctx context.Context) (*ContractFeeLevelRsp, error) {
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
