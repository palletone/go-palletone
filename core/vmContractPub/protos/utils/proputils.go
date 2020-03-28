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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package utils

import (
	"errors"
	"fmt"

	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/contracts/platforms"

	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
)

// GetChaincodeInvocationSpec get the ChaincodeInvocationSpec from the proposal
func GetChaincodeInvocationSpec(prop *peer.PtnProposal) (*peer.PtnChaincodeInvocationSpec, error) {
	if prop == nil {
		return nil, fmt.Errorf("Proposal is nil")
	}
	_, err := GetHeader(prop.Header)
	if err != nil {
		return nil, err
	}
	ccPropPayload := &peer.PtnChaincodeProposalPayload{}
	err = proto.Unmarshal(prop.Payload, ccPropPayload)
	if err != nil {
		return nil, err
	}
	cis := &peer.PtnChaincodeInvocationSpec{}
	err = proto.Unmarshal(ccPropPayload.Input, cis)
	return cis, err
}

// GetChaincodeProposalContext returns creator and transient
func GetChaincodeProposalContext(prop *peer.PtnProposal) ([]byte, map[string][]byte, error) {
	if prop == nil {
		return nil, nil, fmt.Errorf("Proposal is nil")
	}
	if len(prop.Header) == 0 {
		return nil, nil, fmt.Errorf("Proposal's header is nil")
	}
	if len(prop.Payload) == 0 {
		return nil, nil, fmt.Errorf("Proposal's payload is nil")
	}

	//// get back the header
	hdr, err := GetHeader(prop.Header)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not extract the header from the proposal: %s", err)
	}
	if hdr == nil {
		return nil, nil, fmt.Errorf("Unmarshalled header is nil")
	}

	chdr, err := UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not extract the channel header from the proposal: %s", err)
	}

	if common.PtnHeaderType(chdr.Type) != common.PtnHeaderType_ENDORSER_TRANSACTION &&
		common.PtnHeaderType(chdr.Type) != common.PtnHeaderType_CONFIG {
		return nil, nil, fmt.Errorf("Invalid proposal type expected ENDORSER_TRANSACTION or CONFIG. Was: %d", chdr.Type)
	}

	shdr, err := GetSignatureHeader(hdr.SignatureHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not extract the signature header from the proposal: %s", err)
	}

	ccPropPayload := &peer.PtnChaincodeProposalPayload{}
	err = proto.Unmarshal(prop.Payload, ccPropPayload)
	if err != nil {
		return nil, nil, err
	}

	return shdr.Creator, ccPropPayload.TransientMap, nil
}

// GetHeader Get Header from bytes
func GetHeader(bytes []byte) (*common.PtnHeader, error) {
	hdr := &common.PtnHeader{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, err
}

// GetNonce returns the nonce used in Proposal
func GetNonce(prop *peer.PtnProposal) ([]byte, error) {
	if prop == nil {
		return nil, fmt.Errorf("Proposal is nil")
	}
	// get back the header
	hdr, err := GetHeader(prop.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not extract the header from the proposal: %s", err)
	}

	chdr, err := UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("Could not extract the channel header from the proposal: %s", err)
	}

	if common.PtnHeaderType(chdr.Type) != common.PtnHeaderType_ENDORSER_TRANSACTION &&
		common.PtnHeaderType(chdr.Type) != common.PtnHeaderType_CONFIG {
		return nil, fmt.Errorf("Invalid proposal type expected ENDORSER_TRANSACTION or CONFIG. Was: %d", chdr.Type)
	}

	shdr, err := GetSignatureHeader(hdr.SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("Could not extract the signature header from the proposal: %s", err)
	}

	if hdr.SignatureHeader == nil {
		return nil, errors.New("Invalid signature header. It must be different from nil.")
	}

	return shdr.Nonce, nil
}

// GetChaincodeHeaderExtension get chaincode header extension given header
func GetChaincodeHeaderExtension(hdr *common.PtnHeader) (*peer.PtnChaincodeHeaderExtension, error) {
	chdr, err := UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, err
	}

	chaincodeHdrExt := &peer.PtnChaincodeHeaderExtension{}
	err = proto.Unmarshal(chdr.Extension, chaincodeHdrExt)
	return chaincodeHdrExt, err
}

// GetProposalResponse given proposal in bytes
func GetProposalResponse(prBytes []byte) (*peer.PtnProposalResponse, error) {
	proposalResponse := &peer.PtnProposalResponse{}
	err := proto.Unmarshal(prBytes, proposalResponse)
	return proposalResponse, err
}

// GetChaincodeDeploymentSpec returns a ChaincodeDeploymentSpec given args
func GetChaincodeDeploymentSpec(code []byte) (*peer.PtnChaincodeDeploymentSpec, error) {
	cds := &peer.PtnChaincodeDeploymentSpec{}
	err := proto.Unmarshal(code, cds)
	if err != nil {
		return nil, err
	}

	// FAB-2122: Validate the CDS according to platform specific requirements
	platform, err := platforms.Find(cds.ChaincodeSpec.Type)
	if err != nil {
		return nil, err
	}

	err = platform.ValidateDeploymentSpec(cds)
	return cds, err
}

// GetChaincodeAction gets the ChaincodeAction given chaicnode action bytes
func GetChaincodeAction(caBytes []byte) (*peer.PtnChaincodeAction, error) {
	chaincodeAction := &peer.PtnChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, err
}

// GetResponse gets the Response given response bytes
func GetResponse(resBytes []byte) (*peer.PtnResponse, error) {
	response := &peer.PtnResponse{}
	err := proto.Unmarshal(resBytes, response)
	return response, err
}

// GetChaincodeEvents gets the ChaincodeEvents given chaincode event bytes
func GetChaincodeEvents(eBytes []byte) (*peer.PtnChaincodeEvent, error) {
	chaincodeEvent := &peer.PtnChaincodeEvent{}
	err := proto.Unmarshal(eBytes, chaincodeEvent)
	return chaincodeEvent, err
}

// GetProposalResponsePayload gets the proposal response payload
func GetProposalResponsePayload(prpBytes []byte) (*peer.PtnProposalResponsePayload, error) {
	prp := &peer.PtnProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, err
}

// GetProposal returns a Proposal message from its bytes
func GetProposal(propBytes []byte) (*peer.PtnProposal, error) {
	prop := &peer.PtnProposal{}
	err := proto.Unmarshal(propBytes, prop)
	return prop, err
}

// GetPayload Get Payload from Envelope message
func GetPayload(e *common.Envelope) (*common.PtnPayload, error) {
	payload := &common.PtnPayload{}
	err := proto.Unmarshal(e.Payload, payload)
	return payload, err
}

// GetChaincodeProposalPayload Get ChaincodeProposalPayload from bytes
func GetChaincodeProposalPayload(bytes []byte) (*peer.PtnChaincodeProposalPayload, error) {
	cpp := &peer.PtnChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, err
}

// GetSignatureHeader Get SignatureHeader from bytes
func GetSignatureHeader(bytes []byte) (*common.PtnSignatureHeader, error) {
	sh := &common.PtnSignatureHeader{}
	err := proto.Unmarshal(bytes, sh)
	return sh, err
}

// CreateChaincodeProposalWithTxIDNonceAndTransient creates a proposal from given input
func CreateChaincodeProposalWithTxIDNonceAndTransient(txid string, typ common.PtnHeaderType, chainID string, cis *peer.PtnChaincodeInvocationSpec, nonce, creator []byte, transientMap map[string][]byte) (*peer.PtnProposal, string, error) {
	ccHdrExt := &peer.PtnChaincodeHeaderExtension{ChaincodeId: cis.ChaincodeSpec.ChaincodeId}
	ccHdrExtBytes, err := proto.Marshal(ccHdrExt)
	if err != nil {
		return nil, "", err
	}

	cisBytes, err := proto.Marshal(cis)
	if err != nil {
		return nil, "", err
	}

	ccPropPayload := &peer.PtnChaincodeProposalPayload{Input: cisBytes, TransientMap: transientMap}
	ccPropPayloadBytes, err := proto.Marshal(ccPropPayload)
	if err != nil {
		return nil, "", err
	}

	// TODO: epoch is now set to zero. This must be changed once we
	// get a more appropriate mechanism to handle it in.
	var epoch uint64 = 0

	timestamp := util.CreateUtcTimestamp()

	hdr := &common.PtnHeader{ChannelHeader: MarshalOrPanic(&common.PtnChannelHeader{
		Type:      int32(typ),
		TxId:      txid,
		Timestamp: timestamp,
		ChannelId: chainID,
		Extension: ccHdrExtBytes,
		Epoch:     epoch}),
		SignatureHeader: MarshalOrPanic(&common.PtnSignatureHeader{Nonce: nonce, Creator: creator})}

	hdrBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, "", err
	}

	return &peer.PtnProposal{Header: hdrBytes, Payload: ccPropPayloadBytes}, txid, nil
}

// GetBytesProposalResponsePayload gets proposal response payload
func GetBytesProposalResponsePayload(hash []byte, response *peer.PtnResponse, result []byte, event []byte, ccid *peer.PtnChaincodeID) ([]byte, error) {
	cAct := &peer.PtnChaincodeAction{Events: event, Results: result, Response: response, ChaincodeId: ccid}
	cActBytes, err := proto.Marshal(cAct)
	if err != nil {
		return nil, err
	}

	prp := &peer.PtnProposalResponsePayload{Extension: cActBytes, ProposalHash: hash}
	prpBytes, err := proto.Marshal(prp)
	return prpBytes, err
}

// GetBytesChaincodeProposalPayload gets the chaincode proposal payload
func GetBytesChaincodeProposalPayload(cpp *peer.PtnChaincodeProposalPayload) ([]byte, error) {
	cppBytes, err := proto.Marshal(cpp)
	return cppBytes, err
}

// GetBytesResponse gets the bytes of Response
func GetBytesResponse(res *peer.PtnResponse) ([]byte, error) {
	resBytes, err := proto.Marshal(res)
	return resBytes, err
}

// GetBytesChaincodeEvent gets the bytes of ChaincodeEvent
func GetBytesChaincodeEvent(event *peer.PtnChaincodeEvent) ([]byte, error) {
	eventBytes, err := proto.Marshal(event)
	return eventBytes, err
}

//// GetBytesChaincodeActionPayload get the bytes of ChaincodeActionPayload from the message
//func GetBytesChaincodeActionPayload(cap *peer.ChaincodeActionPayload) ([]byte, error) {
//	capBytes, err := proto.Marshal(cap)
//	return capBytes, err
//}

// GetBytesProposalResponse gets proposal bytes response
func GetBytesProposalResponse(pr *peer.PtnProposalResponse) ([]byte, error) {
	respBytes, err := proto.Marshal(pr)
	return respBytes, err
}

// GetBytesProposal returns the bytes of a proposal message
func GetBytesProposal(prop *peer.PtnProposal) ([]byte, error) {
	propBytes, err := proto.Marshal(prop)
	return propBytes, err
}

// GetBytesHeader get the bytes of Header from the message
func GetBytesHeader(hdr *common.PtnHeader) ([]byte, error) {
	bytes, err := proto.Marshal(hdr)
	return bytes, err
}

// GetBytesSignatureHeader get the bytes of SignatureHeader from the message
func GetBytesSignatureHeader(hdr *common.PtnSignatureHeader) ([]byte, error) {
	bytes, err := proto.Marshal(hdr)
	return bytes, err
}

// CreateProposalFromCIS returns a proposal given a serialized identity and a ChaincodeInvocationSpec
func CreateProposalFromCISAndTxid(txid string, typ common.PtnHeaderType, chainID string, cis *peer.PtnChaincodeInvocationSpec, creator []byte) (*peer.PtnProposal, string, error) {
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, "", err
	}
	return CreateChaincodeProposalWithTxIDNonceAndTransient(txid, typ, chainID, cis, nonce, creator, nil)
}

// CreateChaincodeProposal creates a proposal from given input.
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposal(typ common.PtnHeaderType, chainID string, cis *peer.PtnChaincodeInvocationSpec, creator []byte) (*peer.PtnProposal, string, error) {
	return CreateChaincodeProposalWithTransient(typ, chainID, cis, creator, nil)
}

// CreateChaincodeProposalWithTransient creates a proposal from given input
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposalWithTransient(typ common.PtnHeaderType, chainID string, cis *peer.PtnChaincodeInvocationSpec, creator []byte, transientMap map[string][]byte) (*peer.PtnProposal, string, error) {
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

// ComputeProposalTxID computes TxID as the Hash computed
// over the concatenation of nonce and creator.
func ComputeProposalTxID(nonce, creator []byte) (string, error) {
	// channel configuration
	//glh
	//digest, err := factory.GetDefault().Hash(
	//	append(nonce, creator...),
	//	&bccsp.SHA256Opts{})
	//digest, err := util.ComputeSHA256(append(nonce, creator...)
	digest := util.ComputeSHA256(append(nonce, creator...))

	return hex.EncodeToString(digest), nil
}
