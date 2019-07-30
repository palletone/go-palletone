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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package core

import (
	"io"
	"strconv"

	"github.com/ethereum/go-ethereum/rlp"
)

type chainParameters struct {
	ChainParametersBase

	// TxCoinYearRate     string
	DepositDailyReward string
	DepositPeriod      string

	UccMemory     string
	//UccMemorySwap string
	UccCpuShares  string
	UccCpuQuota   string
	//UccCpuPeriod  string
	UccDisk       string

	TempUccMemory     string
	//TempUccMemorySwap string
	TempUccCpuShares  string
	TempUccCpuQuota   string

	ContractSignatureNum      string
	ContractElectionNum       string
	ContractTxTimeoutUnitFee  string
	ContractTxSizeUnitFee     string
	ContractTxInstallFeeLevel string
	ContractTxDeployFeeLevel  string
	ContractTxInvokeFeeLevel  string
	ContractTxStopFeeLevel    string
}

func (cp *ChainParameters) EncodeRLP(w io.Writer) error {
	cpt := cp.getCPT()

	return rlp.Encode(w, cpt)
}

func (cp *ChainParameters) getCPT() *chainParameters {
	return &chainParameters{
		ChainParametersBase: cp.ChainParametersBase,

		// TxCoinYearRate:     strconv.FormatFloat(float64(cp.TxCoinYearRate), 'f', -1, 64),
		DepositDailyReward: strconv.FormatInt(int64(cp.PledgeDailyReward), 10),
		DepositPeriod:      strconv.FormatInt(int64(cp.DepositPeriod), 10),

		UccMemory:    strconv.FormatInt(int64(cp.UccMemory), 10),
		UccCpuShares: strconv.FormatInt(int64(cp.UccCpuShares), 10),
		UccCpuQuota:  strconv.FormatInt(int64(cp.UccCpuQuota), 10),
		UccDisk:      strconv.FormatInt(int64(cp.UccDisk), 10),

		TempUccMemory:    strconv.FormatInt(int64(cp.TempUccMemory), 10),
		TempUccCpuShares: strconv.FormatInt(int64(cp.TempUccCpuShares), 10),
		TempUccCpuQuota:  strconv.FormatInt(int64(cp.TempUccCpuQuota), 10),

		ContractSignatureNum: strconv.FormatInt(int64(cp.ContractSignatureNum), 10),
		ContractElectionNum:  strconv.FormatInt(int64(cp.ContractElectionNum), 10),

		ContractTxTimeoutUnitFee:  strconv.FormatUint(uint64(cp.ContractTxTimeoutUnitFee), 10),
		ContractTxSizeUnitFee:     strconv.FormatUint(uint64(cp.ContractTxSizeUnitFee), 10),
		ContractTxInstallFeeLevel: strconv.FormatFloat(float64(cp.ContractTxInstallFeeLevel), 'f', -1, 64),
		ContractTxDeployFeeLevel:  strconv.FormatFloat(float64(cp.ContractTxDeployFeeLevel), 'f', -1, 64),
		ContractTxInvokeFeeLevel:  strconv.FormatFloat(float64(cp.ContractTxInvokeFeeLevel), 'f', -1, 64),
		ContractTxStopFeeLevel:    strconv.FormatFloat(float64(cp.ContractTxStopFeeLevel), 'f', -1, 64),
	}
}

func (cpt *chainParameters) getCP(cp *ChainParameters) error {
	cp.ChainParametersBase = cpt.ChainParametersBase

	// TxCoinYearRate, err := strconv.ParseFloat(cpt.TxCoinYearRate, 64)
	// if err != nil {
	// 	return err
	// }
	// cp.TxCoinYearRate = float64(TxCoinYearRate)

	DepositDailyReward, err := strconv.ParseInt(cpt.DepositDailyReward, 10, 64)
	if err != nil {
		return err
	}
	cp.PledgeDailyReward = uint64(DepositDailyReward)

	DepositPeriod, err := strconv.ParseInt(cpt.DepositPeriod, 10, 64)
	if err != nil {
		return err
	}
	cp.DepositPeriod = int(DepositPeriod)

	UccMemory, err := strconv.ParseInt(cpt.UccMemory, 10, 64)
	if err != nil {
		return err
	}
	cp.UccMemory = int64(UccMemory)

	UccCpuShares, err := strconv.ParseInt(cpt.UccCpuShares, 10, 64)
	if err != nil {
		return err
	}
	cp.UccCpuShares = int64(UccCpuShares)

	UccCpuQuota, err := strconv.ParseInt(cpt.UccCpuQuota, 10, 64)
	if err != nil {
		return err
	}
	cp.UccCpuQuota = int64(UccCpuQuota)

	UccDisk, err := strconv.ParseInt(cpt.UccDisk, 10, 64)
	if err != nil {
		return err
	}
	cp.UccDisk = int64(UccDisk)

	TempUccMemory, err := strconv.ParseInt(cpt.TempUccMemory, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccMemory = int64(TempUccMemory)

	TempUccCpuShares, err := strconv.ParseInt(cpt.TempUccCpuShares, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccCpuShares = int64(TempUccCpuShares)

	TempUccCpuQuota, err := strconv.ParseInt(cpt.TempUccCpuQuota, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccCpuQuota = int64(TempUccCpuQuota)

	ContractSignatureNum, err := strconv.ParseInt(cpt.ContractSignatureNum, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractSignatureNum = int(ContractSignatureNum)

	ContractElectionNum, err := strconv.ParseInt(cpt.ContractElectionNum, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractElectionNum = int(ContractElectionNum)

	ContractTxTimeoutUnitFee, err := strconv.ParseUint(cpt.ContractTxTimeoutUnitFee, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractTxTimeoutUnitFee = ContractTxTimeoutUnitFee

	ContractTxSizeUnitFee, err := strconv.ParseUint(cpt.ContractTxSizeUnitFee, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractTxSizeUnitFee = ContractTxSizeUnitFee

	ContractTxInstallFeeLevel, err := strconv.ParseFloat(cpt.ContractTxInstallFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxInstallFeeLevel = ContractTxInstallFeeLevel

	ContractTxDeployFeeLevel, err := strconv.ParseFloat(cpt.ContractTxDeployFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxDeployFeeLevel = ContractTxDeployFeeLevel

	ContractTxInvokeFeeLevel, err := strconv.ParseFloat(cpt.ContractTxInvokeFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxInvokeFeeLevel = ContractTxInvokeFeeLevel

	ContractTxStopFeeLevel, err := strconv.ParseFloat(cpt.ContractTxStopFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxStopFeeLevel = ContractTxStopFeeLevel

	return nil
}

func (cp *ChainParameters) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}

	cpt := &chainParameters{}
	err = rlp.DecodeBytes(raw, cpt)
	if err != nil {
		return err
	}

	err = cpt.getCP(cp)
	if err != nil {
		return err
	}

	return nil
}
