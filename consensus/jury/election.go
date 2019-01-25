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
package jury

import (
	"github.com/palletone/go-palletone/common"
)

func GetContractJurors(num int, data []byte) ([]common.Address, error) {
	return nil, nil
}

func (p *Processor) ProcessElectionRequestEvent(event *ElectionEvent) (result *ElectionEvent, err error) {
	//产生vrf证明
	//计算二项式分步，确定自己是否选中
	//如果选中，则对请求结果返回

	//pi, err := vrf.ECVRF_prove(pk, sk, msg[:])
	//if err != nil {
	//
	//}
	//
	//algorithm.Selected()

	return nil, nil
}

func (p *Processor) ProcessElectionResultEvent(event *ElectionEvent) error {

	return nil
}
