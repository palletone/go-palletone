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
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration104beta_105alpha struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration104beta_105alpha) FromVersion() string {
	return "1.0.4-beta"
}

func (m *Migration104beta_105alpha) ToVersion() string {
	return "1.0.5-alpha"
}

func (m *Migration104beta_105alpha) ExecuteUpgrade() error {
	//
	if err := m.upgradeContractInfoFromProperdbToStatedb(); err != nil {
		return err
	}
	return nil
}

func getprefix(db storage.DatabaseReader, prefix []byte) map[string][]byte {
	iter := db.NewIteratorWithPrefix(prefix)
	result := make(map[string][]byte)
	for iter.Next() {
		key := make([]byte, 0)
		value := make([]byte, 0)
		key = append(key, iter.Key()...)
		result[string(key)] = append(value, iter.Value()...)
	}
	return result
}

func (m *Migration104beta_105alpha) upgradeContractInfoFromProperdbToStatedb() error {
	//  状态数据库
	rows := getprefix(m.statedb, constants.CONTRACT_PREFIX)
	oldContract := make([]*OldContract, 0, len(rows))
	for _, v := range rows {
		contract := &OldContract{}
		rlp.DecodeBytes(v, contract)
		oldContract = append(oldContract, contract)
	}
	//  状态数据库
	for _, c := range oldContract {
		newContract := NewContract{
			ContractId:   c.ContractId,
			TemplateId:   c.TemplateId,
			Name:         c.Name,
			Status:       c.Status,
			Creator:      c.Creator,
			CreationTime: c.CreationTime,
			DuringTime:   c.DuringTime,
			Version:      "",
		}

		//  用户合约
		if len(c.TemplateId) != 0 {
			cc := list.CCInfo{}
			key := append(constants.JURY_PROPERTY_USER_CONTRACT_KEY, c.ContractId...)
			err := storage.RetrieveFromRlpBytes(m.propdb, key, &cc)
			if err != nil {
				return err
			}
			newContract.Version = cc.Version
			//  保存对应的陪审员地址
			juryAddr, _ := common.StringToAddress(cc.Address)
			key1 := append(constants.CONTRACT_JURY_PREFIX, juryAddr.Bytes()...)
			key2 := append(key1, cc.Id...)
			log.Debugf("save contract id = %v with jury address = %s,key1 = %v", cc.Id, juryAddr.String(), key1)
			//  保存陪审员对应的状态
			err = storage.StoreToRlpBytes(m.statedb, key2, &newContract)
			if err != nil {
				return err
			}
		}
		//  保存新的 contract
		key := append(constants.CONTRACT_PREFIX, c.ContractId...)
		err := storage.StoreToRlpBytes(m.statedb, key, &newContract)
		if err != nil {
			return err
		}
	}
	return nil
}

type OldContract struct {
	// 根据用户创建合约实例申请的RequestId截取其后20字节生成
	ContractId   []byte
	TemplateId   []byte
	Name         string
	Status       byte   // 合约状态
	Creator      []byte // address 20bytes
	CreationTime uint64 // creation  date
	DuringTime   uint64 //合约部署持续时间，单位秒
}

type NewContract struct {
	// 根据用户创建合约实例申请的RequestId截取其后20字节生成
	ContractId   []byte
	TemplateId   []byte
	Name         string
	Status       byte   // 合约状态
	Creator      []byte // address 20bytes
	CreationTime uint64 // creation  date
	DuringTime   uint64 //合约部署持续时间，单位秒
	Version      string
}
