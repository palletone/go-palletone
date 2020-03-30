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
 * @date 2018-2010
 */

package modules

import (
	"github.com/ethereum/go-ethereum/rlp"
)

type transactionTempV0 struct {
	TxMessages []messageTemp
	CertId     []byte // should be big.Int byte
	Illegal    bool   // not hash, 1:no valid, 0:ok
}

func tx2TempV0(tx *Transaction) (*transactionTempV0, error) {
	temp := transactionTempV0{}
	temp.Illegal = tx.Illegal()
	temp.CertId = tx.CertId()

	for _, m := range tx.Messages() {
		m1 := messageTemp{
			App: m.App,
		}
		d, err := rlp.EncodeToBytes(m.Payload)
		if err != nil {
			return nil, err
		}
		m1.Data = d

		temp.TxMessages = append(temp.TxMessages, m1)

	}
	return &temp, nil
}
func tempV02Tx(temp *transactionTempV0, tx *Transaction) error {
	d := transaction_sdw{}
	msgs, err := convertTxMessages(temp.TxMessages)
	if err != nil {
		return err
	}
	d.TxMessages = msgs
	d.Illegal = temp.Illegal
	d.CertId = temp.CertId
	// init tx
	tx.txdata = d
	return nil

}
