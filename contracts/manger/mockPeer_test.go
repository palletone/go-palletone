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

package manger

import (
	"sync"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"encoding/hex"
	"fmt"
)


type chainSupport struct {
	//bundleSource *resourcesconfig.BundleSource
	//channelconfig.Resources
	//channelconfig.Application
	//ledger     ledger.PeerLedger
	//fileLedger *fileledger.FileLedger
}

// chain is a local struct to manage objects in a chain
type chain struct {
	cs        *chainSupport
	//cb        *common.Block
	//committer committer.Committer
}


// chains is a local map of chainID->chainObject
var chains = struct {
	sync.RWMutex
	list map[string]*chain
}{list: make(map[string]*chain)}


//MockInitialize resets chains for test env
func MockInitialize() {
	//ledgermgmt.InitializeTestEnvWithCustomProcessors(ConfigTxProcessors)
	chains.list = nil
	chains.list = make(map[string]*chain)
	//chainInitializer = func(string) { return }
}

func MockCreateChain(cid string) error {
	chains.Lock()
	defer chains.Unlock()

	chains.list[cid] = &chain{
		cs: &chainSupport{
			//Resources: &mockchannelconfig.Resources{
			//	PolicyManagerVal: &mockpolicies.Manager{
			//		Policy: &mockpolicies.Policy{},
			//	},
			//	ConfigtxValidatorVal: &mockconfigtx.Validator{},
			//},
			//ledger: ledger},
		},
	}

	return nil
}


func MarshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}
// CreateChaincodeProposalWithTxIDNonceAndTransient creates a proposal from given input
func CreateChaincodeProposalWithTxIDNonceAndTransient(txid string, typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, nonce, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	ccHdrExt := &peer.ChaincodeHeaderExtension{ChaincodeId: cis.ChaincodeSpec.ChaincodeId}
	ccHdrExtBytes, err := proto.Marshal(ccHdrExt)
	if err != nil {
		return nil, "", err
	}

	cisBytes, err := proto.Marshal(cis)
	if err != nil {
		return nil, "", err
	}

	ccPropPayload := &peer.ChaincodeProposalPayload{Input: cisBytes, TransientMap: transientMap}
	ccPropPayloadBytes, err := proto.Marshal(ccPropPayload)
	if err != nil {
		return nil, "", err
	}

	// TODO: epoch is now set to zero. This must be changed once we
	// get a more appropriate mechanism to handle it in.
	var epoch uint64 = 0

	timestamp := util.CreateUtcTimestamp()

	hdr := &common.Header{ChannelHeader: MarshalOrPanic(&common.ChannelHeader{
		Type:      int32(typ),
		TxId:      txid,
		Timestamp: timestamp,
		ChannelId: chainID,
		Extension: ccHdrExtBytes,
		Epoch:     epoch}),
		SignatureHeader: MarshalOrPanic(&common.SignatureHeader{Nonce: nonce, Creator: creator})}

	hdrBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, "", err
	}

	return &peer.Proposal{Header: hdrBytes, Payload: ccPropPayloadBytes}, txid, nil
}

// GetBytesProposal returns the bytes of a proposal message
func GetBytesProposal(prop *peer.Proposal) ([]byte, error) {
	propBytes, err := proto.Marshal(prop)
	return propBytes, err
}

// ComputeProposalTxID computes TxID as the Hash computed
// over the concatenation of nonce and creator.
func ComputeProposalTxID(nonce, creator []byte) (string, error) {
	// TODO: Get the Hash function to be used from
	// channel configuration
	//digest, err := factory.GetDefault().Hash(
	//	append(nonce, creator...),
	//	&bccsp.SHA256Opts{})
	opdata := append(nonce, creator...)
	digest := util.ComputeSHA256(opdata)

	//chaincodeLogger.Debugf("+++++++++++++++++ opdata:")
	//for _, i := range opdata{
	//	chaincodeLogger.Debugf("0x%2d ", i)
	//}
	//chaincodeLogger.Debugf("+++++++++++++++++ digest:")
	//for _, i := range digest{
	//	chaincodeLogger.Debugf("0x%2d ", i)
	//}

	return hex.EncodeToString(digest), nil
}
// CreateChaincodeProposalWithTransient creates a proposal from given input
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposalWithTransient(typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, "", err
	}

	// compute txid
	txid, err := ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, "", err
	}

	return CreateChaincodeProposalWithTxIDNonceAndTransient(txid, typ, chainID, cis, nonce, creator, transientMap)
}

// CreateChaincodeProposal creates a proposal from given input.
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposal(typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return CreateChaincodeProposalWithTransient(typ, chainID, cis, creator, nil)
}

// MockSignedEndorserProposalOrPanic creates a SignedProposal with the passed arguments
func MockSignedEndorserProposalOrPanic(chainID string, cs *peer.ChaincodeSpec, creator, signature []byte) (*peer.SignedProposal, *peer.Proposal) {
	prop, _, err := CreateChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		chainID,
		&peer.ChaincodeInvocationSpec{ChaincodeSpec: cs},
		creator)
	if err != nil {
		panic(err)
	}

	propBytes, err := GetBytesProposal(prop)
	if err != nil {
		panic(err)
	}

	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: signature}, prop
}

//glh
//func MockSignedEndorserProposal2OrPanic(chainID string, cs *peer.ChaincodeSpec, signer msp.SigningIdentity) (*peer.SignedProposal, *peer.Proposal) {
func MockSignedEndorserProposal2OrPanic(chainID string, cs *peer.ChaincodeSpec) (*peer.SignedProposal, *peer.Proposal) {
	//glh
	//serializedSigner, err := signer.Serialize()
	//if err != nil {
	//	panic(err)
	//}

	prop, _, err := CreateChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		chainID,
		&peer.ChaincodeInvocationSpec{ChaincodeSpec: &peer.ChaincodeSpec{}},
		nil)
		//glh
		//serializedSigner)

	if err != nil {
		panic(err)
	}

	//glh
	//sProp, err := GetSignedProposal(prop, signer)
	sProp, err := GetSignedProposal(prop)
	if err != nil {
		panic(err)
	}

	return sProp, prop
}


// GetSignedProposal returns a signed proposal given a Proposal message and a signing identity

//glh
// func GetSignedProposal(prop *peer.Proposal, signer msp.SigningIdentity) (*peer.SignedProposal, error) {
func GetSignedProposal(prop *peer.Proposal) (*peer.SignedProposal, error) {
	// check for nil argument
	//glh
	//if prop == nil || signer == nil {
	//	return nil, fmt.Errorf("Nil arguments")
	//}

	if prop == nil {
		return nil, fmt.Errorf("Nil arguments")
	}

	propBytes, err := GetBytesProposal(prop)
	if err != nil {
		return nil, err
	}

	//glh
	//signature, err := signer.Sign(propBytes)
	//if err != nil {
	//	return nil, err
	//}
	signature := make([]byte, len(propBytes))
	copy(signature, propBytes)


	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: signature}, nil
}
