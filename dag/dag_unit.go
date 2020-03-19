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

package dag

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	dagcommon "github.com/palletone/go-palletone/dag/common"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/consensus/jury"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/txspool"
)

func (dag *Dag) getBePackedTxs(txp txspool.ITxPool, cp *jury.Processor,
	producer common.Address, ks *keystore.KeyStore) ([]*modules.Transaction, error) {
	unitNumber := dag.HeadUnitNum() + 1
	unitId := fmt.Sprintf("%d", unitNumber)

	rwM, err := rwset.NewRwSetMgr(unitId)
	if err != nil {
		errStr := fmt.Sprintf("NewRwSetMgr err: %v", err.Error())
		log.Errorf(errStr)
		return nil, fmt.Errorf(errStr)
	}
	defer rwM.Close()

	if err := cp.AddContractLoop(rwM, txp, producer, ks); err != nil {
		log.Debugf("AddContractLoop err: %v", err.Error())
	}
	// close tx simulator (系统合约)

	//广播节点选取签名请求事件
	go cp.BroadcastElectionSigRequestEvent()

	// 从TxPool抓取排序后的Tx，如果是系统合约请求，则执行
	poolTxs, _ := txp.GetSortedTxs(common.Hash{}, unitNumber)
	log.DebugDynamic(func() string {
		txHash := ""
		for _, tx := range poolTxs {
			txHash += tx.Tx.Hash().String() + ";"
		}
		return "txpool GetSortedTxs return:" + txHash
	})

	//TODO Jay 这里的txpool.GetSortedTxs返回顺序有问题，所以再次排序
	poolTxMap := make(map[common.Hash]*modules.Transaction)
	for _, ptx := range poolTxs {
		poolTxMap[ptx.Tx.Hash()] = ptx.Tx
	}
	sortedTxs, orphanTxs, dsTxs := modules.SortTxs(poolTxMap, dag.GetUtxoEntry)
	log.DebugDynamic(func() string {
		txHash := ""
		for _, tx := range sortedTxs {
			txHash += tx.Hash().String() + ";"
		}
		return "modules.SortTxs return:" + txHash
	})
	if len(orphanTxs) > 0 {
		log.InfoDynamic(func() string {
			otxHash := ""
			for _, tx := range orphanTxs {
				otxHash += tx.Hash().String() + ";"
			}
			return "modules.SortTxs find orphan txs:" + otxHash
		})
	}
	if len(dsTxs) > 0 {
		otxHash := ""
		for _, tx := range dsTxs {
			otxHash += tx.Hash().String() + ";"
		}
		log.Warnf("modules.SortTxs find double spend txs:%s", otxHash)
	}

	//创建TempDAG，用于临时存储Tx执行的结果
	tempDag, err := dag.NewTemp()
	log.Debug("create a new tempDag for generate unit")
	if err != nil {
		log.Errorf("Init temp dag error:%s", err.Error())
	}
	tx4Pack := []*modules.Transaction{}
	for i, tx := range sortedTxs {
		log.Debugf("pack tx[%s] into unit[#%d]", tx.RequestHash().String(), unitNumber)
		if tx.IsSystemContract() && tx.IsNewContractInvokeRequest() { //是未执行的系统合约
			signedTx, err := cp.RunAndSignTx(tx, rwM, tempDag, producer, ks)
			if err != nil {
				log.Errorf("run contract request[%s] fail:%s", tx.Hash().String(), err.Error())
				continue
			}
			err = tempDag.SaveTransaction(signedTx, i+1) //第0条是Coinbase
			if err != nil {
				log.Errorf("save tx[%s] req[%s] get error:%s", signedTx.Hash().String(),
					signedTx.RequestHash().String(), err.Error())
			}
			tx4Pack = append(tx4Pack, signedTx)
		} else { //不需要执行，直接打包
			err = tempDag.SaveTransaction(tx, i+1)
			if err != nil {
				log.Errorf("save tx[%s] req[%s] get error:%s", tx.Hash().String(),
					tx.RequestHash().String(), err.Error())
			}
			tx4Pack = append(tx4Pack, tx)
		}
	}

	return tx4Pack, nil
}

// GenerateUnit, generate unit
func (dag *Dag) GenerateUnit(when time.Time, producer common.Address, groupPubKey []byte, ks *keystore.KeyStore,
	txp txspool.ITxPool, cp *jury.Processor) (*modules.Unit, error) {
	t0 := time.Now()
	defer func(start time.Time) {
		log.Debugf("GenerateUnit cost time: %v", time.Since(start))
	}(t0)

	// 判断是否满足生产的若干条件
	//log.Debugf("generate unit ...")

	// 获取待打包的交易
	ptx, err := dag.getBePackedTxs(txp, cp, producer, ks)
	if err != nil {
		return nil, err
	}

	// 生产unit，添加交易集、时间戳、签名
	unsign_unit, err := dag.createUnit(producer, ptx, when)
	if err != nil {
		errStr := fmt.Sprintf("createUnit error: %v", err.Error())
		log.Debug(errStr)
		return nil, fmt.Errorf(errStr)
	}
	if unsign_unit == nil || unsign_unit.IsEmpty() {
		errStr := fmt.Sprintf("No unit need to be packaged for now.")
		log.Debug(errStr)
		return nil, fmt.Errorf(errStr)
	}

	sign_unit, err := dagcommon.GetUnitWithSig(unsign_unit, ks, producer)
	if err != nil {
		errStr := fmt.Sprintf("GetUnitWithSig error: %v", err.Error())
		log.Debug(errStr)
		return nil, fmt.Errorf(errStr)
	}
	sign_unit.UnitHeader.SetGroupPubkey(groupPubKey)

	// sign_unit.Size()
	log.Debugf("Generate new unit index:[%d],hash:[%s],size:%s, parent unit[%s],txs[%d], spent time: %s ",
		sign_unit.NumberU64(), sign_unit.Hash().String(), sign_unit.Size().String(),
		sign_unit.UnitHeader.ParentHash()[0].String(), sign_unit.Txs.Len(), time.Since(t0).String())

	// 将新单元添加到MemDag中
	a, b, c, d, e, err := dag.Memdag.AddUnit(sign_unit, txp, true)
	if a != nil && err == nil {
		if dag.unstableUnitProduceRep != e {
			log.Debugf("send UnstableRepositoryUpdatedEvent")
			go dag.unstableRepositoryUpdatedFeed.Send(modules.UnstableRepositoryUpdatedEvent{})
		}

		dag.unstableUnitRep = a
		dag.unstableUtxoRep = b
		dag.unstableStateRep = c
		dag.unstablePropRep = d
		dag.unstableUnitProduceRep = e
	} else if err != nil {
		errStr := fmt.Sprintf("Memdag AddUnit[%s] error: %v", sign_unit.Hash().String(), err.Error())
		udata, _ := json.Marshal(sign_unit)
		rdata, _ := rlp.EncodeToBytes(sign_unit)
		log.Errorf("%s, Unit data:%s,Rlp:%x", errStr, string(udata), rdata)
		return nil, fmt.Errorf(errStr)
	}
	sign_unit.ReceivedAt = time.Now()

	//4.PostChainEvents
	//TODO add PostChainEvents
	go func() {
		var (
			events = make([]interface{}, 0, 2)
		)
		events = append(events, modules.ChainHeadEvent{Unit: sign_unit})
		events = append(events, modules.SaveUnitEvent{Unit: sign_unit})
		events = append(events, modules.ChainEvent{Unit: sign_unit, Hash: sign_unit.Hash()})
		dag.PostChainEvents(events)
	}()

	return sign_unit, nil
}

// createUnit, create a unit when mediator being produced
//创建未签名的Unit
func (d *Dag) createUnit(mAddr common.Address, txs []*modules.Transaction,
	when time.Time) (*modules.Unit, error) {

	med, err := d.unstableStateRep.RetrieveMediator(mAddr)
	if err != nil {
		return nil, err
	}

	//return d.unstableUnitRep.CreateUnit(med.GetRewardAdd(), txpool, rep, state.GetJurorReward)
	return d.unstableUnitRep.CreateUnit(med.GetRewardAdd(), txs, when,
		d.unstablePropRep, d.unstableStateRep.GetJurorReward)
}

func (d *Dag) GetNewestUnit(token modules.AssetId) (common.Hash, *modules.ChainIndex, error) {
	return d.unstablePropRep.GetNewestUnit(token)
}
