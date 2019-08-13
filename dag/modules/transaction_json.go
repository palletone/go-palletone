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
 *  * @date 2018
 *
 */

package modules

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/palletone/go-palletone/dag/errors"
)

type idxPaymentPayload struct {
	Index int
	*PaymentPayload
}

//type idxConfigPayload struct {
//	Index int
//	*ConfigPayload
//}

type idxSignaturePayload struct {
	Index int
	*SignaturePayload
}

type idxTextPayload struct {
	Index int
	*DataPayload
}

//type idxMediatorCreateOperation struct {
//	Index int
//	*MediatorCreateOperation
//}

type idxAccountUpdateOperation struct {
	Index int
	*AccountStateUpdatePayload
}

//install
type idxContractInstallRequestPayload struct {
	Index int
	*ContractInstallRequestPayload
}

type idxContractTplPayload struct {
	Index int
	*ContractTplPayload
}

//deploy
type idxContractDeployRequestPayload struct {
	Index int
	*ContractDeployRequestPayload
}

type idxContractDeployPayload struct {
	Index int
	*ContractDeployPayload
}

//invoke
type idxContractInvokeRequestPayload struct {
	Index int
	*ContractInvokeRequestPayload
}

type idxContractInvokePayload struct {
	Index int
	*ContractInvokePayload
}

//stop
type idxContractStopRequestPayload struct {
	Index int
	*ContractStopRequestPayload
}

type idxContractStopPayload struct {
	Index int
	*ContractStopPayload
}

type txJsonTemp struct {
	MsgCount int
	CertId   string
	Illegal  bool
	Payment  []*idxPaymentPayload
	//Config                  []*idxConfigPayload
	Text []*idxTextPayload
	//MediatorCreateOperation []*idxMediatorCreateOperation
	AccountUpdateOperation []*idxAccountUpdateOperation
	Signature              []*idxSignaturePayload

	ContractInstallRequest []*idxContractInstallRequestPayload
	ContractDeployRequest  []*idxContractDeployRequestPayload
	ContractInvokeRequest  []*idxContractInvokeRequestPayload
	ContractStopRequest    []*idxContractStopRequestPayload

	ContractTpl    []*idxContractTplPayload
	ContractDeploy []*idxContractDeployPayload
	ContractInvoke []*idxContractInvokePayload
	ContractStop   []*idxContractStopPayload
}

func tx2JsonTemp(tx *Transaction) (*txJsonTemp, error) {
	intCertID := new(big.Int).SetBytes(tx.CertId)
	temp := &txJsonTemp{MsgCount: len(tx.TxMessages), CertId: intCertID.String(), Illegal: tx.Illegal}
	for idx, msg := range tx.TxMessages {
		if msg.App == APP_PAYMENT {
			temp.Payment = append(temp.Payment, &idxPaymentPayload{
				Index: idx, PaymentPayload: msg.Payload.(*PaymentPayload)})
		} else if msg.App == APP_CONTRACT_INVOKE {
			temp.ContractInvoke = append(temp.ContractInvoke, &idxContractInvokePayload{
				Index: idx, ContractInvokePayload: msg.Payload.(*ContractInvokePayload)})
		} else if msg.App == APP_CONTRACT_TPL {
			temp.ContractTpl = append(temp.ContractTpl, &idxContractTplPayload{
				Index: idx, ContractTplPayload: msg.Payload.(*ContractTplPayload)})
		} else if msg.App == APP_CONTRACT_DEPLOY {
			temp.ContractDeploy = append(temp.ContractDeploy, &idxContractDeployPayload{
				Index: idx, ContractDeployPayload: msg.Payload.(*ContractDeployPayload)})
		} else if msg.App == APP_CONTRACT_STOP {
			temp.ContractStop = append(temp.ContractStop, &idxContractStopPayload{
				Index: idx, ContractStopPayload: msg.Payload.(*ContractStopPayload)})
		} else if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			temp.ContractInvokeRequest = append(temp.ContractInvokeRequest,
				&idxContractInvokeRequestPayload{
					Index:                        idx,
					ContractInvokeRequestPayload: msg.Payload.(*ContractInvokeRequestPayload),
				})
		} else if msg.App == APP_CONTRACT_TPL_REQUEST {
			temp.ContractInstallRequest = append(temp.ContractInstallRequest,
				&idxContractInstallRequestPayload{
					Index:                         idx,
					ContractInstallRequestPayload: msg.Payload.(*ContractInstallRequestPayload),
				})
		} else if msg.App == APP_CONTRACT_DEPLOY_REQUEST {
			temp.ContractDeployRequest = append(temp.ContractDeployRequest,
				&idxContractDeployRequestPayload{
					Index:                        idx,
					ContractDeployRequestPayload: msg.Payload.(*ContractDeployRequestPayload),
				})
		} else if msg.App == APP_CONTRACT_STOP_REQUEST {
			temp.ContractStopRequest = append(temp.ContractStopRequest,
				&idxContractStopRequestPayload{
					Index:                      idx,
					ContractStopRequestPayload: msg.Payload.(*ContractStopRequestPayload),
				})
		} else if msg.App == APP_DATA {
			temp.Text = append(temp.Text, &idxTextPayload{Index: idx, DataPayload: msg.Payload.(*DataPayload)})
		} else if msg.App == APP_SIGNATURE {
			temp.Signature = append(temp.Signature, &idxSignaturePayload{
				Index: idx, SignaturePayload: msg.Payload.(*SignaturePayload)})

		} else if msg.App == APP_ACCOUNT_UPDATE {
			temp.AccountUpdateOperation = append(temp.AccountUpdateOperation,
				&idxAccountUpdateOperation{Index: idx, AccountStateUpdatePayload: msg.Payload.(*AccountStateUpdatePayload)})
		} else {
			return nil, errors.New("Unsupport APP" + strconv.Itoa(int(msg.App)) + " please edit transaction_json.go")
		}
	}
	return temp, nil
}

func jsonTemp2tx(tx *Transaction, temp *txJsonTemp) error {
	if len(temp.CertId) > 0 {
		intCertID, _ := new(big.Int).SetString(temp.CertId, 10)
		if intCertID == nil {
			return fmt.Errorf("certid is invalid")
		}
		tx.CertId = intCertID.Bytes()
	}
	tx.Illegal = temp.Illegal
	tx.TxMessages = make([]*Message, temp.MsgCount)
	processed := 0
	for _, p := range temp.Payment {
		tx.TxMessages[p.Index] = NewMessage(APP_PAYMENT, p.PaymentPayload)
		processed++
	}
	//request
	for _, p := range temp.ContractInstallRequest {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_TPL_REQUEST, p.ContractInstallRequestPayload)
		processed++
	}
	for _, p := range temp.ContractDeployRequest {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_DEPLOY_REQUEST, p.ContractDeployRequestPayload)
		processed++
	}
	for _, p := range temp.ContractInvokeRequest {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_INVOKE_REQUEST, p.ContractInvokeRequestPayload)
		processed++
	}
	for _, p := range temp.ContractStopRequest {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_STOP_REQUEST, p.ContractStopRequestPayload)
		processed++
	}

	//content
	for _, p := range temp.ContractTpl {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_TPL, p.ContractTplPayload)
		processed++
	}
	for _, p := range temp.ContractDeploy {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_DEPLOY, p.ContractDeployPayload)
		processed++
	}
	for _, p := range temp.ContractInvoke {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_INVOKE, p.ContractInvokePayload)
		processed++
	}
	for _, p := range temp.ContractStop {
		tx.TxMessages[p.Index] = NewMessage(APP_CONTRACT_STOP, p.ContractStopPayload)
		processed++
	}

	for _, p := range temp.Text {
		tx.TxMessages[p.Index] = NewMessage(APP_DATA, p.DataPayload)
		processed++
	}
	for _, p := range temp.Signature {
		tx.TxMessages[p.Index] = NewMessage(APP_SIGNATURE, p.SignaturePayload)
		processed++
	}
	for _, p := range temp.AccountUpdateOperation {
		tx.TxMessages[p.Index] = NewMessage(APP_ACCOUNT_UPDATE, p.AccountStateUpdatePayload)
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
	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	return jsonTemp2tx(tx, temp)
}
