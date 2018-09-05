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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package ccintf

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"

)

func TestGetName(t *testing.T) {
	spec := &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: "ccname"}}
	ccid := &CCID{ChaincodeSpec: spec, Version: "ver"}
	name := ccid.GetName()
	assert.Equal(t, "ccname-ver", name, "unexpected name")

	ccid.ChainID = GetCCHandlerKey()
	hash := hex.EncodeToString(util.ComputeSHA256([]byte(ccid.ChainID)))
	name = ccid.GetName()
	assert.Equal(t, "ccname-ver-"+hash, name, "unexpected name with hash")
}

func TestBadCCID(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("Should never reach here... GetName should have paniced")
		} else {
			assert.Equal(t, "nil chaincode spec", err, "expected 'nil chaincode spec'")
		}
	}()

	ccid := &CCID{Version: "ver"}
	ccid.GetName()
}
