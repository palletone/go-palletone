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

	TxCoinYearRate string
	DepositRate    string
	DepositPeriod  string

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

func (cp *ChainParameters) EncodeRLP(w io.Writer) error {
	cpt := cp.getCPT()

	return rlp.Encode(w, cpt)
}

func (cp *ChainParameters) getCPT() *chainParameters {
	return &chainParameters{
		ChainParametersBase: cp.ChainParametersBase,

		TxCoinYearRate: strconv.FormatFloat(float64(cp.TxCoinYearRate), 'f', -1, 64),
		DepositRate:    strconv.FormatFloat(float64(cp.DepositRate), 'f', -1, 64),
		DepositPeriod:  strconv.FormatInt(int64(cp.DepositPeriod), 10),

		UccMemory:     strconv.FormatInt(int64(cp.UccMemory), 10),
		UccMemorySwap: strconv.FormatInt(int64(cp.UccMemorySwap), 10),
		UccCpuShares:  strconv.FormatInt(int64(cp.UccCpuShares), 10),
		UccCpuQuota:   strconv.FormatInt(int64(cp.UccCpuQuota), 10),
		UccCpuPeriod:  strconv.FormatInt(int64(cp.UccCpuPeriod), 10),

		TempUccMemory:     strconv.FormatInt(int64(cp.TempUccMemory), 10),
		TempUccMemorySwap: strconv.FormatInt(int64(cp.TempUccMemorySwap), 10),
		TempUccCpuShares:  strconv.FormatInt(int64(cp.TempUccCpuShares), 10),
		TempUccCpuQuota:   strconv.FormatInt(int64(cp.TempUccCpuQuota), 10),

		ContractSignatureNum: strconv.FormatInt(int64(cp.ContractSignatureNum), 10),
		ContractElectionNum:  strconv.FormatInt(int64(cp.ContractElectionNum), 10),
	}
}

func (cpt *chainParameters) getCP(cp *ChainParameters) error {
	cp.ChainParametersBase = cpt.ChainParametersBase

	TxCoinYearRate, err := strconv.ParseFloat(cpt.TxCoinYearRate, 64)
	if err != nil {
		return err
	}
	cp.TxCoinYearRate = float64(TxCoinYearRate)

	DepositRate, err := strconv.ParseFloat(cpt.DepositRate, 64)
	if err != nil {
		return err
	}
	cp.DepositRate = float64(DepositRate)

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

	UccMemorySwap, err := strconv.ParseInt(cpt.UccMemorySwap, 10, 64)
	if err != nil {
		return err
	}
	cp.UccMemorySwap = int64(UccMemorySwap)

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

	UccCpuPeriod, err := strconv.ParseInt(cpt.UccCpuPeriod, 10, 64)
	if err != nil {
		return err
	}
	cp.UccCpuPeriod = int64(UccCpuPeriod)

	TempUccMemory, err := strconv.ParseInt(cpt.TempUccMemory, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccMemory = int64(TempUccMemory)

	TempUccMemorySwap, err := strconv.ParseInt(cpt.TempUccMemorySwap, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccMemorySwap = int64(TempUccMemorySwap)

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
