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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration103alpha_103beta struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration103alpha_103beta) FromVersion() string {
	return "1.0.3-alpha"
}

func (m *Migration103alpha_103beta) ToVersion() string {
	return "1.0.3-beta"
}

func (m *Migration103alpha_103beta) ExecuteUpgrade() error {
	//给默认的超级节点列表添加对应Juror账户信息
	if err := m.upgradeDefaultMediatorsWithJurorInfo(); err != nil {
		return err
	}
	return nil
}

func (m *Migration103alpha_103beta) upgradeDefaultMediatorsWithJurorInfo() error {
	// 处理genesis中的几个特殊的mediator（由于没有初始保证金的bug）
	dagDb := storage.NewDagDb(m.dagdb)
	ghash, err := dagDb.GetGenesisUnitHash()
	if err != nil {
		errStr := fmt.Sprintf("GetGenesisUnitHash err: %v", err.Error())
		log.Error(errStr)
		return fmt.Errorf(errStr)
	}

	uHeader, err := dagDb.GetHeaderByHash(ghash)
	if err != nil {
		errStr := fmt.Sprintf("GetHeaderByHash err: %v", err.Error())
		log.Error(errStr)
		return fmt.Errorf(errStr)
	}

	statedb := storage.NewStateDb(m.statedb)
	// 获取mediator候选列表
	list, err := statedb.GetCandidateMediatorList()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	for addr := range list {
		var pubKey string
		var isFind bool
		var version *modules.StateVersion

		juror := modules.JurorDeposit{}
		juror.Address = addr
		juror.Role = modules.Jury
		juror.Balance = 0

		// genesis的mediator可能没有缴纳保证金
		if pubKey, isFind = constants.OldGenesisMediatorAndPubKey[addr]; isFind {
			version = &modules.StateVersion{
				Height:  uHeader.GetNumber(),
				TxIndex: ^uint32(0),
			}

			ti := time.Unix(uHeader.Timestamp(), 0)
			juror.EnterTime = ti.UTC().Format(modules.Layout2)
		} else if pubKey, isFind = constants.OldMediatorAndPubKey[addr]; isFind {
			//  获取超级节点进入时间
			var mediatorByte []byte
			mediatorByte, version, err = statedb.GetContractState(syscontract.DepositContractAddress.Bytes(),
				storage.MediatorDepositKey(addr))
			if err != nil {
				log.Errorf(err.Error())
				//continue
				return err
			}

			mediator := modules.MediatorDeposit{}
			err = json.Unmarshal(mediatorByte, &mediator)
			if err != nil {
				log.Errorf(err.Error())
				return err
			}
			juror.EnterTime = mediator.EnterTime
		} else {
			errStr := fmt.Sprintf("not find this mediator's PubKey: %v", addr)
			log.Error(errStr)
			return fmt.Errorf(errStr)
		}

		pubByte, err := hex.DecodeString(pubKey)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		//  判断公钥的正确性
		if crypto.PubkeyBytesToAddress(pubByte).String() != addr {
			errStr := fmt.Sprintf("address = %s, public key = %s error", addr, pubKey)
			log.Error(errStr)
			return fmt.Errorf(errStr)
		}
		juror.PublicKey = pubByte
		jurorByte, err := json.Marshal(juror)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		ws1 := modules.NewWriteSet(string(constants.DEPOSIT_JURY_BALANCE_PREFIX)+addr, jurorByte)
		//ws := []modules.ContractWriteSet{*ws1}
		err = statedb.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws1, version)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
	}

	return nil
}
