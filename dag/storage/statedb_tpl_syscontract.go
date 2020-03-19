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
	"encoding/hex"
	"encoding/json"

	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/modules"
)

func (statedb *StateDb) GetContractTplFromSysContract(tplId []byte) (*modules.ContractTemplate, error) {
	id := syscontract.InstallContractAddress.Bytes()
	data, _, err := statedb.GetContractState(id, "Tpl-"+hex.EncodeToString(tplId))
	if err != nil {
		return nil, err
	}
	tpl := &modules.ContractTemplate{}
	err = json.Unmarshal(data, tpl)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}
func (statedb *StateDb) GetContractTplCodeFromSysContract(tplId []byte) ([]byte, error) {
	id := syscontract.InstallContractAddress.Bytes()
	data, _, err := statedb.GetContractState(id, "Code-"+hex.EncodeToString(tplId))
	return data, err
}
func (statedb *StateDb) GetAllContractTplFromSysContract() ([]*modules.ContractTemplate, error) {
	id := syscontract.InstallContractAddress.Bytes()
	kvs, err := statedb.GetContractStatesByPrefix(id, "Tpl-")
	if err != nil {
		return nil, err
	}
	result := make([]*modules.ContractTemplate, 0, len(kvs))
	for _, v := range kvs {
		tpl := &modules.ContractTemplate{}
		err = json.Unmarshal(v.Value, tpl)
		if err != nil {
			return nil, err
		}
		result = append(result, tpl)
	}
	return result, nil
}
