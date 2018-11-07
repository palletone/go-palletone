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
 * @author PalletOne core developer YiRan <dev@pallet.one>
 * @date 2018
 */

package vote

import (
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/stretchr/testify/assert"
)

func TestAddressMultipleVote(t *testing.T) {
	amv := AddressMultipleVote{}
	addrs := make([]common.Address, 0)
	addr1 := common.StringToAddressGodBlessMe("P1GqZ72gaeq7LiS34KLJoMmCnMnaopkcEPn")
	addr2 := common.StringToAddressGodBlessMe("P1L3F4oj1ciogAE69uogGcU8e9Hp5ZMnYJ3")
	addr3 := common.StringToAddressGodBlessMe("P1KYtxHobTsYgR4cWF5rjb5WUM7ZkDncHa9")
	addr4 := common.StringToAddressGodBlessMe("P1M2v9vvP5UJAtW4vQPqPSjsLPxnzgnP9UT")
	addr5 := common.StringToAddressGodBlessMe("P1JT8D85jFajyKguB1DvsaYERv9K8y8vckL")
	addr6 := common.StringToAddressGodBlessMe("P1PjSaHLTxFm52fECLxVFErd3ch8Fif7CEN")
	addrs = append(addrs, addr1, addr2, addr3, addr4, addr5)
	amv.RegisterCandidates(addrs)
	amv.AddToBox(100, []common.Address{addr1})
	amv.AddToBox(200, []common.Address{addr2})
	amv.AddToBox(300, []common.Address{addr3})
	amv.AddToBox(400, []common.Address{addr4})
	amv.AddToBox(500, []common.Address{addr5})
	amv.AddToBox(5000, addrs)

	// [test1] test voting to invalid candidates
	amv.AddToBox(512, []common.Address{addr6})
	_, ok := amv.voteStatus[addr6]
	assert.False(t, ok)

	// [test2] test result
	voteResult := make([]common.Address, 0)
	amv.GetResult(4, &voteResult)
	addr1Score, err := amv.GetScore(voteResult[0])
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(voteResult))
	assert.EqualValues(t, addr5, voteResult[0])
	assert.EqualValues(t, addr4, voteResult[1])
	assert.EqualValues(t, addr3, voteResult[2])
	assert.EqualValues(t, addr2, voteResult[3])
	assert.EqualValues(t, 5500, addr1Score)
}

func TestDeligatePrivilegedVote(t *testing.T) {
	dpv := deligatePrivilegedVote{}
	addrs := make([]common.Address, 0)
	addr1 := common.StringToAddressGodBlessMe("P1GqZ72gaeq7LiS34KLJoMmCnMnaopkcEPn")
	addr2 := common.StringToAddressGodBlessMe("P1L3F4oj1ciogAE69uogGcU8e9Hp5ZMnYJ3")
	addr3 := common.StringToAddressGodBlessMe("P1KYtxHobTsYgR4cWF5rjb5WUM7ZkDncHa9")
	addr4 := common.StringToAddressGodBlessMe("P1M2v9vvP5UJAtW4vQPqPSjsLPxnzgnP9UT")
	addr5 := common.StringToAddressGodBlessMe("P1JT8D85jFajyKguB1DvsaYERv9K8y8vckL")
	addrs = append(addrs, addr1, addr2, addr3, addr4, addr5)
	dpv.RegisterCandidates(addrs)
	//dpv.SetWeight()

}
