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
	ContractId      string `json:"contract_id"` //Hex
	ContractAddress string `json:"contract_address"`
	TemplateId      string `json:"tpl_id"`
	Name            string `json:"contract_name"`
	//1Active 0Stopped
	Status       byte      `json:"status"` // 合约状态
	Creator      string    `json:"creator"`
	CreationTime time.Time `json:"creation_time"` // creation  date
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
	Memory         uint16   `json:"memory"`
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
		Memory:         tpl.Memory,
		AddrHash:       []string{},
	}
	for _, addH := range tpl.AddrHash {
		json.AddrHash = append(json.AddrHash, addH.String())
	}
	return json
}
