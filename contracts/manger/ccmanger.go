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
	"github.com/palletone/go-palletone/contracts/scc"
	"sync"
	"github.com/pkg/errors"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/common"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"encoding/hex"
	"github.com/spf13/viper"
	"time"
	"net"
	"github.com/palletone/go-palletone/contracts/core"
	"os"
	"google.golang.org/grpc"
	"github.com/palletone/go-palletone/contracts/accesscontrol"
	"golang.org/x/net/context"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

type chainSupport struct {
}

type chain struct {
	version   int
	cs        *chainSupport
}

var chains = struct {
	sync.RWMutex
	list map[string]*chain
}{list: make(map[string]*chain)}


type oldSysCCInfo struct {
	origSystemCC       []*scc.SystemChaincode
	origSysCCWhitelist map[string]string
}

func (osyscc *oldSysCCInfo) reset() {
	scc.MockResetSysCCs(osyscc.origSystemCC)
	viper.Set("chaincode.system", osyscc.origSysCCWhitelist)
}


//MockInitialize resets chains for test env
func init() {
	chains.list = nil
	chains.list = make(map[string]*chain)
}

func CreateChain(cid string, version int) error {
	chains.Lock()
	defer chains.Unlock()

	for k, v := range chains.list {
		if k == cid {
			logger.Errorf("chainId[%s] already exit, %v", cid, v)
			return errors.New("chainId already exit")
		}
	}

	chains.list[cid] = &chain{
		version: version,
		cs: &chainSupport{
		},
	}
	logger.Infof("creat chainId[%s] ok", cid)
	return nil
}

func InitCC() {
	initSysCCs(nil)
}

//start chaincodes
func initSysCCs(cids []string) {
	//deploy system chaincodes
	scc.DeploySysCCs("")

	//deploy multe chaincodes
	for	_, cid := range cids{
		if len(cid) > 0{
			scc.DeploySysCCs(cid)
		}
	}
	logger.Infof("Deployed system chaincodes")
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

func ComputeProposalTxID(nonce, creator []byte) (string, error) {
	opdata := append(nonce, creator...)
	digest := util.ComputeSHA256(opdata)

	return hex.EncodeToString(digest), nil
}

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

func CreateChaincodeProposal(typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return CreateChaincodeProposalWithTransient(typ, chainID, cis, creator, nil)
}

func GetBytesProposal(prop *peer.Proposal) ([]byte, error) {
	propBytes, err := proto.Marshal(prop)
	return propBytes, err
}

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

func peerCreateChain(cid string) error {
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

///////

func Init() error {

	PeerServerInit()

	SystemContractInit()

	return nil
}

func Deinit() {
}


func ContractInvoke(chainID string, ccName string,  args [][]byte) error{
	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone")
	ccVersion := "ptn001"

	es := NewEndorserServer(mksupt)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: ccName},
		Type:        pb.ChaincodeSpec_GOLANG,
		Input:       &pb.ChaincodeInput{Args: args},
	}

	cid := &pb.ChaincodeID {
		Path:     "", //no use
		Name:     ccName,
		Version:  ccVersion,
	}

	sprop, prop := MockSignedEndorserProposalOrPanic(chainID, spec, creator, []byte("msg1"))
	rsp, err := es.ProcessProposal(context.Background(), sprop, prop, chainID, "txid001", cid)
	if err != nil {
		logger.Errorf("ProcessProposal error[%v]", err)
	}
	logger.Infof("ProcessProposal rsp=%v", rsp)

	return nil
}


func PeerServerInit() error {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	viper.Set("peer.fileSystemPath", "/home/glh/tmp/chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	peerAddress := "0.0.0.0:21726"
	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		return nil
	}

	ccStartupTimeout := time.Duration(5000) * time.Millisecond
	ca, _ := accesscontrol.NewCA()
	pb.RegisterChaincodeSupportServer(grpcServer, core.NewChaincodeSupport(peerAddress, false, ccStartupTimeout, ca))

	go grpcServer.Serve(lis)

	return nil
}

func PeerDeInit() error{
	defer os.RemoveAll("/home/glh/tmp/chaincodes")
	return nil
}

func SystemContractInit() error {
	chainID := util.GetTestChainID()
	peerCreateChain(chainID)

	scc.RegisterSysCCs()
	scc.DeploySysCCs(chainID)

	return nil
}

func SystemContractDeInit(cid string) error {
	scc.DeDeploySysCCs(cid)
	return nil
}

