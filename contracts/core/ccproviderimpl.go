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

package core

import (
	"context"

	"errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/vmContractPub/ccprovider"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"time"
)

// ccProviderFactory implements the ccprovider.ChaincodeProviderFactory
// interface and returns instances of ccprovider.ChaincodeProvider
type ccProviderFactory struct {
}

// ccProviderImpl is an implementation of the ccprovider.ChaincodeProvider interface
type ccProviderImpl struct {
}

// NewChaincodeProvider returns pointers to ccProviderImpl as an
// implementer of the ccprovider.ChaincodeProvider interface
func (c *ccProviderFactory) NewChaincodeProvider() ccprovider.ChaincodeProvider {
	return &ccProviderImpl{}
}

// init is called when this package is loaded. This implementation registers the factory
func init() {
	ccprovider.RegisterChaincodeProviderFactory(&ccProviderFactory{})
}

// ccProviderContextImpl contains the state that is passed around to calls to methods of ccProviderImpl
type ccProviderContextImpl struct {
	ctx *ccprovider.CCContext
}

//glh
// GetContext returns a context for the supplied ledger, with the appropriate tx simulator
//
//func (c *ccProviderImpl) GetContext(ledger ledger.PeerLedger, txid string) (context.Context, ledger.TxSimulator, error) {
//	var err error
//	// get context for the chaincode execution
//	txsim, err := ledger.NewTxSimulator(txid)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	ctxt := context.WithValue(context.Background(), TXSimulatorKey, txsim)
//	return ctxt, txsim, nil
//}

func (c *ccProviderImpl) GetContext() (context.Context, error) {
	return nil, nil
}

// GetCCContext returns an interface that encapsulates a
// chaincode context; the interface is required to avoid
// referencing the chaincode package from the interface definition
func (c *ccProviderImpl) GetCCContext(contractid []byte, cid, name, version, txid string, syscc bool, signedProp *pb.SignedProposal, prop *pb.Proposal) interface{} {
	ctx := ccprovider.NewCCContext(contractid, cid, name, version, txid, syscc, signedProp, prop)
	return &ccProviderContextImpl{ctx: ctx}
}

// ExecuteChaincode executes the chaincode specified in the context with the specified arguments
func (c *ccProviderImpl) ExecuteChaincode(ctxt context.Context, cccid interface{}, args [][]byte, timeout time.Duration) (*pb.Response, *pb.ChaincodeEvent, error) {
	return ExecuteChaincode(ctxt, cccid.(*ccProviderContextImpl).ctx, args, timeout)
}

// Execute executes the chaincode given context and spec (invocation or deploy)
func (c *ccProviderImpl) Execute(ctxt context.Context, cccid interface{}, spec interface{}, timeout time.Duration) (*pb.Response, *pb.ChaincodeEvent, error) {
	return Execute(ctxt, cccid.(*ccProviderContextImpl).ctx, spec, timeout)
}

// ExecuteWithErrorFilter executes the chaincode given context and spec and returns payload
func (c *ccProviderImpl) ExecuteWithErrorFilter(ctxt context.Context, cccid interface{}, spec interface{}, timeout time.Duration) ([]byte, *pb.ChaincodeEvent, error) {
	return ExecuteWithErrorFilter(ctxt, cccid.(*ccProviderContextImpl).ctx, spec, timeout)
}

// Stop stops the chaincode given context and spec
func (c *ccProviderImpl) Stop(ctxt context.Context, cccid interface{}, spec *pb.ChaincodeDeploymentSpec, dontRmCon bool) error {
	if theChaincodeSupport != nil {
		log.Debugf("theChainode support is not nil.")
		return theChaincodeSupport.Stop(ctxt, cccid.(*ccProviderContextImpl).ctx, spec, dontRmCon)
	}
	return errors.New("Stop:ChaincodeSupport not initialized")
}

func (c *ccProviderImpl) Destroy(ctxt context.Context, cccid interface{}, spec *pb.ChaincodeDeploymentSpec) error {
	if theChaincodeSupport != nil {
		return theChaincodeSupport.Destroy(ctxt, cccid.(*ccProviderContextImpl).ctx, spec)
	}
	return errors.New("Destroy:ChaincodeSupport not initialized")
}
