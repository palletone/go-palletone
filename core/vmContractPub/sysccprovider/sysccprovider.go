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
	//"common/channelconfig"
	//"github.com/hyperledger/fabric/common/policies"
	//"github.com/hyperledger/fabric/core/ledger"
)

// SystemChaincodeProvider provides an abstraction layer that is
// used for different packages to interact with code in the
// system chaincode package without importing it; more methods
// should be added below if necessary
type SystemChaincodeProvider interface {
	// IsSysCC returns true if the supplied chaincode is a system chaincode
	IsSysCC(name string) bool

	// IsSysCCAndNotInvokableCC2CC returns true if the supplied chaincode
	// is a system chaincode and is not invokable through a cc2cc invocation
	IsSysCCAndNotInvokableCC2CC(name string) bool

	// IsSysCCAndNotInvokable returns true if the supplied chaincode
	// is a system chaincode and is not invokable through a proposal
	IsSysCCAndNotInvokableExternal(name string) bool

	// GetQueryExecutorForLedger returns a query executor for the
	// ledger of the supplied channel.
	// That's useful for system chaincodes that require unfettered
	// access to the ledger
	//glh
	//GetQueryExecutorForLedger(cid string) (ledger.QueryExecutor, error)

	// GetApplicationConfig returns the configtxapplication.SharedConfig for the channel
	// and whether the Application config exists
	//glh
	//GetApplicationConfig(cid string) (channelconfig.Application, bool)

	// Returns the policy manager associated to the passed channel
	// and whether the policy manager exists
	//glh
	//PolicyManager(channelID string) (policies.Manager, bool)
}

var sccFactory SystemChaincodeProviderFactory

// SystemChaincodeProviderFactory defines a factory interface so
// that the actual implementation can be injected
type SystemChaincodeProviderFactory interface {
	NewSystemChaincodeProvider() SystemChaincodeProvider
}

// RegisterSystemChaincodeProviderFactory is to be called once to set
// the factory that will be used to obtain instances of ChaincodeProvider
func RegisterSystemChaincodeProviderFactory(sccfact SystemChaincodeProviderFactory) {
	sccFactory = sccfact
}

// GetSystemChaincodeProvider returns instances of SystemChaincodeProvider;
// the actual implementation is controlled by the factory that
// is registered via RegisterSystemChaincodeProviderFactory
func GetSystemChaincodeProvider() SystemChaincodeProvider {
	if sccFactory == nil {
		panic("The factory must be set first via RegisterSystemChaincodeProviderFactory")
	}
	return sccFactory.NewSystemChaincodeProvider()
}

// ChaincodeInstance is unique identifier of chaincode instance
type ChaincodeInstance struct {
	ChainID          string
	ChaincodeName    string
	ChaincodeVersion string
}

func (ci *ChaincodeInstance) String() string {
	return ci.ChainID + "." + ci.ChaincodeName + "#" + ci.ChaincodeVersion
}
