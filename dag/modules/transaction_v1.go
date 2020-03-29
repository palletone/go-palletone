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
