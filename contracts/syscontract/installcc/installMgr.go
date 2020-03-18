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
 *  * @date 2018-2020
 *
 */

//PCGTta3M4t3yXu8uRgkKvaWd2d8DSHHyWEW
package installcc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

type InstallMgr struct {
}

func (p *InstallMgr) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

type ContractInstallRequestPayload struct {
	TplName        string        `json:"tpl_name"`
	TplDescription string        `json:"tpl_description"`
	Path           string        `json:"install_path"`
	Version        string        `json:"tpl_version"`
	Abi            string        `json:"abi"`
	Language       string        `json:"language"`
	AddrHash       []common.Hash `json:"addr_hash"`
	Creator        string        `json:"creator"`
}

func (p *InstallMgr) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "installByteCode": //安装合约模板
		if len(args) != 7 {
			return shim.Error("must input 7 args: name, description,codeBase64,version,abi,language,[]juryAddress")
		}
		name := args[0]
		description := args[1]
		code, err := base64.StdEncoding.DecodeString(args[2])
		if err != nil {
			return shim.Error("Invalid code base64 format")
		}
		version := args[3]
		abi := args[4]
		language := args[5]
		addrs := []common.Address{}
		err = json.Unmarshal([]byte(args[6]), &addrs)
		if err != nil {
			return shim.Error("Invalid jury address:" + args[6])
		}
		err = p.InstallByteCode(stub, name, description, code, version, abi, language, addrs)
		if err != nil {
			return shim.Error("InstallByteCode error:" + err.Error())
		}
		return shim.Success(nil)
	case "installRemoteCode": //安装远程代码
		if len(args) != 7 {
			return shim.Error("must input 7 args: name, description,codeBase64,version,abi,language,[]juryAddress")
		}
		name := args[0]
		description := args[1]
		codeUrl := args[2]
		version := args[3]
		abi := args[4]
		language := args[5]
		addrHashes := []common.Hash{}
		err := json.Unmarshal([]byte(args[6]), &addrHashes)
		if err != nil {
			return shim.Error("Invalid jury address hashed:" + args[6])
		}
		err = p.InstallRemoteCode(stub, name, description, codeUrl, version, abi, language, addrHashes)
		if err != nil {
			return shim.Error("InstallByteCode error:" + err.Error())
		}
		return shim.Success(nil)
	case "getTemplates": //列出合约模板列表
		result, err := p.GetTemplates(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		data, _ := json.Marshal(result)
		return shim.Success(data)

	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func (p *InstallMgr) InstallByteCode(stub shim.ChaincodeStubInterface, name, description string, code []byte,
	version, abi, language string, addrs []common.Address) error {
	if !isDeveloperInvoke(stub) && !isFoundationInvoke(stub) {
		return errors.New("only developer address can call this function")
	}
	if len(addrs) > 0 {
		if !isFoundationInvoke(stub) {
			return errors.New("only foundation can use static jury")
		}
	}
	addrHashes := []common.Hash{}
	for _, addr := range addrs {
		addrHashes = append(addrHashes, util.RlpHash(addr))
	}
	invokeAddr, _ := stub.GetInvokeAddress()
	tplId := getTemplateId(name, "", version)
	dbTpl, _ := getContractTemplate(stub, tplId)
	if dbTpl != nil {
		return fmt.Errorf("TemplateId[%x] already exist", tplId)
	}
	tpl := &modules.ContractTemplate{
		TplId:          tplId,
		TplName:        name,
		TplDescription: description,
		Path:           "",
		Version:        version,
		Abi:            abi,
		Language:       language,
		AddrHash:       addrHashes,
		Size:           uint16(len(code)),
		Creator:        invokeAddr.String(),
	}
	err := saveContractTemplate(stub, tpl)
	if err != nil {
		return err
	}
	err = saveContractTemplateCode(stub, tplId, code)
	if err != nil {
		return err
	}
	return nil
}
func getTemplateId(ccName, ccPath, ccVersion string) []byte {
	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := crypto.Keccak256Hash(buffer.Bytes())
	return tpid[:]
}
func getContractTemplate(stub shim.ChaincodeStubInterface, tplId []byte) (*modules.ContractTemplate, error) {
	key := "Tpl-" + hex.EncodeToString(tplId)
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	tpl := &modules.ContractTemplate{}
	err = json.Unmarshal(value, tpl)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}
func saveContractTemplate(stub shim.ChaincodeStubInterface, tpl *modules.ContractTemplate) error {
	key := "Tpl-" + hex.EncodeToString(tpl.TplId)
	value, err := json.Marshal(tpl)
	if err != nil {
		return err
	}
	return stub.PutState(key, value)
}
func saveContractTemplateCode(stub shim.ChaincodeStubInterface, tplId, code []byte) error {
	key := "Code-" + hex.EncodeToString(tplId)
	return stub.PutState(key, code)
}

func (p *InstallMgr) InstallRemoteCode(stub shim.ChaincodeStubInterface, name, description, url string,
	version, abi, language string, addrHashes []common.Hash) error {
	if !isDeveloperInvoke(stub) {
		return errors.New("only developer address can call this function")
	}
	if len(addrHashes) > 0 {
		if !isFoundationInvoke(stub) {
			return errors.New("only foundation can use static jury")
		}
	}
	invokeAddr, _ := stub.GetInvokeAddress()
	tplId := getTemplateId(name, url, version)
	dbTpl, _ := getContractTemplate(stub, tplId)
	if dbTpl != nil {
		return fmt.Errorf("TemplateId[%x] already exist", tplId)
	}
	code := downloadCode(url)

	tpl := &modules.ContractTemplate{
		TplId:          tplId,
		TplName:        name,
		TplDescription: description,
		Path:           "",
		Version:        version,
		Abi:            abi,
		Language:       language,
		AddrHash:       addrHashes,
		Size:           uint16(len(code)),
		Creator:        invokeAddr.String(),
	}
	err := saveContractTemplate(stub, tpl)
	if err != nil {
		return err
	}
	err = saveContractTemplateCode(stub, tplId, code)
	if err != nil {
		return err
	}
	return nil
}

func (p *InstallMgr) GetTemplates(stub shim.ChaincodeStubInterface) ([]*modules.ContractTemplate, error) {
	kvs, err := stub.GetStateByPrefix("Tpl-")
	if err != nil {
		return nil, err
	}
	result := []*modules.ContractTemplate{}
	for _, kv := range kvs {
		tpl := &modules.ContractTemplate{}
		err = json.Unmarshal(kv.Value, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, tpl)
	}
	return result, nil
}

//  判断是否Dev发起的
func isDeveloperInvoke(stub shim.ChaincodeStubInterface) bool {
	//  判断是否Dev发起的
	//TODO 基金会改为Developer
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return false
	}
	//  获取
	gp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := gp.ChainParameters.FoundationAddress
	// 判断当前请求的是否为基金会
	if invokeAddr.String() != foundationAddress {
		log.Error("please use foundation address")
		return false
	}
	return true
}

//  判断是否基金会发起的
func isFoundationInvoke(stub shim.ChaincodeStubInterface) bool {
	//  判断是否基金会发起的
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return false
	}
	//  获取
	gp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := gp.ChainParameters.FoundationAddress
	// 判断当前请求的是否为基金会
	if invokeAddr.String() != foundationAddress {
		log.Error("please use foundation address")
		return false
	}
	return true
}
func downloadCode(url string) []byte {
	return []byte{}
}
