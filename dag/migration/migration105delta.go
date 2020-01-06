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
 *  * @author PalletOne core developer albert <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration105gamma_105delta struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration105gamma_105delta) FromVersion() string {
	return "1.0.5-gamma"
}

func (m *Migration105gamma_105delta) ToVersion() string {
	return "1.0.5-delta"
}

func (m *Migration105gamma_105delta) ExecuteUpgrade() error {
	// 将 statebd 中的指定陪审员的合约实例关联到对应陪审员
	if err := m.upgradeContracts(); err != nil {
		return err
	}

	return nil
}

func (m *Migration105gamma_105delta) upgradeContracts() error {
	log.Infof("upgradeContracts...")
	//  获取所有合约
	//  状态数据库所有
	rows := getprefix(m.statedb, constants.CONTRACT_PREFIX)
	contracts := make([]*NewContract, 0, len(rows))
	for _, v := range rows {
		contract := &NewContract{}
		rlp.DecodeBytes(v, contract)
		contracts = append(contracts, contract)
	}
	log.Debugf("contracts len = %d", len(contracts))
	//  获取陪审员列表
	list, err := getjurycandidatelist(m.statedb)
	if err != nil {
		return err
	}
	if list == nil {
		log.Debug("list is nil")
		return nil
	}
		log.Debugf("jury list len = %d", len(list))
	//  获取对应的模板
	//  获取陪审员地址
	contractNameAndJuryAddr := make(map[string][]common.Address, 0)
	temp := 0
	for i, c := range contracts {
		if len(contracts[i].TemplateId) != 0 {
			key := append(constants.CONTRACT_TPL, contracts[i].TemplateId...)
			tpl := modules.ContractTemplate{}
			err := storage.RetrieveFromRlpBytes(m.statedb, key, &tpl)
			if err != nil {
				return err
			}
			if len(tpl.AddrHash) != 0 {
				temp++
				for _, v := range tpl.AddrHash {
					for k, _ := range list {
						a, _ := common.StringToAddress(k)
						if util.RlpHash(a).String() == v.String() {
							contractNameAndJuryAddr[c.Name] = append(contractNameAndJuryAddr[c.Name], a)
						}
					}
				}
			}
		}
	}
	log.Debugf("temp len = %d", temp)
	log.Debugf("contractNameAndJuryAddr len = %d", len(contractNameAndJuryAddr))
	//  保存 jury 对应新合约
	if len(contractNameAndJuryAddr) != 0 {
		for _, nc := range contracts {
			//  保存对应的陪审员地址
			for _, juryAddr := range contractNameAndJuryAddr[nc.Name] {
				key1 := append(constants.CONTRACT_JURY_PREFIX, juryAddr.Bytes()...)
				key2 := append(key1, nc.ContractId...)
				log.Debugf("save contract id = %v with jury address = %s,key1 = %v", nc.ContractId, juryAddr.String(), key1)
				//  保存陪审员对应的状态
				err := storage.StoreToRlpBytes(m.statedb, key2, nc)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getjurycandidatelist(db storage.DatabaseReader) (map[string]bool, error) {
	depositeContractAddress := syscontract.DepositContractAddress
	key := getContractStateKey(depositeContractAddress.Bytes(), modules.JuryList)
	data, err := db.Get(key)
	if err != nil {
		log.Error(err.Error())
		return nil, nil
	}
	if len(data) < 28 {
		return nil, errors.New("the data is irregular.")
	}
	verBytes := data[:28]
	objData := data[28:]
	version := &modules.StateVersion{}
	version.SetBytes(verBytes)
	//return objData, version, nil
	candidateList := make(map[string]bool)
	err = json.Unmarshal(objData, &candidateList)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return candidateList, nil
}

func getContractStateKey(id []byte, field string) []byte {
	key := append(constants.CONTRACT_STATE_PREFIX, id...)
	return append(key, field...)
}
