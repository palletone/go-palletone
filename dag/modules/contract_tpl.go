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
 * @date 2018-2019
 */

package modules

import "github.com/palletone/go-palletone/common"

type ContractTemplate struct {
	TplId          []byte        `json:"tpl_id"`
	TplName        string        `json:"tpl_name"`
	TplDescription string        `json:"tpl_description"`
	Path           string        `json:"install_path"`
	Version        string        `json:"tpl_version"`
	Abi            string        `json:"abi"`
	Language       string        `json:"language"`
	AddrHash       []common.Hash `json:"addr_hash" rlp:"nil"`
	Size           uint16        `json:"size"`
	Creator        string        `json:"creator"`
}

func NewContractTemplate(req *ContractInstallRequestPayload, tpl *ContractTplPayload) *ContractTemplate {
	return &ContractTemplate{
		TplId:          tpl.TemplateId,
		TplName:        req.TplName,
		TplDescription: req.TplDescription,
		Path:           req.Path,
		Version:        req.Version,
		Abi:            req.Abi,
		Language:       req.Language,
		AddrHash:       req.AddrHash,
		Size:           uint16(len(tpl.ByteCode[:])),
		Creator:        req.Creator,
	}
}
