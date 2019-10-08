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

package ptnjson

import (
	"encoding/hex"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

type ContractJson struct {
	//Contract Id 即Address，[20]byte，
	// 根据用户创建合约实例申请的RequestId截取其后20字节生成
	ContractId      string                `json:"contract_id"` //Hex
	ContractAddress string                `json:"contract_address"`
	TemplateId      string                `json:"tpl_id"`
	Name            string                `json:"contract_name"`
	Status          byte                  `json:"status"` // 合约状态
	Creator         string                `json:"creator"`
	CreationTime    time.Time             `json:"creation_time"` // creation date
	DuringTime      uint64                `json:"during_time"`   // deploy during date
	Template        *ContractTemplateJson `json:"template"`
}

func ConvertContract2Json(contract *modules.Contract) *ContractJson {
	addr := common.NewAddress(contract.ContractId, common.ContractHash)
	creatorAddr := common.NewAddress(contract.Creator, common.PublicKeyHash)
	return &ContractJson{
		ContractId:      hex.EncodeToString(contract.ContractId),
		ContractAddress: addr.String(),
		TemplateId:      hex.EncodeToString(contract.TemplateId),
		Name:            contract.Name,
		Status:          contract.Status,
		Creator:         creatorAddr.String(),
		CreationTime:    time.Unix(int64(contract.CreationTime), 0),
		DuringTime:      contract.DuringTime,
	}
}

type ContractTemplateJson struct {
	TplId          string   `json:"tpl_id"`
	TplName        string   `json:"tpl_name"`
	TplDescription string   `json:"tpl_description"`
	Path           string   `json:"install_path"`
	Version        string   `json:"tpl_version"`
	Abi            string   `json:"abi"`
	Language       string   `json:"language"`
	AddrHash       []string `json:"addr_hash" rlp:"nil"`
	Size           uint16   `json:"size"`
	Creator        string   `json:"creator"`
}

func ConvertContractTemplate2Json(tpl *modules.ContractTemplate) *ContractTemplateJson {

	json := &ContractTemplateJson{
		TplId:          hex.EncodeToString(tpl.TplId),
		TplName:        tpl.TplName,
		TplDescription: tpl.TplDescription,
		Path:           tpl.Path,
		Version:        tpl.Version,
		Abi:            tpl.Abi,
		Language:       tpl.Language,
		Size:           tpl.Size,
		AddrHash:       []string{},
		Creator:        tpl.Creator,
	}
	for _, addH := range tpl.AddrHash {
		json.AddrHash = append(json.AddrHash, addH.String())
	}
	return json
}

const PRC20_ABI = `[{"constant": false,"inputs": [{"name": "Name","type": "string"},{"name": "Name","type": "string"},{"name": "Decimals","type": "string"},{"name": "TotalSupply","type": "string"},{"name": "SupplyAddress","type": "string"}],"name": "createToken","outputs": [{"name": "","type": "string"}],"payable": false,"stateMutability": "nonpayable","type": "function"},{"constant": false,"inputs": [{"name": "Symbol","type": "string"},{"name": "SupplyAmout","type": "string"}],"name": "supplyToken","outputs": [{"name": "","type": "string"}],"payable": false,"stateMutability": "nonpayable","type": "function"},{"constant": false,"inputs": [{"name": "Symbol","type": "string"},{"name": "NewSupplyAddr","type": "string"}],"name": "changeSupplyAddr","outputs": [{"name": "","type": "string"}],"payable": false,"stateMutability": "nonpayable","type": "function"},{"constant": false,"inputs": [{"name": "Symbol","type": "string"}],"name": "frozenToken","outputs": [{"name": "","type": "string"}],"payable": false,"stateMutability": "nonpayable","type": "function"},{"constant": true,"inputs": [{"name": "Symbol","type": "string"}],"name": "getTokenInfo","outputs": [{"name": "","type": "string"}],"payable": false,"stateMutability": "view","type": "function"},{"constant": true,"inputs": [],"name": "getAllTokenInfo","outputs": [{"name": "","type": "string"}],"payable": false,"stateMutability": "view","type": "function"}]`

func GetSysContractTemplate_PRC20() *ContractTemplateJson {
	json := &ContractTemplateJson{
		TplId:          "",
		TplName:        "PRC20",
		TplDescription: "Fungible Token",
		Path:           "",
		Version:        "v1.0.0",
		Abi:            PRC20_ABI,
		Language:       "Golang",
		Size:           0,
		AddrHash:       []string{},
		Creator:        "",
	}
	return json
}
