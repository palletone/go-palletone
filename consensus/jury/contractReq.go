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
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

func (p *Processor) ContractInstallReq(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string, local bool, addrs []common.Address) (reqId common.Hash, TplId []byte, err error) {
	if from == (common.Address{}) || to == (common.Address{}) || tplName == "" || path == "" || version == "" {
		log.Error("ContractInstallReq", "param is error")
		return common.Hash{}, nil, errors.New("ContractInstallReq request param is error")
	}
	if len(tplName) > 64 || len(path) > 512 || len(version) > 32 || len(addrs) > 5 {
		log.Error("ContractInstallReq", "len(tplName)", len(tplName), "len(path)", len(path), "len(version)", len(version), "len(addrs)", len(addrs))
		return common.Hash{}, nil, errors.New("ContractInstallReq request param len overflow")
	}
	addrHash := make([]common.Hash, 0)
	for _, addr := range addrs {
		addrHash = append(addrHash, util.RlpHash(addr))
	}
	log.Debug("ContractInstallReq", "enter, tplName ", tplName, "path", path, "version", version, "addrHash", addrHash)

	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_TPL_REQUEST,
		Payload: &modules.ContractInstallRequestPayload{
			TplName:  tplName,
			Path:     path,
			Version:  version,
			AddrHash: addrHash,
		},
	}
	reqId, tx, err := p.createContractTxReq(common.Address{}, from, to, daoAmount, daoFee, nil, msgReq, true)
	if err != nil {
		return common.Hash{}, nil, err
	}
	tpl, err := getContractTxContractInfo(tx, modules.APP_CONTRACT_TPL)
	if err != nil || tpl == nil {
		errMsg := fmt.Sprintf("getContractTxContractInfo fail, tpl Name[%s]", tplName)
		return common.Hash{}, nil, errors.New(errMsg)
	}
	templateId := tpl.(*modules.ContractTplPayload).TemplateId

	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{CType: CONTRACT_EVENT_COMMIT, Tx: tx}, true)
	return reqId, templateId, nil
}

func (p *Processor) ContractDeployReq(from, to common.Address, daoAmount, daoFee uint64, templateId []byte, args [][]byte, timeout time.Duration) (common.Hash, []byte, error) {
	if from == (common.Address{}) || to == (common.Address{}) || templateId == nil {
		log.Error("ContractDeployReq", "param is error")
		return common.Hash{}, nil, errors.New("ContractDeployReq request param is error")
	}
	if len(templateId) > 128 || len(args) > 32 {
		log.Error("ContractDeployReq", "len(templateId)", len(templateId), "len(args)", len(args))
		return common.Hash{}, nil, errors.New("ContractDeployReq request param len overflow")
	}

	log.Debug("ContractDeployReq", "enter, templateId ", templateId)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_DEPLOY_REQUEST,
		Payload: &modules.ContractDeployRequestPayload{
			TplId:   templateId,
			Args:    args,
			Timeout: uint32(timeout),
		},
	}
	reqId, tx, err := p.createContractTxReq(common.Address{}, from, to, daoAmount, daoFee, nil, msgReq, false)
	if err != nil {
		return common.Hash{}, nil, err
	}
	contractId := common.BytesToAddress(reqId.Bytes()).Bytes()
	log.Debug("ContractDeployReq", "enter, templateId ", templateId, "contractId", contractId)

	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{Ele: p.mtx[reqId].eleInf, CType: CONTRACT_EVENT_EXEC, Tx: tx}, true)
	return reqId, contractId[:], err
}

func (p *Processor) ContractInvokeReq(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int, contractId common.Address, args [][]byte, timeout uint32) (common.Hash, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) || args == nil {
		log.Error("ContractInvokeReq", "param is error")
		return common.Hash{}, errors.New("ContractInvokeReq request param is error")
	}
	if len(args) > 64 {
		log.Error("ContractInvokeReq", "len(args)", len(args))
		return common.Hash{}, errors.New("ContractInvokeReq request param len overflow")
	}

	log.Debug("ContractInvokeReq", "enter, contractId ", contractId)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId:   contractId.Bytes21(),
			FunctionName: "",
			Args:         args,
			Timeout:      timeout,
		},
	}
	reqId, tx, err := p.createContractTxReq(contractId, from, to, daoAmount, daoFee, certID, msgReq, false)
	if err != nil {
		return common.Hash{}, err
	}
	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{Ele: p.mtx[reqId].eleInf, CType: CONTRACT_EVENT_EXEC, Tx: tx}, true)
	return reqId, nil
}

func (p *Processor) ContractInvokeReqToken(from, to, toToken common.Address, daoAmount, daoFee, daoAmountToken uint64, assetToken string, contractId common.Address, args [][]byte, timeout uint32) (common.Hash, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) || args == nil {
		log.Error("ContractInvokeReqToken", "param is error")
		return common.Hash{}, errors.New("ContractInvokeReqToken request param is error")
	}

	log.Debug("ContractInvokeReqToken", "enter, contractId ", contractId)
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId:   contractId.Bytes21(),
			FunctionName: "",
			Args:         args,
			Timeout:      timeout,
		},
	}
	reqId, tx, err := p.createContractTxReqToken(from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken, msgReq, false)
	if err != nil {
		return common.Hash{}, err
	}
	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{Ele: p.mtx[reqId].eleInf, CType: CONTRACT_EVENT_EXEC, Tx: tx}, true)
	return reqId, nil
}

func (p *Processor) ContractStopReq(from, to common.Address, daoAmount, daoFee uint64, contractId common.Address, deleteImage bool) (common.Hash, error) {
	if from == (common.Address{}) || to == (common.Address{}) || contractId == (common.Address{}) {
		log.Error("ContractStopReq", "param is error")
		return common.Hash{}, errors.New("ContractStopReq request param is error")
	}
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		return common.Hash{}, errors.New("ContractStopReq, GetRandomNonce error")
	}

	log.Debug("ContractStopReq", "enter, contractId ", contractId, "txId", hex.EncodeToString(randNum))
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_STOP_REQUEST,
		Payload: &modules.ContractStopRequestPayload{
			ContractId:  contractId[:],
			Txid:        hex.EncodeToString(randNum),
			DeleteImage: deleteImage,
		},
	}
	reqId, tx, err := p.createContractTxReq(contractId, from, to, daoAmount, daoFee, nil, msgReq, false)
	if err != nil {
		return common.Hash{}, err
	}
	//broadcast
	go p.ptn.ContractBroadcast(ContractEvent{Ele: p.mtx[reqId].eleInf, CType: CONTRACT_EVENT_EXEC, Tx: tx}, true)
	return reqId, nil
}

func (p *Processor) ElectionVrfReq(id uint32) ([]byte, error) {
	reqId := util.RlpHash(id)
	p.mtx[reqId] = &contractTx{
		tm:     time.Now(),
		valid:  true,
		adaInf: make(map[uint32]*AdapterInf),
	}
	p.ElectionRequest(reqId, time.Second*5)

	return nil, nil
}

func (p *Processor) UpdateJuryAccount(addr common.Address, pwd string) bool {
	acc := &JuryAccount{
		Address:  addr,
		Password: pwd,
	}
	accMap := make(map[common.Address]*JuryAccount, 0)
	accMap[addr] = acc
	p.locker.Lock()
	defer p.locker.Unlock()

	p.local = nil
	p.local = accMap

	return true
}

func (p *Processor) GetJuryAccount() []common.Address {
	num := len(p.local)
	addrs := make([]common.Address, num)
	i := 0
	for addr, _ := range p.local {
		addrs[i] = addr
		i++
	}
	return addrs
}
