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

func TestBaseVoteModel(t *testing.T) {
	bvm := NewBaseVoteModel()

	addrs := make([]common.Address, 0)
	addr1 := common.StringToAddressGodBlessMe("P1GqZ72gaeq7LiS34KLJoMmCnMnaopkcEPn")
	addr2 := common.StringToAddressGodBlessMe("P1L3F4oj1ciogAE69uogGcU8e9Hp5ZMnYJ3")
	addr3 := common.StringToAddressGodBlessMe("P1KYtxHobTsYgR4cWF5rjb5WUM7ZkDncHa9")
	addr4 := common.StringToAddressGodBlessMe("P1M2v9vvP5UJAtW4vQPqPSjsLPxnzgnP9UT")
	addr5 := common.StringToAddressGodBlessMe("P1JT8D85jFajyKguB1DvsaYERv9K8y8vckL")
	addrs = append(addrs, addr1, addr2, addr3, addr4, addr5)
	bvm.RegisterCandidates(addrs)

	// add single vote
	bvm.AddToBox(100, addr1)
	bvm.AddToBox(200, addr2)
	bvm.AddToBox(300, addr3)
	bvm.AddToBox(400, addr4)
	bvm.AddToBox(500, addr5)
	// add votes batch
	bvm.AddToBox(5000, addrs)

	// [test1] test voting to invalid candidates
	addr6 := common.StringToAddressGodBlessMe("P1PjSaHLTxFm52fECLxVFErd3ch8Fif7CEN")
	bvm.AddToBox(512, addr6)
	_, ok := bvm.candidatesStatus[addr6]
	assert.False(t, ok)

	// [test2] test result
	voteResult := make([]common.Address, 0)
	bvm.GetResult(4, &voteResult)
	addr1Score, err := bvm.GetScore(voteResult[0])
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(voteResult))
	assert.EqualValues(t, addr5, voteResult[0])
	assert.EqualValues(t, addr4, voteResult[1])
	assert.EqualValues(t, addr3, voteResult[2])
	assert.EqualValues(t, addr2, voteResult[3])
	assert.EqualValues(t, 5500, addr1Score)
}

func TestOpenVoteModel(t *testing.T) {
	ovm := NewOpenVoteModel()

	addrs := make([]common.Address, 0)
	addr1 := common.StringToAddressGodBlessMe("P1GqZ72gaeq7LiS34KLJoMmCnMnaopkcEPn")
	addr2 := common.StringToAddressGodBlessMe("P1L3F4oj1ciogAE69uogGcU8e9Hp5ZMnYJ3")
	addr3 := common.StringToAddressGodBlessMe("P1KYtxHobTsYgR4cWF5rjb5WUM7ZkDncHa9")
	addr4 := common.StringToAddressGodBlessMe("P1M2v9vvP5UJAtW4vQPqPSjsLPxnzgnP9UT")
	addr5 := common.StringToAddressGodBlessMe("P1JT8D85jFajyKguB1DvsaYERv9K8y8vckL")
	addrs = append(addrs, addr1, addr2, addr3, addr4, addr5)
	ovm.RegisterCandidates(addrs)


	ovm.SetCurrentVoter(addr1)
	ovm.AddToBox(100, addr1)
	ovm.SetCurrentVoter(addr2)
	ovm.AddToBox(200, addr2)
	ovm.SetCurrentVoter(addr3)
	ovm.AddToBox(300, addr3)
	ovm.SetCurrentVoter(addr4)
	ovm.AddToBox(400, addr4)
	ovm.SetCurrentVoter(addr5)
	ovm.AddToBox(500, addr5)

	// [test1] test voting to invalid candidates
	addr6 := common.StringToAddressGodBlessMe("P1PjSaHLTxFm52fECLxVFErd3ch8Fif7CEN")
	ovm.AddToBox(512, addr6)
	_, ok := ovm.candidatesStatus[addr6]
	assert.False(t, ok)
}
