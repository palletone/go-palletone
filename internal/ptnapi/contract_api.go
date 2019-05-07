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
	"errors"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core/accounts"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"time"
)

type PublicContractAPI struct {
	b Backend
}

func NewPublicContractAPI(b Backend) *PublicContractAPI {
	return &PublicContractAPI{b}
}

//contract command
//install
func (s *PublicContractAPI) Ccinstall(ctx context.Context, ccname string, ccpath string, ccversion string) (hexutil.Bytes, error) {
	log.Info("CcInstall:" + ccname + ":" + ccpath + "_" + ccversion)

	templateId, err := s.b.ContractInstall(ccname, ccpath, ccversion)
	return hexutil.Bytes(templateId), err
}

func (s *PublicContractAPI) Ccdeploy(ctx context.Context, templateId string, txid string, param []string) (hexutil.Bytes, error) {
	tempId, _ := hex.DecodeString(templateId)

	//log.Info("Ccdeploy:" + templateId + ":" + txid)
	//fmt.Printf("templateid=%v", tempId)
	//fmt.Printf("-----------------parm len=%d", len(param))

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		//fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	//f := "init"
	//args := ut.ToChaincodeArgs(f, "a", "100", "b", "200")
	deployId, err := s.b.ContractDeploy(tempId, txid, args, 30*time.Second)
	return hexutil.Bytes(deployId), err
}

//func (s *PublicBlockChainAPI) Ccinvoke(ctx context.Context, txhex string) (string, error) {
//	txBytes, _ := hex.DecodeString(txhex)
//	rsp, err := s.b.ContractInvoke(txBytes)
//	log.Info("-----ContractInvokeTxReq:" + hex.EncodeToString(rsp))
//	return hex.EncodeToString(rsp), err
//}

func (s *PublicContractAPI) Ccinvoke(ctx context.Context, deployId string, txid string, param []string /*fun string, key string, val string*/) (string, error) {
	depId, _ := hex.DecodeString(deployId)
	log.Info("-----Ccinvoke:" + deployId + ":" + txid)

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		//fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	//参数前面加入msg0和msg1,这里为空
	var fullArgs [][]byte
	msgArg := []byte("query has no msg0")
	msgArg1 := []byte("query has no msg1")
	fullArgs = append(fullArgs, msgArg)
	fullArgs = append(fullArgs, msgArg1)
	fullArgs = append(fullArgs, args...)
	rsp, err := s.b.ContractInvoke(depId, txid, fullArgs, 0)
	log.Info("-----ContractInvokeTxReq:" + hex.EncodeToString(rsp))
	return string(rsp), err
}

func (s *PublicContractAPI) Ccquery(ctx context.Context, deployId string, param []string) (string, error) {
	contractId, _ := common.StringToAddress(deployId)
	log.Info("-----Ccquery:", "contractId", contractId.String())
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		//fmt.Printf("index[%d],value[%s]\n", i, arg)
	}
	//参数前面加入msg0和msg1,这里为空
	var fullArgs [][]byte
	msgArg := []byte("query has no msg0")
	msgArg1 := []byte("query has no msg1")
	fullArgs = append(fullArgs, msgArg)
	fullArgs = append(fullArgs, msgArg1)
	fullArgs = append(fullArgs, args...)

	txid := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))

	rsp, err := s.b.ContractQuery(contractId.Bytes21(), txid[:], fullArgs, 0)
	if err != nil {
		return "", err
	}
	return string(rsp), nil
}

func (s *PublicContractAPI) Ccstop(ctx context.Context, deployId string, txid string) error {
	depId, _ := hex.DecodeString(deployId)
	log.Info("Ccstop:" + deployId + ":" + txid + "_")
	//TODO deleteImage 为 true 时，目前是会删除基础镜像的
	err := s.b.ContractStop(depId, txid, false)
	return err
}

func (s *PublicContractAPI) Ccinstalltx(ctx context.Context, from, to, daoAmount, daoFee, tplName, path, version string, addr []string) (*ContractInstallRsp, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)

	//templateName, _ := hex.DecodeString(tplName)

	log.Info("-----Ccinstalltx:", "fromAddr", fromAddr.String())
	log.Info("-----Ccinstalltx:", "toAddr", toAddr.String())
	log.Info("-----Ccinstalltx:", "amount", amount)
	log.Info("-----Ccinstalltx:", "fee", fee)
	log.Info("-----Ccinstalltx:", "tplName", tplName)
	log.Info("-----Ccinstalltx:", "path", path)
	log.Info("-----Ccinstalltx:", "version", version)

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

	reqId, tplId, err := s.b.ContractInstallReqTx(fromAddr, toAddr, amount, fee, tplName, path, version, addrs)
	sReqId := hex.EncodeToString(reqId[:])
	sTplId := hex.EncodeToString(tplId)
	log.Info("-----Ccinstalltx:", "reqId", sReqId, "tplId", sTplId)

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

	log.Info("-----Ccdeploytx:", "fromAddr", fromAddr.String())
	log.Info("-----Ccdeploytx:", "toAddr", toAddr.String())
	log.Info("-----Ccdeploytx:", "amount", amount)
	log.Info("-----Ccdeploytx:", "fee", fee)
	log.Info("-----Ccdeploytx:", "tplId", templateId)

	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	reqId, depId, err := s.b.ContractDeployReqTx(fromAddr, toAddr, amount, fee, templateId, args, 0)
	addDepId := common.NewAddress(depId, common.ContractHash)
	sReqId := hex.EncodeToString(reqId[:])
	log.Info("-----Ccdeploytx:", "reqId", sReqId, "depId", addDepId.String())
	rsp := &ContractDeployRsp{
		ReqId:      sReqId,
		ContractId: addDepId.String(),
	}
	return rsp, err
}

func (s *PublicContractAPI) DepositContractInvoke(ctx context.Context, from, to, daoAmount, daoFee string, param []string) (string, error) {
	log.Info("---enter DepositContractInvoke---")
	rsp, err := s.Ccinvoketx(ctx, from, to, daoAmount, daoFee, syscontract.DepositContractAddress.String(), param, "")
	return rsp.ReqId, err
}
func (s *PublicContractAPI) DepositContractQuery(ctx context.Context, param []string) (string, error) {
	log.Info("---enter DepositContractQuery---")
	return s.Ccquery(ctx, syscontract.DepositContractAddress.String(), param)
}

func (s *PublicContractAPI) Ccinvoketx(ctx context.Context, from, to, daoAmount, daoFee, deployId string, param []string, certID string) (*ContractDeployRsp, error) {
	contractAddr, _ := common.StringToAddress(deployId)

	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)

	log.Info("-----Ccinvoketx:", "contractId", contractAddr.String())
	log.Info("-----Ccinvoketx:", "fromAddr", fromAddr.String())
	log.Info("-----Ccinvoketx:", "toAddr", toAddr.String())
	log.Info("-----Ccinvoketx:", "daoAmount", daoAmount)
	log.Info("-----Ccinvoketx:", "amount", amount)
	log.Info("-----Ccinvoketx:", "fee", fee)
	log.Info("-----Ccinvoketx:", "param len", len(param))
	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return &ContractDeployRsp{}, fmt.Errorf("certid is invalid")
		}
		log.Debugf("-----Ccinvoketx:", "certificate serial number", certID)
	}
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	reqId, err := s.b.ContractInvokeReqTx(fromAddr, toAddr, amount, fee, intCertID, contractAddr, args, 0)
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

	log.Info("-----CcinvokeToken:", "contractId", contractAddr.String())
	log.Info("-----CcinvokeToken:", "fromAddr", fromAddr.String())
	log.Info("-----CcinvokeToken:", "toAddr", toAddr.String())
	log.Info("-----CcinvokeToken:", "amount", amount)
	log.Info("-----CcinvokeToken:", "fee", fee)
	log.Info("-----CcinvokeToken:", "toAddrToken", toAddrToken.String())
	log.Info("-----CcinvokeToken:", "amountVote", amountOfToken)
	log.Info("-----CcinvokeToken:", "param len", len(param))

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

	log.Info("-----CcinvoketxPass:", "contractId", contractAddr.String())
	log.Info("-----CcinvoketxPass:", "fromAddr", fromAddr.String())
	log.Info("-----CcinvoketxPass:", "toAddr", toAddr.String())
	log.Info("-----CcinvoketxPass:", "amount", amount)
	log.Info("-----CcinvoketxPass:", "fee", fee)

	intCertID := new(big.Int)
	if len(certID) > 0 {
		if _, ok := intCertID.SetString(certID, 10); !ok {
			return "", fmt.Errorf("certid is invalid")
		}
		log.Info("-----CcinvoketxPass:", "certificate serial number", certID)
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

func (s *PublicContractAPI) Ccstoptx(ctx context.Context, from, to, daoAmount, daoFee, contractId, deleteImage string) (string, error) {
	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)
	contractAddr, _ := common.StringToAddress(contractId)
	//TODO delImg 为 true 时，目前是会删除基础镜像的
	delImg := true
	if del, _ := strconv.Atoi(deleteImage); del <= 0 {
		delImg = false
	}
	log.Info("-----Ccstoptx:", "fromAddr", fromAddr.String())
	log.Info("-----Ccstoptx:", "toAddr", toAddr.String())
	log.Info("-----Ccstoptx:", "amount", amount)
	log.Info("-----Ccstoptx:", "fee", fee)
	log.Info("-----Ccstoptx:", "contractId", contractAddr)
	log.Info("-----Ccstoptx:", "delImg", delImg)

	reqId, err := s.b.ContractStopReqTx(fromAddr, toAddr, amount, fee, contractAddr, delImg)
	log.Info("-----Ccstoptx:" + hex.EncodeToString(reqId[:]))
	return hex.EncodeToString(reqId[:]), err
}
