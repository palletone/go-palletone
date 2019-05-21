/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func (statedb *StateDb) SaveContractTpl(tpl *modules.ContractTemplate) error {
	key := append(constants.CONTRACT_TPL, tpl.TplId...)
	if err := StoreToRlpBytes(statedb.db, key, tpl); err != nil {
		return err
	}

	return nil
}
func (statedb *StateDb) SaveContractTplCode(tplId []byte, byteCode []byte) error {
	key := append(constants.CONTRACT_TPL_CODE, tplId...)
	return statedb.db.Put(key, byteCode)
}
func (statedb *StateDb) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	key := append(constants.CONTRACT_TPL, tplId...)
	tpl := &modules.ContractTemplate{}
	err := RetrieveFromRlpBytes(statedb.db, key, tpl)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}
func (statedb *StateDb) GetContractTplCode(tplId []byte) ([]byte, error) {
	key := append(constants.CONTRACT_TPL_CODE, tplId...)
	return statedb.db.Get(key)
}
func (statedb *StateDb) GetAllContractTpl() ([]*modules.ContractTemplate, error) {
	rows := getprefix(statedb.db, constants.CONTRACT_TPL)
	result := make([]*modules.ContractTemplate, 0,len(rows))
	for _, v := range rows {
		tpl := &modules.ContractTemplate{}
		err := rlp.DecodeBytes(v, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, tpl)
	}
	return result, nil
}
