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

package modules

import "github.com/ethereum/go-ethereum/rlp"

func (tx *Transaction) Nonce() uint64   { return tx.txdata.AccountNonce }
func (tx *Transaction) Version() uint32 { return tx.txdata.Version }
func (tx *Transaction) SetNonce(nonce uint64) {
	tx.txdata.AccountNonce = nonce
	tx.resetCache()
}
func (tx *Transaction) SetVersion(v uint32) {
	tx.txdata.Version = v
	tx.resetCache()
}

type transactionTempV1 struct {
	Version    uint32
	TxMessages []messageTemp
	TxExtra    []byte
}
type messageTemp struct {
	App  MessageType
	Data []byte
}
type txExtra struct {
	AccountNonce uint64
	CertId       []byte // should be big.Int byte
	Illegal      bool   // not hash, 1:no valid, 0:ok
}

func tx2TempV1(tx *Transaction) (*transactionTempV1, error) {
	temp := transactionTempV1{}
	txExtra := txExtra{}
	txExtra.AccountNonce = tx.Nonce()
	txExtra.Illegal = tx.Illegal()
	txExtra.CertId = tx.CertId()
	txExtraB, err := rlp.EncodeToBytes(txExtra)
	if err != nil {
		return nil, err
	}
	temp.TxExtra = txExtraB
	temp.Version = tx.Version()
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
func tempV12Tx(temp *transactionTempV1, tx *Transaction) error {
	d := transaction_sdw{}
	msgs, err := convertTxMessages(temp.TxMessages)
	if err != nil {
		return err
	}
	d.TxMessages = msgs
	d.Version = temp.Version
	txExtra := new(txExtra)
	err = rlp.DecodeBytes(temp.TxExtra, &txExtra)
	if err != nil {
		return err
	}
	d.AccountNonce = txExtra.AccountNonce
	d.Illegal = txExtra.Illegal
	d.CertId = txExtra.CertId
	tx.txdata = d
	return nil
}
