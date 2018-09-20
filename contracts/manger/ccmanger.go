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
package manger

import (
	"time"
	"net"
	"os"
	"crypto/md5"
	"io"
	"fmt"
	"encoding/hex"
	"google.golang.org/grpc"
	"github.com/golang/protobuf/proto"

	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/accesscontrol"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	cfg "github.com/palletone/go-palletone/contracts/contractcfg"
)

func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

// CreateChaincodeProposalWithTxIDNonceAndTransient creates a proposal from given input
func createChaincodeProposalWithTxIDNonceAndTransient(txid string, typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, nonce, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// get a more appropriate mechanism to handle it in.
	var epoch uint64 = 0

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

	timestamp := util.CreateUtcTimestamp()
	hdr := &common.Header{ChannelHeader: marshalOrPanic(&common.ChannelHeader{
		Type:      int32(typ),
		TxId:      txid,
		Timestamp: timestamp,
		ChannelId: chainID,
		Extension: ccHdrExtBytes,
		Epoch:     epoch}),
		SignatureHeader: marshalOrPanic(&common.SignatureHeader{Nonce: nonce, Creator: creator})}

	hdrBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, "", err
	}

	return &peer.Proposal{Header: hdrBytes, Payload: ccPropPayloadBytes}, txid, nil
}

func computeProposalTxID(nonce, creator []byte) (string, error) {
	opdata := append(nonce, creator...)
	digest := util.ComputeSHA256(opdata)

	return hex.EncodeToString(digest), nil
}

func createChaincodeProposalWithTransient(typ common.HeaderType, chainID string, txid string, cis *peer.ChaincodeInvocationSpec, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, "", err
	}
	return createChaincodeProposalWithTxIDNonceAndTransient(txid, typ, chainID, cis, nonce, creator, transientMap)
}

func createChaincodeProposal(typ common.HeaderType, chainID string, txid string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return createChaincodeProposalWithTransient(typ, chainID, txid, cis, creator, nil)
}

func GetBytesProposal(prop *peer.Proposal) ([]byte, error) {
	propBytes, err := proto.Marshal(prop)
	return propBytes, err
}

func signedEndorserProposa(chainID string, txid string, cs *peer.ChaincodeSpec, creator, signature []byte) (*peer.SignedProposal, *peer.Proposal, error) {
	prop, _, err := createChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		chainID,
		txid,
		&peer.ChaincodeInvocationSpec{ChaincodeSpec: cs},
		creator)
	if err != nil {
		return nil, nil, err
	}

	propBytes, err := GetBytesProposal(prop)
	if err != nil {
		return nil, nil, err
	}

	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: signature}, prop, nil
}

func peerCreateChain(cid string) error {
	//chains.Lock()
	//defer chains.Unlock()

	//chains.list[cid] = &chain{
	//	//cs: &chainSupport{
	//	//},
	//}

	return nil
}

func peerServerInit() error {
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	//peerAddress := viper.GetString("peer.address")
	peerAddress := cfg.GetConfig().Address
	if peerAddress == "" {
		peerAddress = "0.0.0.0:21726"
	}

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		return err
	}
	ccStartupTimeout := time.Duration(30) * time.Second
	ca, _ := accesscontrol.NewCA()
	pb.RegisterChaincodeSupportServer(grpcServer, core.NewChaincodeSupport(peerAddress, false, ccStartupTimeout, ca))
	go grpcServer.Serve(lis)

	return nil
}

func peerServerDeInit() error {
	defer os.RemoveAll(cfg.GetConfig().ContractFileSystemPath)
	return nil
}

func systemContractInit() error {
	chainID := util.GetTestChainID()
	peerCreateChain(chainID)
	scc.RegisterSysCCs()
	scc.DeploySysCCs(chainID)
	return nil
}

func systemContractDeInit() error {
	chainID := util.GetTestChainID()
	scc.DeDeploySysCCs(chainID)
	return nil
}

func packChaincode(chainID string, ccName string, ccPath string, ccVersion string, args [][]byte) error {

	return nil
}

func recoverChaincodeFromeDb() error {

	return nil
}

func createDeployId(templateName string) string {
	t := time.Now()
	h := md5.New()
	io.WriteString(h, templateName)
	io.WriteString(h, t.String())
	id := fmt.Sprintf("%x", h.Sum(nil))
	return id
}




