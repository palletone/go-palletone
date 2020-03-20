/*
 *  This file is part of go-palletone.
 *  go-palletone is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *  go-palletone is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *  You should have received a copy of the GNU General Public License
 *  along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 *
 *  @author PalletOne core developer <dev@pallet.one>
 *  @date 2018-2020
 */

package txpool2

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/txspool"
)

//维护了一个图
type linkTx struct {
	Parents  []*linkTx
	Tx       *txspool.TxPoolTransaction
	Children []*linkTx
	Number   int
}

//维护了一个有序的Tx DAG
type txList struct {
	txs         map[common.Hash]*linkTx
	linkTxRoots map[common.Hash]*linkTx
	newUtxo     map[modules.OutPoint]*modules.Utxo
	spendUtxo   map[modules.OutPoint]bool
	reqTxMap    map[common.Hash]common.Hash // RequestHash:FullTxHash
}

func newTxList() *txList {
	return &txList{
		txs:         make(map[common.Hash]*linkTx),
		linkTxRoots: make(map[common.Hash]*linkTx),
		newUtxo:     make(map[modules.OutPoint]*modules.Utxo),
		spendUtxo:   make(map[modules.OutPoint]bool),
		reqTxMap:    make(map[common.Hash]common.Hash),
	}
}
func (l *txList) Count() int {
	return len(l.txs)
}

//插入一个Tx,只支持系统合约Request，FullTx，不支持UserContractRequest
func (l *txList) AddTx(tx *txspool.TxPoolTransaction) error {
	//如果是用户合约FullTx，先检查Request是否存在，存在则替换
	if tx.IsUserContractFullTx {
		l.reqTxMap[tx.ReqHash] = tx.TxHash
	}
	txNode := &linkTx{
		Parents:  []*linkTx{},
		Tx:       tx,
		Children: []*linkTx{},
		Number:   0,
	}
	l.txs[tx.TxHash] = txNode
	//根据Tx的依赖Txs，更新Dag结构
	for dependHash := range tx.DependOnTxs {
		if linkTx, ok := l.txs[dependHash]; ok {
			linkTx.Children = append(linkTx.Children, txNode)
			txNode.Parents = append(txNode.Parents, linkTx)
		} else if fullTxHash, ok1 := l.reqTxMap[dependHash]; ok1 {
			if linkTx, ok2 := l.txs[fullTxHash]; ok2 {
				linkTx.Children = append(linkTx.Children, txNode)
				txNode.Parents = append(txNode.Parents, linkTx)
			}
		}
	}
	//取Parent中最大的Number+1为本Node的Number
	for _, p := range txNode.Parents {
		if p.Number+1 > txNode.Number {
			txNode.Number = p.Number + 1
		}
	}
	//如果没有依赖,直接加到Root
	if len(txNode.Parents) == 0 {
		l.linkTxRoots[tx.TxHash] = txNode
	}
	//更新new UTXO好Spend
	for _, o := range tx.Tx.GetSpendOutpoints() {
		l.spendUtxo[*o] = true
	}
	for op, utxo := range tx.Tx.GetNewTxUtxoAndReqUtxos() {
		l.newUtxo[op] = utxo
	}
	return nil
}

//获得交易
func (l *txList) GetTx(hash common.Hash) (*txspool.TxPoolTransaction, error) {
	tx, ok := l.txs[hash]
	if ok {
		return tx.Tx, nil
	}
	return nil, ErrNotFound
}
func (l *txList) GetAllTxs() map[common.Hash]*txspool.TxPoolTransaction {
	result := make(map[common.Hash]*txspool.TxPoolTransaction, len(l.txs))
	for hash, tx := range l.txs {
		result[hash] = tx.Tx
	}
	return result
}

//更新Tx状态为已打包
func (l *txList) UpdateTxStatusPacked(txhash common.Hash, unitHash common.Hash, height uint64) error {
	tx, ok := l.txs[txhash]
	if !ok {
		return ErrNotFound
	}
	//1. 检查前置Txs是否都已经打包
	for _, ptx := range tx.Parents {
		if ptx.Tx.Status != txspool.TxPoolTxStatus_Packed {
			return fmt.Errorf("tx[%s]'s parent tx[%s] not packed yet", txhash.String(), ptx.Tx.TxHash.String())
		}
	}
	//2. 更改状态为已打包
	tx.Tx.Status = txspool.TxPoolTxStatus_Packed
	tx.Tx.UnitHash = unitHash
	tx.Tx.UnitIndex = height
	return nil
}

func (l *txList) UpdateTxStatusUnpacked(hash common.Hash) error {
	tx, ok := l.txs[hash]
	if !ok {
		return ErrNotFound
	}
	//1. 检查后置Txs是否都未打包
	for _, ptx := range tx.Children {
		if ptx.Tx.Status != txspool.TxPoolTxStatus_Unpacked {
			return fmt.Errorf("tx[%s]'s parent tx[%s] still packed", hash.String(), ptx.Tx.TxHash.String())
		}
	}
	//2. 更改状态为未打包
	tx.Tx.Status = txspool.TxPoolTxStatus_Unpacked
	tx.Tx.UnitHash = common.Hash{}
	tx.Tx.UnitIndex = 0
	return nil
}

func (l *txList) DiscardTx(hash common.Hash) error {
	//检查是否存在
	tx, ok := l.txs[hash]
	if !ok {
		return ErrNotFound
	}

	//删除link，如果有必要，更新Root
	err := l.deleteTxLink(tx)
	if err != nil {
		return err
	}
	//删除对应的UTXO
	l.deleteTxUtxo(tx.Tx)
	//删除Map中的值
	delete(l.txs, hash)
	//删除Req：Tx的映射
	if tx.Tx.IsUserContractFullTx {
		delete(l.reqTxMap, tx.Tx.ReqHash)
	}
	return nil
}
func (l *txList) deleteTxUtxo(tx *txspool.TxPoolTransaction) {
	for _, o := range tx.Tx.GetSpendOutpoints() {
		delete(l.spendUtxo, *o)
	}
	for op := range tx.Tx.GetNewTxUtxoAndReqUtxos() {
		delete(l.newUtxo, op)
	}
}
func (l *txList) deleteTxLink(tx *linkTx) error {
	_, ok := l.linkTxRoots[tx.Tx.TxHash]
	if ok {
		delete(l.linkTxRoots, tx.Tx.TxHash)
		for _, child := range tx.Children {
			child.Parents = deleteSliceItem(child.Parents, tx)
			if len(child.Parents) == 0 {
				l.linkTxRoots[child.Tx.TxHash] = child
				child.Parents = nil
			}
		}
		return nil
	}
	return fmt.Errorf("Tx[%s] not a root tx in txpool list", tx.Tx.TxHash.String())
}
func deleteSliceItem(array []*linkTx, tx *linkTx) []*linkTx {
	result := make([]*linkTx, 0, len(array)-1)
	for _, item := range array {
		if item == tx {
			continue
		}
		result = append(result, item)
	}
	return result
}

//用来标记已遍历的
var nodeHashAll map[common.Hash]bool

func (l *txList) GetSortedTxs() ([]*txspool.TxPoolTransaction, error) {
	log.Debug("start GetSortedTxs...")
	nodeHashAll = make(map[common.Hash]bool)
	roots := []*linkTx{}
	for _, tx := range l.linkTxRoots {
		roots = append(roots, tx)
	}
	result := []*txspool.TxPoolTransaction{}
	processor := func(tx *txspool.TxPoolTransaction) (getNext bool, err error) {
		result = append(result, tx)
		return true, nil
	}
	err := l.getSortedTxs(roots, processor)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//广度优先遍历这个图
func (l *txList) getSortedTxs(nodes []*linkTx, processor func(tx *txspool.TxPoolTransaction) (getNext bool, err error)) error {
	if len(nodes) == 0 {
		return nil
	}
	stop := false
	var childrenNodeList []*linkTx
	for _, node := range nodes {
		if _, ok := nodeHashAll[node.Tx.TxHash]; ok { //已经遍历过该节点
			continue
		}
		nodeHashAll[node.Tx.TxHash] = true
		//log.Debugf("process tx[%s]",node.Tx.TxHash.String())
		for _, child := range node.Children {
			if child.Number == node.Number+1 {
				childrenNodeList = append(childrenNodeList, child)
			}
		}
		//log.Debugf("Children count [%d]",len(childrenNodeList))
		if node.Tx.Status == txspool.TxPoolTxStatus_Unpacked {
			getNext, err := processor(node.Tx)
			if !getNext {
				stop = true
				break
			}
			if err != nil {
				stop = true
				break
			}
		}
	}
	if stop {
		return nil
	}
	return l.getSortedTxs(childrenNodeList, processor)
}

func (l *txList) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	if _, ok := l.spendUtxo[*outpoint]; ok {
		return nil, ErrDoubleSpend
	}
	if utxo, ok := l.newUtxo[*outpoint]; ok {
		return utxo, nil
	}
	return nil, ErrNotFound
}

//不排序，直接返回所有查询到的Txs
func (l *txList) GetTxsByStatus(status txspool.TxPoolTxStatus) (map[common.Hash]*txspool.TxPoolTransaction, error) {
	result := make(map[common.Hash]*txspool.TxPoolTransaction)
	for hash, tx := range l.txs {
		if tx.Tx.Status == status {
			result[hash] = tx.Tx
		}
	}
	return result, nil
}
