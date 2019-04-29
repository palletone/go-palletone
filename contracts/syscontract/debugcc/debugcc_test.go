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

package debugcc

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPutState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	stub := shim.NewMockChaincodeStubInterface(mockCtrl)
	db := make(map[string][]byte)
	stub.EXPECT().PutState(gomock.Any(), gomock.Any()).Do(func(key string, value []byte) {
		db[key] = value
	}).AnyTimes()
	stub.EXPECT().GetState(gomock.Any()).DoAndReturn(func(key string) ([]byte, error) {
		value, ok := db[key]
		if ok {
			return value, nil
		} else {
			return nil, errors.New("not found")
		}
	}).AnyTimes()
	cc := &DebugChainCode{}
	cc.addBalance(stub, []string{"a", "100"})
	result := cc.getBalance(stub, []string{"a"})
	assert.Equal(t, result.Payload, []byte("100"))

	cc.addBalance(stub, []string{"a", "100"})
	result = cc.getBalance(stub, []string{"a"})
	assert.Equal(t, result.Payload, []byte("200"))
}
