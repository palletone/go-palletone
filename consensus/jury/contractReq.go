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
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"

	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

func (p *Processor) ContractInstallReq(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string,
	description, abi, language string, local bool, addrs []common.Address) (reqId common.Hash, TplId []byte, err error) {
	if from == (common.Address{}) || to == (common.Address{}) || tplName == "" || path == "" || version == "" {
		log.Error("ContractInstallReq, param is error")
		return common.Hash{}, nil, errors.New("ContractInstallReq request param is error")
	}
	if len(tplName) > MaxLengthTplName || len(path) > MaxLengthTplPath || len(version) > MaxLengthTplVersion ||
		len(description) > MaxLengthDescription || len(abi) > MaxLengthAbi || len(language) > MaxLengthLanguage ||
		len(addrs) > MaxNumberTplEleAddrHash {
		log.Error("ContractInstallReq", "request param len overflow，len(tplName)",
			len(tplName), "len(path)", len(path), "len(version)", len(version), "len(description)", len(description),
			"len(abi)", len(abi), "len(language)", len(language), "len(addrs)", len(addrs))
		return common.Hash{}, nil, errors.New("ContractInstallReq, request param len overflow")
	}
	if !p.dag.IsContractDeveloper(from) {
		return common.Hash{}, nil, fmt.Errorf("ContractInstallReq, address[%s] is not developer", from.String())
	}
	if len(addrs) > 0 {
		jjhAd := p.dag.GetChainParameters().FoundationAddress
		if jjhAd != from.Str() {
			log.Debugf("ContractInstallReq, requestAddr[%s] not foundationAddress", from.String())
			return common.Hash{}, nil, fmt.Errorf("ContractInstallReq,"+
				" request address[%s] is not foundationAddress(when specifying the contract install address)", from.String())
		}
	}
	addrHash := make([]common.Hash, 0)
	//去重
	resultAddress := getValidAddress(addrs)
	for _, addr := range resultAddress {
		addrHash = append(addrHash, util.RlpHash(addr))
	}
	log.Debug("ContractInstallReq", "enter, tplName ", tplName, "path", path, "version", version, "addrHash", addrHash)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_TPL_REQUEST,
		Payload: &modules.ContractInstallRequestPayload{
			TplName:        tplName,
			Path:           path,
			Version:        version,
			AddrHash:       addrHash,
			TplDescription: description,
			Abi:            abi,
			Language:       language,
			Creator:        from.String(),
		},
	}
	reqId, _, err = p.createContractTxReq(common.Address{}, from, to, daoAmount, daoFee, nil, msgReq)
	if err != nil {
		return common.Hash{}, nil, err
	}
	isLocal := true //todo
	if isLocal {
		if err = p.runContractReq(reqId, nil); err != nil {
			return common.Hash{}, nil, err
		}

		ctx := p.mtx[reqId]
		ctx.rstTx, err = p.GenContractSigTransaction(from, "", ctx.rstTx, p.ptn.GetKeyStore())
		if err != nil {
			return common.Hash{}, nil, err
		}
		tx := ctx.rstTx
		tpl, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_TPL)
		if err != nil || tpl == nil {
			errMsg := fmt.Sprintf("[%s]getContractTxContractInfo fail, tpl Name[%s]", shortId(reqId.String()), tplName)
			return common.Hash{}, nil, errors.New(errMsg)
		}
		templateId := tpl.(*modules.ContractTplPayload).TemplateId
		log.Infof("[%s]ContractInstallReq ok, reqId[%s] templateId[%x]", shortId(reqId.String()), reqId.String(), templateId)
		//broadcast
		go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Tx: tx}, false)
		return reqId, templateId, nil
	}
	//net mode

	return reqId, nil, nil
}

func (p *Processor) ContractDeployReq(from, to common.Address, daoAmount, daoFee uint64, templateId []byte,
	args [][]byte, extData []byte, timeout time.Duration) (common.Hash, common.Address, error) {
	if from == (common.Address{}) || to == (common.Address{}) || templateId == nil {
		log.Error("ContractDeployReq, param is error")
		return common.Hash{}, common.Address{}, errors.New("ContractDeployReq request param is error")
	}
	if len(templateId) > MaxLengthTplId || len(args) > MaxNumberArgs || len(extData) > MaxLengthExtData {
		log.Error("ContractDeployReq", "request param len overflow, len(templateId)",
			len(templateId), "len(args)", len(args), "len(extData)", len(extData))
		return common.Hash{}, common.Address{}, errors.New("ContractDeployReq request param len overflow")
	}
	for _, arg := range args {
		if len(arg) > MaxLengthArgs {
			log.Error("ContractDeployReq", "request param len overflow,len(arg)", len(arg))
			return common.Hash{}, common.Address{}, errors.New("ContractDeployReq request param len overflow")
		}
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_DEPLOY_REQUEST,
		Payload: &modules.ContractDeployRequestPayload{
			TemplateId: templateId,
			Args:       args,
			ExtData:    extData,
			Timeout:    uint32(timeout),
		},
	}
	reqId, tx, err := p.createContractTxReq(common.Address{}, from, to, daoAmount, daoFee, nil, msgReq)
	if err != nil {
		return common.Hash{}, common.Address{}, err
	}
	contractId := crypto.RequestIdToContractAddress(reqId)
	log.Infof("[%s]ContractDeployReq ok, reqId[%s] templateId[%x],contractId[%s] ",
		shortId(reqId.String()), reqId.String(), templateId, contractId.String())

	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{Ele: nil, CType: CONTRACT_EVENT_ELE, Tx: tx}, true)
	return reqId, contractId, err
}

func (p *Processor) ContractInvokeReq(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	contractId common.Address, args [][]byte, timeout uint32) (common.Hash, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) || args == nil {
		log.Error("ContractInvokeReq, param is error")
		return common.Hash{}, errors.New("ContractInvokeReq request param is error")
	}
	if len(args) > MaxNumberArgs {
		log.Error("ContractInvokeReq", "len(args)", len(args))
		return common.Hash{}, errors.New("ContractInvokeReq request param len overflow")
	}
	for _, arg := range args {
		if len(arg) > MaxLengthArgs {
			log.Error("ContractInvokeReq", "request param len overflow,len(arg)", len(arg))
			return common.Hash{}, errors.New("ContractInvokeReq request param args len overflow")
		}
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: contractId.Bytes(),
			Args:       args,
			Timeout:    timeout,
		},
	}
	reqId, tx, err := p.createContractTxReq(contractId, from, to, daoAmount, daoFee, certID, msgReq)
	if err != nil {
		return common.Hash{}, err
	}
	log.Infof("[%s]ContractInvokeReq ok, reqId[%s], contractId[%s]",
		shortId(reqId.String()), reqId.String(), contractId.String())
	log.DebugDynamic(func() string {
		rjson, _ := json.Marshal(tx)
		rdata, _ := rlp.EncodeToBytes(tx)
		return fmt.Sprintf("Request data fro debug json:%s,\r\n rlp:%x", string(rjson), rdata)
	})
	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_EXEC, Ele: p.mtx[reqId].eleNode, Tx: tx}, true)
	return reqId, nil
}

func (p *Processor) ContractInvokeReqToken(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64,
	assetToken string, contractId common.Address, args [][]byte, timeout uint32) (common.Hash, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) || args == nil {
		log.Error("ContractInvokeReqToken, param is error")
		return common.Hash{}, errors.New("ContractInvokeReqToken request param is error")
	}
	if len(args) > MaxNumberArgs {
		log.Error("ContractInvokeReqToken", "len(args)", len(args))
		return common.Hash{}, errors.New("ContractInvokeReqToken request param len overflow")
	}
	for _, arg := range args {
		if len(arg) > MaxLengthArgs {
			log.Error("ContractInvokeReqToken", "request param len overflow,len(arg)", len(arg))
			return common.Hash{}, errors.New("ContractInvokeReqToken request param args len overflow")
		}
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: contractId.Bytes(),
			Args:       args,
			Timeout:    timeout,
		},
	}
	reqId, tx, err := p.createContractTxReqToken(contractId, from, to, toToken, daoAmount, daoFee,
		daoAmountToken, assetToken, msgReq)
	if err != nil {
		return common.Hash{}, err
	}
	log.Infof("[%s]ContractInvokeReqToken ok, reqId[%s] contractId[%s]",
		shortId(reqId.String()), reqId.String(), contractId.Bytes())
	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_EXEC, Ele: p.mtx[reqId].eleNode, Tx: tx}, true)
	return reqId, nil
}

func (p *Processor) ContractStopReq(from, to common.Address, daoAmount, daoFee uint64,
	contractId common.Address, deleteImage bool) (common.Hash, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) {
		log.Error("ContractStopReq, param is error")
		return common.Hash{}, errors.New("ContractStopReq request param is error")
	}
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		return common.Hash{}, errors.New("ContractStopReq, GetRandomNonce error")
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_STOP_REQUEST,
		Payload: &modules.ContractStopRequestPayload{
			ContractId:  contractId.Bytes(),
			Txid:        hex.EncodeToString(randNum),
			DeleteImage: deleteImage,
		},
	}
	reqId, tx, err := p.createContractTxReq(contractId, from, to, daoAmount, daoFee, nil, msgReq)
	if err != nil {
		return common.Hash{}, err
	}
	log.Infof("[%s]ContractStopReq ok, reqId[%s], contractId[%s], txId[%s]",
		shortId(reqId.String()), reqId.String(), contractId, hex.EncodeToString(randNum))
	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_EXEC, Ele: p.mtx[reqId].eleNode, Tx: tx}, true)
	return reqId, nil
}

func (p *Processor) ElectionVrfReq(id uint32) ([]byte, error) {
	reqId := util.RlpHash(id)
	p.mtx[reqId] = &contractTx{
		tm:     time.Now(),
		valid:  true,
		adaInf: make(map[uint32]*AdapterInf),
	}
	//p.ElectionRequest(reqId, time.Second*5)

	return nil, nil
}

func (p *Processor) UpdateJuryAccount(addr common.Address, pwd string) bool {
	acc := &JuryAccount{
		Address:  addr,
		Password: pwd,
	}
	accMap := make(map[common.Address]*JuryAccount)
	accMap[addr] = acc
	p.locker.Lock()
	defer p.locker.Unlock()

	p.local = nil
	p.local = accMap

	return true
}
func (p *Processor) CheckTxValid(tx *modules.Transaction) bool {
	_, _, err := p.validator.ValidateTx(tx, false)
	if err != nil {
		log.Debugf("[%s]checkTxValid, Validate fail, txHash[%s], err:%s",
			shortId(tx.RequestHash().String()), tx.Hash().String(), err.Error())
	}
	return err == nil
}
