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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/rwset"
)

type ContractResp struct {
	Err  error
	Resp interface{}
}

type ContractReqInf interface {
	do(rwM rwset.TxManager, v contracts.ContractInf) (interface{}, error)
}

type ContractInstallReq struct {
	chainID       string
	ccName        string
	ccPath        string
	ccVersion     string
	ccLanguage    string
	ccAbi         string
	ccDescription string
	addrHash      []common.Hash
}

func (req ContractInstallReq) do(rwM rwset.TxManager, v contracts.ContractInf) (interface{}, error) {
	return v.Install(req.chainID, req.ccName, req.ccPath, req.ccVersion, req.ccDescription, req.ccAbi, req.ccLanguage)
}

type ContractDeployReq struct {
	chainID    string
	templateId []byte
	txid       string
	args       [][]byte
	timeout    time.Duration
}

func (req ContractDeployReq) do(rwM rwset.TxManager, v contracts.ContractInf) (interface{}, error) {
	_, payload, err := v.Deploy(rwM, req.chainID, req.templateId, req.txid, req.args, req.timeout)
	return payload, err
}

type ContractInvokeReq struct {
	chainID  string
	deployId []byte
	txid     string //common.Hash
	args     [][]byte
	timeout  time.Duration
}

func (req ContractInvokeReq) do(rwM rwset.TxManager, v contracts.ContractInf) (interface{}, error) {
	return v.Invoke(rwM, req.chainID, req.deployId, req.txid, req.args, req.timeout)
}

type ContractStopReq struct {
	chainID     string
	deployId    []byte
	txid        string
	deleteImage bool
}

func (req ContractStopReq) do(rwM rwset.TxManager, v contracts.ContractInf) (interface{}, error) {
	return v.Stop(rwM, req.chainID, req.deployId, req.txid, req.deleteImage)
}

func ContractProcess(rwM rwset.TxManager, contract *contracts.Contract, req ContractReqInf) (interface{}, error) {
	if contract == nil || req == nil {
		log.Error("ContractProcess", "param is nil,", "err")
		return nil, errors.New("ContractProcess param is nil")
	}

	return req.do(rwM, contract)
}
