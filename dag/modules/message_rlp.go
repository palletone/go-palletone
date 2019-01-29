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

//
//import (
//	"github.com/ethereum/go-ethereum/rlp"
//	"io"
//	"time"
//)
//
//type ContractDeployRequestPayloadRlp struct {
//	TplId   []byte   `json:"tpl_name"`
//	TxId    string   `json:"transaction_id"`
//	Args    [][]byte `json:"args"`
//	Timeout uint32   `json:"timeout"`
//}
//
//func (input *ContractDeployRequestPayload) DecodeRLP(s *rlp.Stream) error {
//	raw, err := s.Raw()
//	if err != nil {
//		return err
//	}
//	temp := &ContractDeployRequestPayloadRlp{}
//	err = rlp.DecodeBytes(raw, temp)
//	if err != nil {
//		return err
//	}
//
//	input.TplId = temp.TplId
//	input.TxId = temp.TxId
//	input.Args = temp.Args
//	input.Timeout = time.Duration(temp.Timeout)
//	return nil
//}
//func (input *ContractDeployRequestPayload) EncodeRLP(w io.Writer) error {
//	temp := &ContractDeployRequestPayloadRlp{}
//	temp.TplId = input.TplId
//	temp.TxId = input.TxId
//	temp.Args = input.Args
//	temp.Timeout = uint32(input.Timeout)
//	return rlp.Encode(w, temp)
//}
