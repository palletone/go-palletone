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

package modules

import (
	"fmt"
	"io"

	"github.com/palletone/go-palletone/common/rlp"
	vote2 "github.com/palletone/go-palletone/dag/vote"
)

type transactionTemp struct {
	TxMessages []messageTemp
}
type messageTemp struct {
	App  MessageType
	Data []byte
}

func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	txTemp := &transactionTemp{}
	err = rlp.DecodeBytes(raw, txTemp)
	if err != nil {
		return err
	}
	return temp2Tx(txTemp, tx)
}
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	temp, err := tx2Temp(tx)
	if err != nil {
		return err
	}
	return rlp.Encode(w, temp)
}
func tx2Temp(tx *Transaction) (*transactionTemp, error) {
	temp := &transactionTemp{}

	for _, m := range tx.TxMessages {
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
	return temp, nil
}
func temp2Tx(temp *transactionTemp, tx *Transaction) error {
	for _, m := range temp.TxMessages {
		m1 := &Message{
			App: m.App,
		}
		if m.App == APP_PAYMENT {
			var pay PaymentPayload
			err := rlp.DecodeBytes(m.Data, &pay)
			if err != nil {
				return err
			}
			m1.Payload = &pay
		} else if m.App == APP_DATA {
			var text DataPayload
			rlp.DecodeBytes(m.Data, &text)
			m1.Payload = &text
		} else if m.App == APP_CONTRACT_TPL_REQUEST {
			var payload ContractInstallRequestPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_TPL {
			var payload ContractTplPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_DEPLOY_REQUEST {
			var payload ContractDeployRequestPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_DEPLOY {
			var payload ContractDeployPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_INVOKE_REQUEST {
			var payload ContractInvokeRequestPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_INVOKE {
			var payload ContractInvokePayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_STOP_REQUEST {
			var payload ContractStopRequestPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONTRACT_STOP {
			var payload ContractStopPayload
			rlp.DecodeBytes(m.Data, &payload)
			m1.Payload = &payload
		} else if m.App == APP_CONFIG {
			var conf ConfigPayload
			rlp.DecodeBytes(m.Data, &conf)
			m1.Payload = &conf
		} else if m.App == APP_SIGNATURE {
			var sigPayload SignaturePayload
			rlp.DecodeBytes(m.Data, &sigPayload)
			m1.Payload = &sigPayload
		} else if m.App == APP_VOTE {
			var vote vote2.VoteInfo
			rlp.DecodeBytes(m.Data, &vote)
			m1.Payload = &vote
		} else if m.App == OP_MEDIATOR_CREATE {
			var mediatorCreateOp MediatorCreateOperation
			rlp.DecodeBytes(m.Data, &mediatorCreateOp)
			m1.Payload = &mediatorCreateOp
		} else {
			fmt.Println("Unknown message app type:", m.App)
		}
		tx.TxMessages = append(tx.TxMessages, m1)

	}
	return nil

}
