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

package modules

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ContractDeployPayload2 struct {
	TemplateId []byte             `json:"template_id"`    // contract template id
	ContractId []byte             `json:"contract_id"`    // contract id
	Name       string             `json:"name"`           // the name for contract
	Args       [][]byte           `json:"args"`           // contract arguments list
	EleList    []ElectionInf      `json:"election_list"`  // contract jurors list
	ReadSet    []ContractReadSet  `json:"read_set"`       // the set data of read, and value could be any type
	WriteSet   []ContractWriteSet `json:"write_set"`      // the set data of write, and value could be any type
	ErrMsg     ContractError      `json:"contract_error"` // contract error message
}

func TestRlpDeploy(t *testing.T) {
	data, _ := hex.DecodeString("f3809400000000000000000000000000000000000000049553797374656d20436f6e666967204d616e61676572c0c0c0c0c28080")
	deploy := &ContractDeployPayload2{}
	err := rlp.DecodeBytes(data, deploy)
	assert.Nil(t, err)
	t.Logf("%#v", deploy)
}
