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

package scc

import (
	"github.com/palletone/go-palletone/contracts/example/go/samplesyscc"
	"github.com/palletone/go-palletone/contracts/example/go/samplesyscc1"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/contracts/syscontract/blacklistcc"
	"github.com/palletone/go-palletone/contracts/syscontract/coinbasecc"
	"github.com/palletone/go-palletone/contracts/syscontract/debugcc"
	depositv1 "github.com/palletone/go-palletone/contracts/syscontract/deposit/v1"
	depositv2 "github.com/palletone/go-palletone/contracts/syscontract/deposit/v2"
	"github.com/palletone/go-palletone/contracts/syscontract/digitalidcc"
	"github.com/palletone/go-palletone/contracts/syscontract/exchangecc"
	"github.com/palletone/go-palletone/contracts/syscontract/installcc"
	"github.com/palletone/go-palletone/contracts/syscontract/packetcc"
	"github.com/palletone/go-palletone/contracts/syscontract/partitioncc"
	prc20v1 "github.com/palletone/go-palletone/contracts/syscontract/prc20/v1"
	prc20v2 "github.com/palletone/go-palletone/contracts/syscontract/prc20/v2"

	prc721v1 "github.com/palletone/go-palletone/contracts/syscontract/prc721/v1"
	prc721v2 "github.com/palletone/go-palletone/contracts/syscontract/prc721/v2"
	"github.com/palletone/go-palletone/contracts/syscontract/sysconfigcc"
	votev1 "github.com/palletone/go-palletone/contracts/syscontract/vote/v1"
	votev2 "github.com/palletone/go-palletone/contracts/syscontract/vote/v2"
	"github.com/palletone/go-palletone/contracts/syscontract/auctioncc"
)

var systemChaincodes = []*SystemChaincode{
	{
		Id:        syscontract.TestRunContractAddress.Bytes(), //PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX
		Enabled:   true,
		Name:      "sample_syscc",
		Path:      "../example/go/samplesyscc/samplesyscc",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &samplesyscc.SampleSysCC{},
	},
	{
		Id:        syscontract.TestRunContractAddress.Bytes(), //PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX
		Enabled:   true,
		Name:      "sample_syscc",
		Path:      "../example/go/samplesyscc1/samplesyscc1",
		Version:   "ptn002",
		InitArgs:  [][]byte{},
		Chaincode: &samplesyscc1.SampleSysCC{},
	},
	{
		Id:        syscontract.DepositContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "deposit_syscc",
		Path:      "../syscontract/deposit/v1",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &depositv1.DepositChaincode{},
	},
	{
		Id:        syscontract.DepositContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "deposit_syscc",
		Path:      "../syscontract/deposit/v2",
		Version:   "ptn002",
		InitArgs:  [][]byte{},
		Chaincode: &depositv2.DepositChaincode{},
	},
	{
		Id:        syscontract.CreateTokenContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "createToken_sycc",
		Path:      "../syscontract/prc20/prc20_1",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &prc20v1.PRC20{},
	},
	{
		Id:        syscontract.CreateTokenContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "createToken_sycc",
		Path:      "../syscontract/prc20/prc20_2",
		Version:   "ptn002",
		InitArgs:  [][]byte{},
		Chaincode: &prc20v2.PRC20{},
	},
	{
		Id:        syscontract.VoteTokenContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "voteToken_sycc",
		Path:      "../syscontract/vote/vote",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &votev1.Vote{},
	},
	{
		Id:        syscontract.VoteTokenContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "voteToken_sycc",
		Path:      "../syscontract/vote/voteV2",
		Version:   "ptn002",
		InitArgs:  [][]byte{},
		Chaincode: &votev2.Vote{},
	},
	{
		Id:        syscontract.SysConfigContractAddress.Bytes(),
		Enabled:   true,
		Name:      "sysconfig_sycc",
		Path:      "../syscontract/sysconfigcc/sysconfigcc",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &sysconfigcc.SysConfigChainCode{},
	},
	{
		Id:        syscontract.CreateToken721ContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "createToken721_sycc",
		Path:      "../syscontract/prc721/prc721",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &prc721v1.PRC721{},
	},
	{
		Id:        syscontract.CreateToken721ContractAddress.Bytes(), //合约ID为20字节
		Enabled:   true,
		Name:      "createToken721_sycc",
		Path:      "../syscontract/prc721/prc721_v2",
		Version:   "ptn002",
		InitArgs:  [][]byte{},
		Chaincode: &prc721v2.PRC721{},
	},
	{
		Id:        syscontract.DigitalIdentityContractAddress.Bytes(),
		Enabled:   true,
		Name:      "digital_identity_sycc",
		Path:      "../syscontract/digitalidcc/digitalidcc",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &digitalidcc.DigitalIdentityChainCode{},
	},
	{
		Id:        syscontract.PartitionContractAddress.Bytes(),
		Enabled:   true,
		Name:      "partition_manager_sycc",
		Path:      ".",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &partitioncc.PartitionMgr{},
	},
	{
		Id:        syscontract.TestContractAddress.Bytes(),
		Enabled:   true,
		Name:      "debug_sycc",
		Path:      "../syscontract/debugcc/debugcc",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &debugcc.DebugChainCode{},
	},
	{
		Id:        syscontract.CoinbaseContractAddress.Bytes(),
		Enabled:   true,
		Name:      "coinbase_sycc",
		Path:      "",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &coinbasecc.CoinbaseChainCode{},
	},
	{
		Id:        syscontract.BlacklistContractAddress.Bytes(),
		Enabled:   true,
		Name:      "blacklist_sycc",
		Path:      "./BlacklistContractAddress",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &blacklistcc.BlacklistMgr{},
	},
	{
		Id:        syscontract.ExchangeContractAddress.Bytes(),
		Enabled:   true,
		Name:      "exchange_sycc",
		Path:      "./ExchangeContractAddress",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &exchangecc.ExchangeMgr{},
	},
	{
		Id:        syscontract.PacketContractAddress.Bytes(),
		Enabled:   true,
		Name:      "packet_sycc",
		Path:      "./Packet",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &packetcc.PacketMgr{},
	},
	{
		Id:        syscontract.InstallContractAddress.Bytes(),
		Enabled:   true,
		Name:      "install_sycc",
		Path:      "./installcc",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &installcc.InstallMgr{},
	},
	{
		Id:        syscontract.AuctionContractAddress.Bytes(),
		Enabled:   true,
		Name:      "auction_sycc",
		Path:      "./auctioncc",
		Version:   "ptn001",
		InitArgs:  [][]byte{},
		Chaincode: &auctioncc.AuctionMgr{},
	},
	//TODO add other system chaincodes ...
}

//DeploySysCCs is the hook for system chaincodes where system chaincodes are registered
//note the chaincode must still be deployed and launched like a user chaincode will be
func DeploySysCCs(chainID string) {
	for _, sysCC := range systemChaincodes {
		deploySysCC(chainID, sysCC)
	}
}

//DeDeploySysCCs is used in unit tests to stop and remove the system chaincodes before
//restarting them in the same process. This allows clean start of the system
//in the same process
func DeDeploySysCCs(chainID string) {
	for _, sysCC := range systemChaincodes {
		DeDeploySysCC(chainID, sysCC, true)
	}
}

//IsSysCC returns true if the name matches a system chaincode's
//system chaincode names are system, chain wide
func IsSysCC(name string) bool {
	for _, sysCC := range systemChaincodes {
		if sysCC.Name == name {
			return true
		}
	}
	return false
}

// IsSysCCAndNotInvokableExternal returns true if the chaincode
// is a system chaincode and *CANNOT* be invoked through
// a proposal to this peer
func IsSysCCAndNotInvokableExternal(name string) bool {
	for _, sysCC := range systemChaincodes {
		if sysCC.Name == name {
			return !sysCC.InvokableExternal
		}
	}
	return false
}

// IsSysCCAndNotInvokableCC2CC returns true if the chaincode
// is a system chaincode and *CANNOT* be invoked through
// a cc2cc invocation
func IsSysCCAndNotInvokableCC2CC(name string) bool {
	for _, sysCC := range systemChaincodes {
		if sysCC.Name == name {
			return !sysCC.InvokableCC2CC
		}
	}
	return false
}

// MockRegisterSysCCs is used only for testing
// This is needed to break import cycle
func MockRegisterSysCCs(mockSysCCs []*SystemChaincode) []*SystemChaincode {
	orig := systemChaincodes
	systemChaincodes = mockSysCCs

	RegisterSysCCs()
	return orig
}

// MockResetSysCCs restore orig system ccs - is used only for testing
func MockResetSysCCs(mockSysCCs []*SystemChaincode) {
	systemChaincodes = mockSysCCs
}
