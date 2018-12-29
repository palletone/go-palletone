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

package modules

import (
	"encoding/json"
	"github.com/palletone/go-palletone/dag/errors"
	"strconv"
)

type idxPaymentPayload struct {
	Index int
	*PaymentPayload
}
type idxContractInvokeRequestPayload struct {
	Index int
	*ContractInvokeRequestPayload
}
type idxContractInvokePayload struct {
	Index int
	*ContractInvokePayload
}
type idxConfigPayload struct {
	Index int
	*ConfigPayload
}
type idxSignaturePayload struct {
	Index int
	*SignaturePayload
}
type idxTextPayload struct {
	Index int
	*TextPayload
}
type idxMediatorCreateOperation struct {
	Index int
	*MediatorCreateOperation
}
type idxContractTplPayload struct {
	Index int
	*ContractTplPayload
}
type idxContractDeployPayload struct {
	Index int
	*ContractDeployPayload
}
type idxContractInstallRequestPayload struct {
	Index int
	*ContractInstallRequestPayload
}
type idxContractStopPayload struct {
	Index int
	*ContractStopPayload
}
type idxContractStopRequestPayload struct {
	Index int
	*ContractStopRequestPayload
}
type txJsonTemp struct {
	MsgCount                int
	Payment                 []*idxPaymentPayload
	ContractTpl             []*idxContractTplPayload
	ContractDeploy          []*idxContractDeployPayload
	ContractInvoke          []*idxContractInvokePayload
	ContractStop            []*idxContractStopPayload
	ContractInvokeRequest   []*idxContractInvokeRequestPayload
	ContractInstallRequest  []*idxContractInstallRequestPayload
	ContractStopRequest     []*idxContractStopRequestPayload
	Config                  []*idxConfigPayload
	Signature               []*idxSignaturePayload
	Text                    []*idxTextPayload
	MediatorCreateOperation []*idxMediatorCreateOperation
}

func tx2JsonTemp(tx *Transaction) (*txJsonTemp, error) {
	temp := &txJsonTemp{MsgCount: len(tx.TxMessages)}
	for idx, msg := range tx.TxMessages {
		if msg.App == APP_PAYMENT {
			temp.Payment = append(temp.Payment, &idxPaymentPayload{Index: idx, PaymentPayload: msg.Payload.(*PaymentPayload)})
		} else if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			temp.ContractInvokeRequest = append(temp.ContractInvokeRequest,
				&idxContractInvokeRequestPayload{
					Index: idx,
					ContractInvokeRequestPayload: msg.Payload.(*ContractInvokeRequestPayload),
				})
		} else if msg.App == APP_CONTRACT_INVOKE {
			temp.ContractInvoke = append(temp.ContractInvoke, &idxContractInvokePayload{Index: idx, ContractInvokePayload: msg.Payload.(*ContractInvokePayload)})
		} else if msg.App == APP_TEXT {
			temp.Text = append(temp.Text, &idxTextPayload{Index: idx, TextPayload: msg.Payload.(*TextPayload)})
		} else if msg.App == APP_SIGNATURE {
			temp.Signature = append(temp.Signature, &idxSignaturePayload{Index: idx, SignaturePayload: msg.Payload.(*SignaturePayload)})
		} else if msg.App == APP_CONFIG {
			temp.Config = append(temp.Config, &idxConfigPayload{Index: idx, ConfigPayload: msg.Payload.(*ConfigPayload)})
		} else if msg.App == OP_MEDIATOR_CREATE {
			temp.MediatorCreateOperation = append(temp.MediatorCreateOperation, &idxMediatorCreateOperation{Index: idx, MediatorCreateOperation: msg.Payload.(*MediatorCreateOperation)})
		} else {
			return nil, errors.New("Unsupport APP" + strconv.Itoa(int(msg.App)) + " please edit transaction_json.go")
		}
	}
	return temp, nil
}

func jsonTemp2tx(tx *Transaction, temp *txJsonTemp) error {
	tx.TxMessages = make([]*Message, temp.MsgCount)
	processed := 0
	for _, p := range temp.Payment {
		tx.TxMessages[p.Index] = NewMessage(APP_PAYMENT, p.PaymentPayload)
		processed++
	}
	for _, p := range temp.ContractInvokeRequest {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_INVOKE_REQUEST, p.ContractInvokeRequestPayload)
		processed++
	}
	for _, p := range temp.ContractInvoke {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_INVOKE, p.ContractInvokePayload)
		processed++
	}
	for _, p := range temp.Text {
		tx.TxMessages[p.Index] = NewMessage(APP_TEXT, p.TextPayload)
		processed++
	}
	for _, p := range temp.Signature {
		tx.TxMessages[p.Index] = NewMessage(APP_SIGNATURE, p.SignaturePayload)
		processed++
	}
	for _, p := range temp.Config {
		tx.TxMessages[p.Index] = NewMessage(APP_CONFIG, p.ConfigPayload)
		processed++
	}
	for _, p := range temp.MediatorCreateOperation {
		tx.TxMessages[p.Index] = NewMessage(OP_MEDIATOR_CREATE, p.MediatorCreateOperation)
		processed++
	}
	if processed < temp.MsgCount {
		return errors.New("Some message don't process in transaction_json.go")
	}
	return nil
}
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	temp, err := tx2JsonTemp(tx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(temp)
}
func (tx *Transaction) UnmarshalJSON(data []byte) error {
	temp := &txJsonTemp{}
	err := json.Unmarshal([]byte(data), temp)
	if err != nil {
		return err
	}
	return jsonTemp2tx(tx, temp)
}
