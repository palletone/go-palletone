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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package rwset

type RwSetTxMgr struct {
	//db                    DB
	//rwLock            	sync.RWMutex
	name                    string
	baseTxSim               map[string]TxSimulator
}

func NewRwSetMgr(name string) (*RwSetTxMgr, error) {
	return &RwSetTxMgr{name, make(map[string] TxSimulator)}, nil
}

// NewTxSimulator implements method in interface `txmgmt.TxMgr`
func  (m *RwSetTxMgr)NewTxSimulator(chainid string, txid string) (TxSimulator, error) {
	logger.Debugf("constructing new tx simulator")

	//for k, _ := range m.baseTxSim{
	//	if k == chainid {
	//		return m.baseTxSim[chainid], nil
	//	}
	//}
	if _, ok := m.baseTxSim[chainid]; ok{
		logger.Infof("chainid[%s] already exit")
		return m.baseTxSim[chainid], nil
	}

	t, err := newBasedTxSimulator(txid)
	if err != nil {
		return nil, err
	}
	m.baseTxSim[chainid] = t
	logger.Infof("creat new rwSetTx")

	return t, nil
}

func init() {

}
