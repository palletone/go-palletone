/* This file is part of go-palletone.
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

   @author PalletOne core developers <dev@pallet.one>
   @date 2018
*/

package modules

type MessageType byte

const (
	APP_PAYMENT MessageType = iota
	APP_CONTRACT_TPL
	APP_CONTRACT_DEPLOY
	APP_CONTRACT_INVOKE
	APP_CONFIG
	APP_TEXT
	APP_VOTE
	OP_MEDIATOR_CREATE
)

// key: message.UnitHash(message+timestamp)
type Message struct {
	App     MessageType `json:"app"`     // message type
	Payload interface{} `json:"payload"` // the true transaction data
}

// return message struct
func NewMessage(app MessageType, payload interface{}) *Message {
	m := new(Message)
	m.App = app
	m.Payload = payload
	return m
}

func (msg *Message) CopyMessages(cpyMsg *Message) *Message {
	msg.App = cpyMsg.App
	msg.Payload = cpyMsg.Payload
	switch cpyMsg.App {
	case APP_PAYMENT, APP_CONTRACT_TPL, APP_TEXT, APP_VOTE:
		msg.Payload = cpyMsg.Payload
	case APP_CONFIG:
		payload, _ := cpyMsg.Payload.(*ConfigPayload)
		newPayload := ConfigPayload{
			ConfigSet: []PayloadMapStruct{},
		}
		for _, p := range payload.ConfigSet {
			newPayload.ConfigSet = append(newPayload.ConfigSet, PayloadMapStruct{Key: p.Key, Value: p.Value})
		}
		msg.Payload = newPayload
	case APP_CONTRACT_DEPLOY:
		payload, _ := cpyMsg.Payload.(*ContractDeployPayload)
		newPayload := ContractDeployPayload{
			TemplateId:   payload.TemplateId,
			ContractId:   payload.ContractId,
			Args:         payload.Args,
			Excutiontime: payload.Excutiontime,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Version: &StateVersion{Height: rs.Version.Height, TxIndex: rs.Version.TxIndex}})
		}
		writeSet := []PayloadMapStruct{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	case APP_CONTRACT_INVOKE:
		payload, _ := cpyMsg.Payload.(*ContractInvokePayload)
		newPayload := ContractInvokePayload{
			ContractId:   payload.ContractId,
			Args:         payload.Args,
			Excutiontime: payload.Excutiontime,
		}
		readSet := []ContractReadSet{}
		for _, rs := range payload.ReadSet {
			readSet = append(readSet, ContractReadSet{Key: rs.Key, Version: &StateVersion{Height: rs.Version.Height, TxIndex: rs.Version.TxIndex}})
		}
		writeSet := []PayloadMapStruct{}
		for _, ws := range payload.WriteSet {
			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
		}
		newPayload.ReadSet = readSet
		newPayload.WriteSet = writeSet
		msg.Payload = newPayload
	}
	return msg
}
