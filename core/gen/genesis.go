// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package gen

import (
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"

	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	dagCommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

const deFaultNode = "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f" +
	"59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *configure.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisUnit(genesis *core.Genesis, ks *keystore.KeyStore, account accounts.Account) (*modules.Unit, error) {

	//var unitRep dagCommon.IUnitRepository
	//unitRep = dagCommon.NewUnitRepository4Db(db)
	//genesisUnit, err := dag.GetGenesisUnit(0)
	//if err != nil && err.Error() != errors.ErrNotFound.Error() {
	//	log.Info("get genesis error", "error", err)
	//	return nil, err
	//}
	// check genesis unit existing
	//if genesisUnit != nil {
	//	return nil, fmt.Errorf("Genesis unit(%s) has been created.", genesisUnit.UnitHash.String())
	//}
	unit, err := setupGenesisUnit(genesis, ks)
	if err != nil {
		return unit, err
	}

	// modify by Albert·Gou
	unit, err = dagCommon.GetUnitWithSig(unit, ks, account.Address)
	if err != nil {
		return unit, err
	}

	// to save unit in db
	//if err := CommitDB(dag, unit, true); err != nil {
	//	log.Error("Commit genesis unit to db:", "error", err.Error())
	//	return unit, err
	//}
	return unit, nil
}

func setupGenesisUnit(genesis *core.Genesis, ks *keystore.KeyStore) (*modules.Unit, error) {

	if genesis == nil {
		log.Info("Writing default main-net genesis block")
		genesis = DefaultGenesisBlock()
	} else {
		log.Info("Writing custom genesis block")
	}
	txs, asset := GetGensisTransctions(ks, genesis)
	log.Info("-> Genesis transactions:")
	for i, tx := range txs {
		msg := fmt.Sprintf("Tx[%d]: %s\n", i, tx.Hash().String())
		log.Info(msg)
	}
	//return modules.NewGenesisUnit(genesis, txs)
	return dagCommon.NewGenesisUnit(txs, genesis.InitialTimestamp, asset, genesis.ParentUnitHeight, genesis.ParentUnitHash)
}

func GetGensisTransctions(ks *keystore.KeyStore, genesis *core.Genesis) (modules.Transactions, *modules.Asset) {
	// step1, generate payment payload message: coin creation
	holder, err := common.StringToAddress(genesis.TokenHolder)

	if err != nil || holder.GetType() != common.PublicKeyHash {
		log.Error("Genesis holder address is an invalid p2pkh address.")
		return nil, nil
	}

	//assetInfo := modules.FungibleToken{
	//	Name:        genesis.GasToken,
	//	TotalSupply: genesis.GetTokenAmount(),
	//	Decimals:    byte(genesis.TokenDecimal),
	//	Symbol:      genesis.DecimalUnit,
	//	//SupplyAddress: holder,
	//}
	// get new asset id
	//asset := modules.NewPTNAsset()
	//err = err
	//asset := &modules.Asset{
	//	AssetId: assetId,
	//}
	//assetInfo.AssetID = asset
	//extra, err := rlp.EncodeToBytes(assetInfo)
	//if err != nil {
	//	log.Error("Get genesis assetinfo bytes error.")
	//	return nil, nil
	//}
	//txin := &modules.Input{
	//	Extra: extra, // save asset info
	//}
	// generate p2pkh bytes
	addr, _ := common.StringToAddress(holder.String())
	pkscript := tokenengine.Instance.GenerateP2PKHLockScript(addr.Bytes())
	asset, _ := modules.StringToAsset(genesis.GasToken)
	txout := &modules.Output{
		Value: genesis.GetTokenAmount(),
		//Value:    genesis.TokenAmount,
		Asset:    asset,
		PkScript: pkscript,
	}
	pay := &modules.PaymentPayload{
		//Inputs:  []*modules.Input{txin},
		Outputs: []*modules.Output{txout},
	}
	msg0 := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay,
	}
	// step2 generate text payload
	msg1 := &modules.Message{
		App:     modules.APP_DATA,
		Payload: &modules.DataPayload{MainData: []byte("Genesis Text"), ExtraData: []byte(genesis.Text)},
	}
	tx := &modules.Transaction{
		TxMessages: []*modules.Message{msg0, msg1},
	}
	// step3, generate global config payload message
	configPayloads, err := dagCommon.GenGenesisConfigPayload(genesis, asset)
	if err != nil {
		log.Error("Generate genesis unit config payload error.")
		return nil, nil
	}

	for _, payload := range configPayloads {
		newMsg := &modules.Message{
			App:     modules.APP_CONTRACT_INVOKE,
			Payload: payload,
		}
		tx.TxMessages = append(tx.TxMessages, newMsg)
	}
	//Init system contract
	for _, sc := range genesis.SystemContracts {
		newMsg := &modules.Message{
			App:     modules.APP_CONTRACT_DEPLOY,
			Payload: &modules.ContractDeployPayload{ContractId: sc.Address.Bytes(), Name: sc.Name},
		}
		if sc.Active {
			tx.TxMessages = append(tx.TxMessages, newMsg)
		}
	}
	// step4, generate initial mediator info payload
	//initialMediatorMsgs := dagCommon.GetInitialMediatorMsgs(genesis)
	//tx.TxMessages = append(tx.TxMessages, initialMediatorMsgs...)

	// tx.CreationDate = tx.CreateDate()
	//tx.TxHash = tx.Hash()

	txs := []*modules.Transaction{tx}
	return txs, asset
}

//func sigData(key *ecdsa.PrivateKey, data interface{}) ([]byte, error) {
//	txBytes, _ := rlp.EncodeToBytes(data)
//	hash := crypto.Keccak256(txBytes)
//	sign, err := crypto.Sign(hash, key)
//
//	return sign, err
//}

func GenContractTransction(orgTx *modules.Transaction, msgs []*modules.Message) (*modules.Transaction, error) {
	if orgTx == nil || len(orgTx.TxMessages) < 2 {
		return nil, errors.New(fmt.Sprintf("GenContractTransction param is error"))
	}
	tx := &modules.Transaction{}
	for i := 0; i < len(orgTx.TxMessages); i++ {
		tx.AddMessage(orgTx.TxMessages[i])
	}
	for i := 0; i < len(msgs); i++ {
		tx.AddMessage(msgs[i])
	}
	tx.CertId = orgTx.CertId
	tx.Illegal = orgTx.Illegal
	return tx, nil
}

//func CommitDB(dag dag.IDag, unit *modules.Unit, isGenesis bool) error {
//	// save genesis unit to leveldb
//	if err := dag.SaveUnit(unit, isGenesis); err != nil {
//		return err
//	} else {
//		log.Info("Save genesis unit success.")
//	}
//
//	return nil
//}

// DefaultGenesisBlock returns the PalletOne main net genesis block.
func DefaultGenesisBlock() *core.Genesis {
	//SystemConfig := core.SystemConfig{
	//	DepositRate:               core.DefaultDepositRate,
	//	FoundationAddress:         core.DefaultFoundationAddress,
	//	DepositAmountForMediator:  core.DefaultDepositAmountForMediator,
	//	DepositAmountForJury:      core.DefaultDepositAmountForJury,
	//	DepositAmountForDeveloper: core.DefaultDepositAmountForDeveloper,
	//	DepositPeriod:             core.DefaultDepositPeriod,
	//	UccMemory:                 core.DefaultUccMemory,
	//	UccMemorySwap:             core.DefaultUccMemorySwap,
	//	UccCpuShares:              core.DefaultUccCpuShares,
	//	UccCpuPeriod:              core.DefaultCpuPeriod,
	//	UccCpuQuota:               core.DefaultUccCpuQuota,
	//	UccCpuSetCpus:             core.DefaultUccCpuSetCpus,
	//	TempUccMemory:             core.DefaultTempUccMemory,
	//	TempUccMemorySwap:         core.DefaultTempUccMemorySwap,
	//	TempUccCpuShares:          core.DefaultTempUccCpuShares,
	//	TempUccCpuQuota:           core.DefaultTempUccCpuQuota,
	//	ContractSignatureNum:      core.DefaultContractSignatureNum,
	//	ContractElectionNum:       core.DefaultContractElectionNum,
	//	ActiveMediatorCount:       strconv.FormatUint(core.DefaultMediatorCount, 10),
	//}

	//DigitalIdentityConfig := core.DigitalIdentityConfig{
	//	// default root ca holder, 默认是基金会地址
	//	RootCAHolder: core.DefaultFoundationAddress,
	//	RootCABytes:  core.DefaultRootCABytes,
	//}

	initParams := core.NewChainParams()

	return &core.Genesis{
		Version:     configure.Version,
		TokenAmount: core.DefaultTokenAmount,
		//TokenDecimal:           core.DefaultTokenDecimal,
		ParentUnitHeight: -1,
		ChainID:          1,
		TokenHolder:      core.DefaultTokenHolder,
		//SystemConfig:          SystemConfig,
		DigitalIdentityConfig: core.DefaultDigitalIdentityConfig(),
		InitialParameters:     initParams,
		ImmutableParameters:   core.NewImmutChainParams(),
		InitialTimestamp:      InitialTimestamp(initParams.MediatorInterval),
		//InitialActiveMediators: core.DefaultMediatorCount,
		InitialMediatorCandidates: InitialMediatorCandidates(core.DefaultActiveMediatorCount,
			core.DefaultMediator),
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *core.Genesis {
	//SystemConfig := core.SystemConfig{
	//	DepositRate:               core.DefaultDepositRate,
	//	FoundationAddress:         core.DefaultFoundationAddress,
	//	DepositAmountForJury:      core.DefaultDepositAmountForJury,
	//	DepositAmountForMediator:  core.DefaultDepositAmountForMediator,
	//	DepositAmountForDeveloper: core.DefaultDepositAmountForDeveloper,
	//	DepositPeriod:             core.DefaultDepositPeriod,
	//	UccMemory:                 core.DefaultUccMemory,
	//	UccMemorySwap:             core.DefaultUccMemorySwap,
	//	UccCpuShares:              core.DefaultUccCpuShares,
	//	UccCpuPeriod:              core.DefaultCpuPeriod,
	//	UccCpuQuota:               core.DefaultUccCpuQuota,
	//	UccCpuSetCpus:             core.DefaultUccCpuSetCpus,
	//	TempUccMemory:             core.DefaultTempUccMemory,
	//	TempUccMemorySwap:         core.DefaultTempUccMemorySwap,
	//	TempUccCpuShares:          core.DefaultTempUccCpuShares,
	//	TempUccCpuQuota:           core.DefaultTempUccCpuQuota,
	//	ContractSignatureNum:      core.DefaultContractSignatureNum,
	//	ContractElectionNum:       core.DefaultContractElectionNum,
	//	ActiveMediatorCount:       strconv.FormatUint(core.DefaultMediatorCount, 10),
	//}

	//DigitalIdentityConfig := core.DigitalIdentityConfig{
	//	// default root ca holder, 默认是基金会地址
	//	RootCAHolder: core.DefaultFoundationAddress,
	//	RootCABytes:  core.DefaultRootCABytes,
	//}

	initParams := core.NewChainParams()

	return &core.Genesis{
		Version:     configure.Version,
		TokenAmount: core.DefaultTokenAmount,
		//TokenDecimal:           core.DefaultTokenDecimal,
		ParentUnitHeight: -1,
		ChainID:          1,
		TokenHolder:      core.DefaultTokenHolder,
		//SystemConfig:          SystemConfig,
		DigitalIdentityConfig: core.DefaultDigitalIdentityConfig(),
		InitialParameters:     initParams,
		ImmutableParameters:   core.NewImmutChainParams(),
		InitialTimestamp:      InitialTimestamp(initParams.MediatorInterval),
		//InitialActiveMediators: core.DefaultMediatorCount,
		InitialMediatorCandidates: InitialMediatorCandidates(core.DefaultActiveMediatorCount,
			core.DefaultMediator),
	}
}

func InitialMediatorCandidates(len int, address string) []*core.InitialMediator {
	initialMediator := make([]*core.InitialMediator, len)
	for i := 0; i < len; i++ {
		im := core.NewInitialMediator()
		im.AddStr = address
		im.RewardAdd = address
		im.InitPubKey = core.DefaultInitPubKey
		im.Node = deFaultNode
		initialMediator[i] = im
	}

	return initialMediator
}

func InitialTimestamp(mediatorInterval uint8) int64 {
	mi := int64(mediatorInterval)
	return time.Now().Unix() / mi * mi
}
