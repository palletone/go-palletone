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
 *  * @date 2018-2019
 *
 */

package modules

const defaultTxInOutAlloc = 0

// Token exchange message and verify message
// App: payment
type PaymentPayload struct {
	Inputs   []*Input  `json:"inputs"`
	Outputs  []*Output `json:"outputs"`
	LockTime uint32    `json:"lock_time"`
}
type Output struct {
	Value    uint64 `json:"value,string"`
	PkScript []byte `json:"pk_script"`
	Asset    *Asset `json:"asset"`
}

type Input struct {
	SignatureScript []byte `json:"signature_script"`
	// if user creating a new asset, this field should be it's config data. Otherwise it is null.
	Extra            []byte    `json:"extra" rlp:"nil"`
	PreviousOutPoint *OutPoint `json:"pre_outpoint"`
}

// NewTxIn returns a new ptn transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *OutPoint, signatureScript []byte) *Input {
	return &Input{
		PreviousOutPoint: prevOut,
		SignatureScript:  signatureScript,
	}
}

// NewTxOut returns a new bitcoin transaction output with the provided
// transaction value and public key script.
func NewTxOut(value uint64, pkScript []byte, asset *Asset) *Output {
	return &Output{
		Value:    value,
		PkScript: pkScript,
		Asset:    asset,
	}
}
func NewPaymentPayload(inputs []*Input, outputs []*Output) *PaymentPayload {
	return &PaymentPayload{
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: defaultTxInOutAlloc,
	}
}
func (pay *PaymentPayload) IsCoinbase() bool {
	if len(pay.Inputs) == 0 {
		return true
	}
	for _, input := range pay.Inputs {
		if input.PreviousOutPoint == nil {
			return true
		}
	}
	return false
}

// AddTxIn adds a transaction input to the message.
func (pld *PaymentPayload) AddTxIn(ti *Input) {
	pld.Inputs = append(pld.Inputs, ti)
}

// AddTxOut adds a transaction output to the message.
func (pld *PaymentPayload) AddTxOut(to *Output) {
	pld.Outputs = append(pld.Outputs, to)
}
