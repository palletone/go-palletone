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
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package core

import (
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"sort"
)

type UtxoInterface interface {
	GetAmount() uint64
}

type Utxos []UtxoInterface

func (c Utxos) Len() int {
	return len(c)
}
func (c Utxos) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c Utxos) Less(i, j int) bool {
	return c[i].GetAmount() < c[j].GetAmount()
}

func find_min(utxos []UtxoInterface) UtxoInterface {
	amout := utxos[0].GetAmount()
	min_utxo := utxos[0]
	for _, utxo := range utxos {
		if utxo.GetAmount() < amout {
			min_utxo = utxo
			amout = min_utxo.GetAmount()
		}
	}
	return min_utxo
}

func Merge_Utxos(utxos Utxos, poolutxos Utxos) (Utxos, error) {
	return nil, nil
}

func Select_utxo_Greedy(utxos Utxos, amount uint64) (Utxos, uint64, error) {
	var greaters Utxos
	var lessers Utxos
	var taken_lutxo Utxos
	var taken_gutxo Utxos
	var accum uint64
	var change uint64
	logPickedAmt := ""
	accum = 0
	for _, utxo := range utxos {
		if utxo.GetAmount() >= amount {
			greaters = append(greaters, utxo)
		}
		if utxo.GetAmount() < amount {
			lessers = append(lessers, utxo)
		}
	}
	if len(lessers) > 0 {

		sort.Sort(lessers)
		for _, utxo := range lessers {
			accum += utxo.GetAmount()
			logPickedAmt += fmt.Sprintf("%d,", utxo.GetAmount())
			taken_lutxo = append(taken_lutxo, utxo)
			if accum >= amount {
				change = accum - amount
				log.Debugf("Pickup count[%d] utxos, each amount:%s to match wanted amount:%d", len(taken_lutxo), logPickedAmt, amount)
				return taken_lutxo, change, nil
			}
		}
	}
	if accum < amount && len(greaters) == 0 {
		return nil, 0, errors.New("Amount Not Enough to pay")
	}

	min_greater := find_min(greaters)
	change = min_greater.GetAmount() - amount
	logPickedAmt = fmt.Sprintf("%d,", min_greater.GetAmount())
	taken_gutxo = append(taken_gutxo, min_greater)

	log.Debugf("Pickup count[%d] utxos, each amount:%s to match wanted amount:%d", len(taken_gutxo), logPickedAmt, amount)
	return taken_gutxo, change, nil
}
