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
	//  testnet
	addrAndPubKey := make(map[string]string)
	addrAndPubKey["P1HHQt1cMWrj3WBVmsTXYBsawwg2wrD6of6"] = "02d25b72c792d2aa83b45b07b4bfca3a6430dd3475bf855133742c51a40e81a2c8"
	addrAndPubKey["P1JxYp1dRpq2ZeYk58XEkSZJrptYEeuvZyq"] = "033f6c1422e15dd57ea960e957defad143b8be33e7c0ff055bdfdee00f6d094371"
	addrAndPubKey["P1KP5TZwTY8UowE7X3zSZ3gZDHqwCqcCThR"] = "02bbe16817006c18969ff559ea14522c67e4115d883fcd6dce9a86991ca097153a"
	addrAndPubKey["P1P3jZb43Y7stahiv8G3yvRUawGcWWvJBPt"] = "02c2f34e741a5840d49606f9533c1fb810ff4c5df7982773bbf32a058abfb4eaa4"
	addrAndPubKey["P1PuhsNTmpsSV36wyoEF49b5dhRdaTQYC2C"] = "03bcefd355332c533086f86570aba4ea470e36ada86788f55568f9e08e2f21c3bb"
	//  testnet
	statedb := storage.NewStateDb(m.statedb)
	//  获取列表
	list, err := statedb.GetCandidateMediatorList()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	for addr := range list {
		juror := modules.Juror{}
		juror.Address = addr
		juror.Role = modules.Jury
		juror.Balance = 0
		//  获取超级节点进入时间
		mediatorByte, v, err := statedb.GetContractState(syscontract.DepositContractAddress.Bytes(), storage.MediatorDepositKey(addr))
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		mediator := modules.MediatorDeposit{}
		err = json.Unmarshal(mediatorByte, &mediator)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
		juror.EnterTime = mediator.EnterTime
		pubbyte, err := hex.DecodeString(addrAndPubKey[addr])
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
		juror.PublicKey = pubbyte
		jurorByte, err := json.Marshal(juror)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
		ws1 := modules.NewWriteSet(string(constants.DEPOSIT_JURY_BALANCE_PREFIX)+addr, jurorByte)
		ws := []modules.ContractWriteSet{*ws1}
		err = statedb.SaveContractStates(syscontract.DepositContractAddress.Bytes(), ws, v)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}
	}
	return nil
}
