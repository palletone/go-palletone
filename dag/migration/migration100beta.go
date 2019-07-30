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
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
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
	dbop := storage.NewUtxoDb(m.utxodb)
	utxos, err := dbop.GetAllUtxos()
	if err != nil {
		return err
	}
	for outpoint, utxo := range utxos {
		if utxo.IsSpent() {
			err = dbop.DeleteUtxo(&outpoint, common.Hash{}, 0)
			if err != nil {
				log.Errorf("Migrate utxo db,delete spent utxo error:%s", err.Error())
				return err
			}
			//log.Debugf("Deleted spent UTXO by key:%s", outpoint.String())
		}
	}
	dagdb := storage.NewDagDb(m.dagdb)
	txs, err := dagdb.GetAllTxs()
	if err != nil {
		log.Error(err.Error())
	}
	log.Debugf("Tx count:%d", len(txs))
	for i, tx := range txs {
		if tx == nil {
			log.Errorf("tx[%d] is nil", i)
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
	statedb := storage.NewStateDb(m.statedb)
	oldMediators := statedb.GetPrefix(constants.MEDIATOR_INFO_PREFIX)

	for key, value := range oldMediators {
		oldMediator := &OldMediatorInfo{}
		err := rlp.DecodeBytes(value, oldMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		newMediator := &modules.MediatorInfo{
			MediatorInfoBase:   oldMediator.MediatorInfoBase,
			MediatorApplyInfo:  &core.MediatorApplyInfo{Description: oldMediator.ApplyInfo},
			MediatorInfoExpand: oldMediator.MediatorInfoExpand,
		}

		err = storage.StoreToRlpBytes(m.statedb, []byte(key), newMediator)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}
	}

	return nil
}

type OldMediatorInfo struct {
	*core.MediatorInfoBase
	*OldMediatorApplyInfo
	*core.MediatorInfoExpand
}

type OldMediatorApplyInfo struct {
	ApplyInfo string `json:"applyInfo"` //  申请信息
}

func (m *Migration100_101) upgradeGP() error {
	oldGp := OldGlobalProperty{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, &oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	newData := &modules.GlobalPropertys{}
	newData.ActiveJuries = oldGp.ActiveJuries
	newData.ActiveMediators = oldGp.ActiveMediators
	newData.PrecedingMediators = oldGp.PrecedingMediators
	newData.ImmutableParameters = oldGp.ImmutableParameters
	newData.ChainParameters.ChainParametersBase = oldGp.ChainParameters.ChainParametersBase

	newData.ChainParameters.UccDisk = core.DefaultUccDisk

	newData.ChainParameters.ContractTxTimeoutUnitFee = core.DefaultContractTxTimeoutUnitFee
	newData.ChainParameters.ContractTxSizeUnitFee = core.DefaultContractTxSizeUnitFee
	newData.ChainParameters.ContractTxInstallFeeLevel = core.DefaultContractTxInstallFeeLevel
	newData.ChainParameters.ContractTxDeployFeeLevel = core.DefaultContractTxDeployFeeLevel
	newData.ChainParameters.ContractTxInvokeFeeLevel = core.DefaultContractTxInvokeFeeLevel
	newData.ChainParameters.ContractTxStopFeeLevel = core.DefaultContractTxStopFeeLevel

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type OldGlobalProperty struct {
	OldGlobalPropBase

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

type OldGlobalPropBase struct {
	ImmutableParameters core.ImmutableChainParameters // 不可改变的区块链网络参数
	ChainParameters     OldChainParameters            // 区块链网络参数
}

type OldChainParameters struct {
	core.ChainParametersBase
	DepositDailyReward   string
	DepositPeriod        string
	UccMemory            string
	UccMemorySwap        string
	UccCpuShares         string
	UccCpuQuota          string
	UccCpuPeriod         string
	TempUccMemory        string
	TempUccMemorySwap    string
	TempUccCpuShares     string
	TempUccCpuQuota      string
	ContractSignatureNum string
	ContractElectionNum  string
}
