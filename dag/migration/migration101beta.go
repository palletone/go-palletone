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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"strconv"

	"github.com/palletone/go-palletone/tokenengine"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration100_101 struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration100_101) FromVersion() string {
	return "1.0.0-beta"
}

func (m *Migration100_101) ToVersion() string {
	return "1.0.1-beta"
}

func (m *Migration100_101) utxoToStxo() error {
	//删除已经花费的UTXO
	dbop := storage.NewUtxoDb(m.utxodb, tokenengine.Instance)
	utxos, err := dbop.GetAllUtxos()
	if err != nil {
		return err
	}
	for outpoint, utxo := range utxos {
		outpoint := outpoint
		if utxo.IsSpent() {
			err = dbop.DeleteUtxo(&outpoint, common.Hash{}, 0)
			if err != nil {
				log.Errorf("Migrate utxo db,delete spent utxo error:%s", err.Error())
				return err
			}
		}
	}
	dagdb := storage.NewDagDb(m.dagdb)
	iter := m.dagdb.NewIteratorWithPrefix(constants.TRANSACTION_PREFIX)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		tx := new(modules.Transaction)
		err := rlp.DecodeBytes(value, tx)
		if err != nil {
			log.Errorf("Cannot decode key[%s] rlp tx:%x", key, value)
			continue
		}

		spents := tx.GetSpendOutpoints()
		for _, spent := range spents {
			stxo, err := dbop.GetStxoEntry(spent)
			if err == nil && stxo != nil {
				stxo.SpentByTxId = tx.Hash()
				lookup, _ := dagdb.GetTxLookupEntry(tx.Hash())
				stxo.SpentTime = lookup.Timestamp
				log.Debugf("Update stxo spentTxId:%s,spentTime:%d", stxo.SpentByTxId.String(), stxo.SpentTime)
				dbop.SaveStxoEntry(spent, stxo)
			}
		}
	}

	// txs, err := dagdb.GetAllTxs()
	// if err != nil {
	// 	log.Error(err.Error())
	// }
	// log.Debugf("Tx count:%d", len(txs))
	// for i, tx := range txs {
	// 	if tx == nil {
	// 		log.Errorf("tx[%d] is nil", i)
	// 	}
	// 	spents := tx.GetSpendOutpoints()
	// 	for _, spent := range spents {
	// 		stxo, err := dbop.GetStxoEntry(spent)
	// 		if err == nil && stxo != nil {
	// 			stxo.SpentByTxId = tx.Hash()
	// 			lookup, _ := dagdb.GetTxLookupEntry(tx.Hash())
	// 			stxo.SpentTime = lookup.Timestamp
	// 			log.Debugf("Update stxo spentTxId:%s,spentTime:%d", stxo.SpentByTxId.String(), stxo.SpentTime)
	// 			dbop.SaveStxoEntry(spent, stxo)
	// 		}
	// 	}
	// }

	return nil
}

func (m *Migration100_101) ExecuteUpgrade() error {
	// utxo迁移成Stxo
	if err := m.utxoToStxo(); err != nil {
		return err
	}

	// 转换mediator结构体
	if err := m.upgradeMediatorInfo(); err != nil {
		return err
	}

	//转换GLOBALPROPERTY结构体
	if err := m.upgradeGP(); err != nil {
		return err
	}

	return nil
}

func (m *Migration100_101) upgradeMediatorInfo() error {
	oldMediatorsIterator := m.statedb.NewIteratorWithPrefix(constants.MEDIATOR_INFO_PREFIX)
	for oldMediatorsIterator.Next() {
		oldMediator := &MediatorInfo100{}
		err := rlp.DecodeBytes(oldMediatorsIterator.Value(), oldMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		newMediator := &MediatorInfo101{
			MediatorInfoBase101: oldMediator.MediatorInfoBase101,
			MediatorApplyInfo:   &core.MediatorApplyInfo{Description: oldMediator.ApplyInfo},
			MediatorInfoExpand:  oldMediator.MediatorInfoExpand,
		}

		err = storage.StoreToRlpBytes(m.statedb, oldMediatorsIterator.Key(), newMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}
	}

	return nil
}

type MediatorInfo100 struct {
	*MediatorInfoBase101
	*MediatorApplyInfo100
	*core.MediatorInfoExpand
}

type MediatorApplyInfo100 struct {
	ApplyInfo string `json:"applyInfo"` //  申请信息
}

func (m *Migration100_101) upgradeGP() error {
	oldGp := &GlobalProperty100{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &GlobalProperty101{}
	newData.ActiveJuries = oldGp.ActiveJuries
	newData.ActiveMediators = oldGp.ActiveMediators
	newData.PrecedingMediators = oldGp.PrecedingMediators
	newData.ImmutableParameters = oldGp.ImmutableParameters
	newData.ChainParameters.ChainParametersBase102delta = oldGp.ChainParameters.ChainParametersBase102delta

	//UccMemory, err := strconv.ParseInt(oldGp.ChainParameters.UccMemory, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.UccMemory = UccMemory
	//UccCpuShares, err := strconv.ParseInt(oldGp.ChainParameters.UccCpuShares, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.UccCpuShares = UccCpuShares
	//UccCpuQuota, err := strconv.ParseInt(oldGp.ChainParameters.UccCpuQuota, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.UccCpuQuota = UccCpuQuota
	//newData.ChainParameters.UccDisk = core.DefaultUccDisk
	//
	//TempUccMemory, err := strconv.ParseInt(oldGp.ChainParameters.TempUccMemory, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.TempUccMemory = TempUccMemory
	//TempUccCpuShares, err := strconv.ParseInt(oldGp.ChainParameters.TempUccCpuShares, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.TempUccCpuShares = TempUccCpuShares
	//TempUccCpuQuota, err := strconv.ParseInt(oldGp.ChainParameters.TempUccCpuQuota, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.TempUccCpuQuota = TempUccCpuQuota
	//
	//ContractSignatureNum, err := strconv.ParseInt(oldGp.ChainParameters.ContractSignatureNum, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.ContractSignatureNum = int(ContractSignatureNum)
	//ContractElectionNum, err := strconv.ParseInt(oldGp.ChainParameters.ContractElectionNum, 10, 64)
	//if err != nil {
	//	return err
	//}
	//newData.ChainParameters.ContractElectionNum = int(ContractElectionNum)
	//
	//newData.ChainParameters.ContractTxTimeoutUnitFee = core.DefaultContractTxTimeoutUnitFee
	//newData.ChainParameters.ContractTxSizeUnitFee = core.DefaultContractTxSizeUnitFee
	//
	//newData.ChainParameters.ContractTxInstallFeeLevel = core.DefaultContractTxInstallFeeLevel
	//newData.ChainParameters.ContractTxDeployFeeLevel = core.DefaultContractTxDeployFeeLevel
	//newData.ChainParameters.ContractTxInvokeFeeLevel = core.DefaultContractTxInvokeFeeLevel
	//newData.ChainParameters.ContractTxStopFeeLevel = core.DefaultContractTxStopFeeLevel

	newData.ChainParameters.UccMemory = oldGp.ChainParameters.UccMemory
	newData.ChainParameters.UccCpuShares = oldGp.ChainParameters.UccCpuShares
	newData.ChainParameters.UccCpuQuota = oldGp.ChainParameters.UccCpuQuota
	newData.ChainParameters.UccDisk = strconv.FormatInt(core.DefaultUccDisk, 10)

	newData.ChainParameters.TempUccMemory = oldGp.ChainParameters.TempUccMemory
	newData.ChainParameters.TempUccCpuShares = oldGp.ChainParameters.TempUccCpuShares
	newData.ChainParameters.TempUccCpuQuota = oldGp.ChainParameters.TempUccCpuQuota

	newData.ChainParameters.ContractSignatureNum = oldGp.ChainParameters.ContractSignatureNum
	newData.ChainParameters.ContractElectionNum = oldGp.ChainParameters.ContractElectionNum

	newData.ChainParameters.ContractTxTimeoutUnitFee = strconv.FormatInt(core.DefaultContractTxTimeoutUnitFee, 10)
	newData.ChainParameters.ContractTxSizeUnitFee = strconv.FormatInt(core.DefaultContractTxSizeUnitFee, 10)

	newData.ChainParameters.ContractTxInstallFeeLevel =
		strconv.FormatFloat(core.DefaultContractTxInstallFeeLevel, 'f', -1, 64)
	newData.ChainParameters.ContractTxDeployFeeLevel =
		strconv.FormatFloat(core.DefaultContractTxDeployFeeLevel, 'f', -1, 64)
	newData.ChainParameters.ContractTxInvokeFeeLevel =
		strconv.FormatFloat(core.DefaultContractTxInvokeFeeLevel, 'f', -1, 64)
	newData.ChainParameters.ContractTxStopFeeLevel =
		strconv.FormatFloat(core.DefaultContractTxStopFeeLevel, 'f', -1, 64)

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type GlobalProperty100 struct {
	GlobalPropBase100

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

type GlobalPropBase100 struct {
	ImmutableParameters ImmutableChainParameters101 // 不可改变的区块链网络参数
	ChainParameters     ChainParameters100          // 区块链网络参数
}

type ChainParameters100 struct {
	ChainParametersBase102delta

	DepositDailyReward string
	DepositPeriod      string

	UccMemory     string
	UccMemorySwap string
	UccCpuShares  string
	UccCpuQuota   string
	UccCpuPeriod  string

	TempUccMemory     string
	TempUccMemorySwap string
	TempUccCpuShares  string
	TempUccCpuQuota   string

	ContractSignatureNum string
	ContractElectionNum  string
}
