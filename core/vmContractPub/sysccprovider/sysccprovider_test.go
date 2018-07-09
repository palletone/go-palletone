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

package sysccprovider

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var mockChaincodeProvider = MockChaincodeProvider{}

type MockFactory struct {
}

type MockChaincodeProvider struct {
}
//glh
//func (p MockChaincodeProvider) GetQueryExecutorForLedger(str string) (ledger.QueryExecutor, error) {
//	return nil, nil
//}

func (p MockChaincodeProvider) IsSysCC(name string) bool {
	return true
}

func (p MockChaincodeProvider) IsSysCCAndNotInvokableCC2CC(name string) bool {
	return true
}

func (p MockChaincodeProvider) IsSysCCAndNotInvokableExternal(name string) bool {
	return true
}
//
//func (p MockChaincodeProvider) GetApplicationConfig(cid string) (channelconfig.Application, bool) {
//	return nil, false
//}
//
//func (p MockChaincodeProvider) PolicyManager(channelID string) (policies.Manager, bool) {
//	return nil, false
//}

func (f MockFactory) NewSystemChaincodeProvider() SystemChaincodeProvider {
	return mockChaincodeProvider
}

func TestRegisterSystemChaincodeProviderFactory(t *testing.T) {
	factory := MockFactory{}

	RegisterSystemChaincodeProviderFactory(factory)
	assert.NotNil(t, sccFactory, "sccFactory should not be nil")
	assert.Equal(t, sccFactory, factory, "sccFactory should equal mock factory")
}

func TestGetSystemChaincodeProviderError(t *testing.T) {
	sccFactory = nil
	assert.Panics(t, func() { GetSystemChaincodeProvider() }, "Should panic because factory isnt set")
}

func TestGetSystemChaincodeProviderSuccess(t *testing.T) {
	sccFactory = MockFactory{}
	provider := GetSystemChaincodeProvider()
	assert.NotNil(t, provider, "provider should not be nil")
	assert.Equal(t, provider, mockChaincodeProvider, "provider equal mockChaincodeProvider")
}

func TestString(t *testing.T) {
	chaincodeInstance := ChaincodeInstance{
		ChainID:          "ChainID",
		ChaincodeName:    "ChaincodeName",
		ChaincodeVersion: "ChaincodeVersion",
	}

	assert.NotNil(t, chaincodeInstance.String(), "str should not be nil")
	assert.Equal(t, chaincodeInstance.String(), "ChainID.ChaincodeName#ChaincodeVersion", "str should be the correct value")

}
