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
	"time"

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
	"strings"
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

//contract command
//install
func (s *PublicContractAPI) Ccinstall(ctx context.Context, ccname, ccpath, ccversion, ccdescription, ccabi, cclanguage string) (hexutil.Bytes, error) {
	log.Info("CcInstall:", "ccname", ccname, "ccpath", ccpath, "ccversion", ccversion)
	templateId, err := s.b.ContractInstall(ccname, ccpath, ccversion, ccdescription, ccabi, cclanguage)
	return hexutil.Bytes(templateId), err
}

func (s *PublicContractAPI) Ccdeploy(ctx context.Context, templateId string, param []string) (*ContractDeployRsp, error) {
	tempId, _ := hex.DecodeString(templateId)
	rd, err := crypto.GetRandomBytes(32)
	txid := util.RlpHash(rd)

	log.Info("Ccdeploy:", "templateId", templateId, "txid", txid.String())
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		log.Info("Ccdeploy", "param index:", i, "arg", arg)
	}
	_, err = s.b.ContractDeploy(tempId, txid.String(), args, time.Duration(30)*time.Second)
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

func (s *PublicContractAPI) Ccinvoke(ctx context.Context, contractAddr string, param []string) (string, error) {
	contractId, _ := common.StringToAddress(contractAddr)
	//contractId, _ := hex.DecodeString(contractAddr)
	rd, err := crypto.GetRandomBytes(32)
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
	rd, err := crypto.GetRandomBytes(32)
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

func (s *PublicContractAPI) Ccstop(ctx context.Context, contractAddr string) error {
	contractId, _ := common.StringToAddress(contractAddr)
	//contractId, _ := hex.DecodeString(contractAddr)
	txid := "123"
	log.Info("Ccstop:", "contractId", contractId, "txid", txid)
	err := s.b.ContractStop(contractId.Bytes(), txid, false)
	return err
}

//contract tx
func (s *PublicContractAPI) Ccinstalltx(ctx context.Context, from, to, daoAmount, daoFee, tplName, path, version, ccdescription, ccabi, cclanguage string, addr []string) (*ContractInstallRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)

	//templateName, _ := hex.DecodeString(tplName)

	log.Debug("-----Ccinstalltx:", "fromAddr", fromAddr.String())
	log.Debug("-----Ccinstalltx:", "toAddr", toAddr.String())
	log.Debug("-----Ccinstalltx:", "amount", amount)
	log.Debug("-----Ccinstalltx:", "fee", fee)
	log.Debug("-----Ccinstalltx:", "tplName", tplName)
	log.Debug("-----Ccinstalltx:", "path", path)
	log.Debug("-----Ccinstalltx:", "version", version)
	log.Debug("-----Ccinstalltx:", "description", ccdescription)
	log.Debug("-----Ccinstalltx:", "abi", ccabi)
	log.Debug("-----Ccinstalltx:", "language", cclanguage)

	if strings.ToLower(cclanguage) == "go" {
		cclanguage = "golang"
	}
	language := strings.ToUpper(cclanguage)
	if _, ok := peer.ChaincodeSpec_Type_value[language]; !ok {
		return nil, errors.New(cclanguage + " language is not supported")
	}
	/*
		"P1QFTh1Xq2JpfTbu9bfaMfWh2sR1nHrMV8z", "P1NHVBFRkooh8HD9SvtvU3bpbeVmuGKPPuF",
		"P1PpgjUC7Nkxgi5KdKCGx2tMu6F5wfPGrVX", "P1MBXJypFCsQpafDGi9ivEooR8QiYmxq4qw"
	*/
	//addr := []string{"P1QFTh1Xq2JpfTbu9bfaMfWh2sR1nHrMV8z", "P1NHVBFRkooh8HD9SvtvU3bpbeVmuGKPPuF",
	//	"P1PpgjUC7Nkxgi5KdKCGx2tMu6F5wfPGrVX", "P1MBXJypFCsQpafDGi9ivEooR8QiYmxq4qw"}
	//var addr []string

	addrs := make([]common.Address, 0)
	for _, s := range addr {
		a, _ := common.StringToAddress(s)
		addrs = append(addrs, a)
	}
	log.Debug("-----Ccinstalltx:", "addrHash", addrs, "len", len(addrs))

	reqId, tplId, err := s.b.ContractInstallReqTx(fromAddr, toAddr, amount, fee, tplName, path, version, ccdescription, ccabi, language, addrs)
	sReqId := hex.EncodeToString(reqId[:])
	sTplId := hex.EncodeToString(tplId)
	log.Debug("-----Ccinstalltx:", "reqId", sReqId, "tplId", sTplId)

	rsp := &ContractInstallRsp{
		ReqId: sReqId,
		TplId: sTplId,
	}

	return rsp, err
}
func (s *PublicContractAPI) Ccdeploytx(ctx context.Context, from, to, daoAmount, daoFee, tplId string, param []string) (*ContractDeployRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)
	templateId, _ := hex.DecodeString(tplId)

	log.Debug("-----Ccdeploytx:", "fromAddr", fromAddr.String())
	log.Debug("-----Ccdeploytx:", "toAddr", toAddr.String())
	log.Debug("-----Ccdeploytx:", "amount", amount)
	log.Debug("-----Ccdeploytx:", "fee", fee)
	log.Debug("-----Ccdeploytx:", "tplId", templateId)

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	fullArgs := [][]byte{defaultMsg0}
	fullArgs = append(fullArgs, args...)
	reqId, _, err := s.b.ContractDeployReqTx(fromAddr, toAddr, amount, fee, templateId, fullArgs, 0)
	contractAddr := crypto.RequestIdToContractAddress(reqId)
	sReqId := hex.EncodeToString(reqId[:])
	log.Debug("-----Ccdeploytx:", "reqId", sReqId, "depId", contractAddr.String())
	rsp := &ContractDeployRsp{
		ReqId:      sReqId,
		ContractId: contractAddr.String(),
	}
	return rsp, err
}

func (s *PublicContractAPI) Ccinvoketx(ctx context.Context, from, to, daoAmount, daoFee, deployId string, param []string, certID string, timeout string) (*ContractDeployRsp, error) {
	contractAddr, _ := common.StringToAddress(deployId)

	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)
	timeout64, _ := strconv.ParseUint(timeout, 10, 64)

	log.Debug("-----Ccinvoketx:", "contractId", contractAddr.String())
	log.Debug("-----Ccinvoketx:", "fromAddr", fromAddr.String())
	log.Debug("-----Ccinvoketx:", "toAddr", toAddr.String())
	log.Debug("-----Ccinvoketx:", "daoAmount", daoAmount)
	log.Debug("-----Ccinvoketx:", "amount", amount)
	log.Debug("-----Ccinvoketx:", "fee", fee)
	log.Debug("-----Ccinvoketx:", "param len", len(param))
	log.Debug("-----Ccinvoketx:", "timeout64", timeout64)
	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return &ContractDeployRsp{}, fmt.Errorf("certid is invalid")
		}
		log.Debug("-----Ccinvoketx:", "certificate serial number", certID)
	}
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	reqId, err := s.b.ContractInvokeReqTx(fromAddr, toAddr, amount, fee, intCertID, contractAddr, args, uint32(timeout64))
	log.Debug("-----ContractInvokeTxReq:" + hex.EncodeToString(reqId[:]))
	rsp1 := &ContractDeployRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: deployId,
	}
	return rsp1, err
}

func (s *PublicContractAPI) CcinvokeToken(ctx context.Context, from, to, toToken, daoAmount, daoFee, assetToken, amountToken, deployId string, param []string) (*ContractDeployRsp, error) {
	contractAddr, _ := common.StringToAddress(deployId)

	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	toAddrToken, _ := common.StringToAddress(toToken)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)
	amountOfToken, _ := strconv.ParseUint(amountToken, 10, 64)

	log.Debug("-----CcinvokeToken:", "contractId", contractAddr.String())
	log.Debug("-----CcinvokeToken:", "fromAddr", fromAddr.String())
	log.Debug("-----CcinvokeToken:", "toAddr", toAddr.String())
	log.Debug("-----CcinvokeToken:", "amount", amount)
	log.Debug("-----CcinvokeToken:", "fee", fee)
	log.Debug("-----CcinvokeToken:", "toAddrToken", toAddrToken.String())
	log.Debug("-----CcinvokeToken:", "amountVote", amountOfToken)
	log.Debug("-----CcinvokeToken:", "param len", len(param))

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	reqId, err := s.b.ContractInvokeReqTokenTx(fromAddr, toAddr, toAddrToken, amount, fee, amountOfToken, assetToken, contractAddr, args, 0)
	log.Debug("-----ContractInvokeTxReq:" + hex.EncodeToString(reqId[:]))
	rsp1 := &ContractDeployRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: deployId,
	}
	return rsp1, err
}

func (s *PublicContractAPI) CcinvoketxPass(ctx context.Context, from, to, daoAmount, daoFee, deployId string, param []string, password string, duration *uint64, certID string) (string, error) {
	contractAddr, _ := common.StringToAddress(deployId)

	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)

	log.Debug("-----CcinvoketxPass:", "contractId", contractAddr.String())
	log.Debug("-----CcinvoketxPass:", "fromAddr", fromAddr.String())
	log.Debug("-----CcinvoketxPass:", "toAddr", toAddr.String())
	log.Debug("-----CcinvoketxPass:", "amount", amount)
	log.Debug("-----CcinvoketxPass:", "fee", fee)

	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return "", fmt.Errorf("certid is invalid")
		}
		log.Debug("-----CcinvoketxPass:", "certificate serial number", certID)
	}
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}

	//2.
	err := s.unlockKS(fromAddr, password, duration)
	if err != nil {
		return "", err
	}

	reqId, err := s.b.ContractInvokeReqTx(fromAddr, toAddr, amount, fee, intCertID, contractAddr, args, 0)
	log.Debug("-----ContractInvokeTxReq:" + hex.EncodeToString(reqId[:]))

	return hex.EncodeToString(reqId[:]), err
}

func (s *PublicContractAPI) Ccstoptx(ctx context.Context, from, to, daoAmount, daoFee, contractId string) (string, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)
	contractAddr, _ := common.StringToAddress(contractId)
	//TODO delImg 为 true 时，目前是会删除基础镜像的
	//delImg := true
	//if del, _ := strconv.Atoi(deleteImage); del <= 0 {
	//	delImg = false
	//}
	log.Debug("-----Ccstoptx:", "fromAddr", fromAddr.String())
	log.Debug("-----Ccstoptx:", "toAddr", toAddr.String())
	log.Debug("-----Ccstoptx:", "amount", amount)
	log.Debug("-----Ccstoptx:", "fee", fee)
	log.Debug("-----Ccstoptx:", "contractId", contractAddr)

	reqId, err := s.b.ContractStopReqTx(fromAddr, toAddr, amount, fee, contractAddr, false)
	log.Debug("-----Ccstoptx:" + hex.EncodeToString(reqId[:]))
	return hex.EncodeToString(reqId[:]), err
}

func (s *PublicContractAPI) unlockKS(addr common.Address, password string, duration *uint64) error {
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
		errors.New("get addr by outpoint is err")
		return err
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
func (s *PublicContractAPI) DepositContractInvoke(ctx context.Context, from, to, daoAmount, daoFee string,
	param []string) (string, error) {
	log.Debug("---enter DepositContractInvoke---")
	// append by albert·gou
	if param[0] == modules.ApplyMediator {
		//return "", fmt.Errorf("please use mediator.apply()")
		var args MediatorCreateArgs
		err := json.Unmarshal([]byte(param[1]), &args)
		if err != nil {
			return "", fmt.Errorf("param error(%v), please use mediator.apply()", err.Error())
		} else {
			// 参数补全
			args.setDefaults(from)

			// 参数验证
			err := args.Validate()
			if err != nil {
				return "", fmt.Errorf("error(%v), please use mediator.apply()", err.Error())
			}

			// 参数序列化
			argsB, err := json.Marshal(args)
			if err != nil {
				return "", fmt.Errorf("error(%v), please use mediator.apply()", err.Error())
			}

			param[1] = string(argsB)
		}
	}

	rsp, err := s.Ccinvoketx(ctx, from, to, daoAmount, daoFee, syscontract.DepositContractAddress.String(),
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

func (s *PublicContractAPI) SysConfigContractInvoke(ctx context.Context, from, to, daoAmount, daoFee string,
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
		err := core.CheckSysConfigArgs(field, value)
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
				err := core.CheckSysConfigArgs(oneTopic.TopicTitle, oneOption)
				if err != nil {
					log.Debugf(err.Error())
					return "", err
				}
			}
		}
	}

	rsp, err := s.Ccinvoketx(ctx, from, to, daoAmount, daoFee, syscontract.SysConfigContractAddress.String(),
		param, "", "0")

	return rsp.ReqId, err
}

//  TODO
func (s *PublicContractAPI) GetContractState(contractid []byte, key string) ([]byte, *modules.StateVersion, error) {
	return s.b.GetContractState(contractid, key)
}
