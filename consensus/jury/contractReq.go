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
	"sync"
	"time"

	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/contracts/comm"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/ucc"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
)

//deploy -->invoke
func (p *Processor) ContractQuery(id []byte, args [][]byte, timeout time.Duration) (rsp []byte, err error) {
	exist := false
	chainId := rwset.ChainId

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	rd1, _ := crypto.GetRandomBytes(32)
	rd2, _ := crypto.GetRandomBytes(32)
	depTxId := util.RlpHash(rd1)
	invTxId := util.RlpHash(rd2)

	addr, err := common.StringToAddress(string(id))
	if err != nil {
		return nil, err
	}

	log.Debugf("ContractQuery enter, addr[%s][%v]", addr.String(), id)

	if addr.IsSystemContractAddress() {
		log.Debugf("ContractQuery, is system contract, addr[%s]", addr.String())
	} else {
		//cons, err := utils.GetAllContainers(client)
		cons, err := p.pDocker.GetAllContainers()
		if err != nil {
			log.Errorf("ContractQuery, id[%s], GetAllContainers err:%s", addr.String(), err.Error())
			return nil, err
		}
		//cas, _ := utils.GetAllContainerAddr(cons, "Up")
		cas, _ := p.pDocker.GetAllContainersAddrsWithStatus(cons, "Up")
		for _, ca := range cas {
			name := ca[:35]
			contractAddr, _ := common.StringToAddress(name)
			if contractAddr.Equal(addr) { //use first
				log.Debugf("ContractQuery, contractId[%s],find container(Up)", addr.String())
				exist = true
				break
			}
		}
		if !exist {
			cc, err := p.dag.GetContract(addr.Bytes())
			if err != nil {
				log.Errorf("ContractQuery, GetContract err:%s ", err.Error())
				return nil, err
			}
			ct, err := p.dag.GetContractTpl(cc.TemplateId)
			if err != nil {
				log.Errorf("ContractQuery, GetContractTpl err:%s ", err.Error())
				return nil, err
			}
			cv := ct.Version + ":" + contractcfg.GetConfig().ContractAddress
			spec := &pb.PtnChaincodeSpec{
				Type: pb.PtnChaincodeSpec_Type(pb.PtnChaincodeSpec_Type_value["GOLANG"]),
				Input: &pb.PtnChaincodeInput{
					Args: [][]byte{},
				},
				ChaincodeId: &pb.PtnChaincodeID{
					Name:    addr.String(),
					Path:    ct.Path,
					Version: cv,
				},
			}
			cp := p.dag.GetChainParameters()
			spec.CpuQuota = cp.UccCpuQuota
			spec.CpuShare = cp.UccCpuShares
			spec.Memory = cp.UccMemory
			dag, err := comm.GetCcDagHand()
			if err != nil {
				log.Error("getCcDagHand err:", "error", err)
				return nil, err
			}
			_, chaincodeData, err := ucc.RecoverChainCodeFromDb(dag, cc.TemplateId)
			if err != nil {
				log.Error("ContractQuery", "chainid:", chainId, "templateId:", cc.TemplateId, "RecoverChainCodeFromDb err", err)
				return nil, err
			}
			err = ucc.DeployUserCC(addr.Bytes(), chaincodeData, spec, chainId, depTxId.String(), nil, timeout)
			if err != nil {
				log.Error("ContractQuery ", "DeployUserCC error", err)
				return nil, nil
			}
			//juryAddrs := p.GetLocalJuryAddrs()
			//juryAddr := ""
			//if len(juryAddrs) != 0 {
			//	juryAddr = juryAddrs[0].String()
			//}
			//cInf := &list.CCInfo{
			//	Id:       addr.Bytes(),
			//	Name:     addr.String(),
			//	Path:     ct.Path,
			//	TempleId: ct.TplId,
			//	Version:  cv,
			//	Language: ct.Language,
			//	SysCC:    false,
			//	Address:  juryAddr,
			//}
			//_,err = p.dag.GetContract(addr.Bytes())
			//if err != nil {
			//	err = p.dag.SaveChaincode(addr, cInf)
			//	if err != nil {
			//		log.Debugf("ContractQuery, SaveChaincode err:%s", err.Error())
			//	}
			//}
		}
	}

	log.Debugf("ContractQuery, begin to invoke contract:%s", addr.String())
	rwM, _ := rwset.NewRwSetMgr("query")
	ctx := &contracts.ContractProcessContext{RwM: rwM, Dag: p.dag}

	rst, err := p.contract.Invoke(ctx, chainId, addr.Bytes(), invTxId.String(), args, timeout)
	rwM.Close()
	if err != nil {
		log.Warnf("ContractQuery, id[%s], Invoke err:%s", addr.String(), err.Error())
		return nil, err
	}
	log.Debugf("ContractQuery, id[%s], query result:%s", addr.String(), hex.EncodeToString(rst.Payload))
	return rst.Payload, nil
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
			tx.RequestHash().ShortStr(), tx.Hash().String(), err.Error())
	}
	return err == nil
}
