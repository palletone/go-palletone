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
	"github.com/palletone/go-palletone/contracts"
	"time"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

type ContractResp struct {
	Err  error
	Resp interface{}
}

type ContractReqInf interface {
	do(v contracts.ContractInf) (interface{}, error)
}

//////
type ContractInstallReq struct {
	chainID   string
	ccName    string
	ccPath    string
	ccVersion string
}

func (req ContractInstallReq) do(v contracts.ContractInf) (interface{}, error) {
	//var resp ContractResp

	//payload, err := v.Install(req.chainID, req.ccName, req.ccPath, req.ccVersion)
	//resp = ContractResp{err, payload}
	return v.Install(req.chainID, req.ccName, req.ccPath, req.ccVersion)
}

type ContractDeployReq struct {
	chainID    string
	templateId []byte
	txid       string
	args       [][]byte
	timeout    time.Duration
}

func (req ContractDeployReq) do(v contracts.ContractInf) (interface{}, error) {
	//var resp ContractResp
	//
	//_, payload, err := v.Deploy(req.chainID, req.templateId, req.txid, req.args, req.timeout)
	//resp = ContractResp{err, payload}
	//return resp
	 _,payload, err :=v.Deploy(req.chainID, req.templateId, req.txid, req.args, req.timeout)
	 return payload, err
}

type ContractInvokeReq struct {
	chainID  string
	deployId []byte
	txid     string //common.Hash
	args     [][]byte
	tx       *modules.Transaction
	timeout  time.Duration
}

func (req ContractInvokeReq) do(v contracts.ContractInf) (interface{}, error) {
	//var resp ContractResp
	//
	//payload, err := v.Invoke(req.chainID, req.deployId, req.txid, req.tx, req.args, req.timeout)
	//resp = ContractResp{err, payload}
	//return resp
	 payload, err:= v.Invoke(req.chainID, req.deployId, req.txid, req.tx, req.args, req.timeout)

	return payload.ToContractInvokePayload(), err
}

type ContractStopReq struct {
	chainID     string
	deployId    []byte
	txid        string
	deleteImage bool
}

func (req ContractStopReq) do(v contracts.ContractInf) (interface{}, error) {
	//var resp ContractResp
	//
	//err := v.Stop(req.chainID, req.deployId, req.txid, req.deleteImage)
	//resp = ContractResp{err, nil}
	//return resp
	return nil, v.Stop(req.chainID, req.deployId, req.txid, req.deleteImage)
}

func ContractProcess(contract *contracts.Contract, req ContractReqInf) (interface{}, error) {
	if contract == nil || req == nil {
		log.Error("ContractProcess", "param is nil,", "err")
		return nil, errors.New("ContractProcess param is nil")
	}
	//var resp interface{}
	return req.do(contract)
	//return resp, nil

	//var resp interface{}
	//c := make(chan struct{})
	//
	//go func() {
	//	defer close(c)
	//	resp = req.do(contract)
	//}()
	//
	//select {
	//case <-c:
	//	log.Info("ContractProcess", "req resp", "ok")
	//	return resp, nil
	//}

	return nil, nil
}
