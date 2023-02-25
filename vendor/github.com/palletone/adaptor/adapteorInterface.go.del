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
package adaptor

type adapter interface {
	NewPrivateKey() (priKey string)
	GetPublicKey(priKey string) (pubKey string)
	GetAddress(priKey string) (address string)
	CreateMultiSigAddress(params string) string

	GenTransaction(params string) string
	ContractTransaction(params string) string
	DecodeTransaction(params string) string

	SignTransaction(params string) string
	SendTransaction(params string) string

	Query(params string) string
	ContractQuery(params string) string
}

type adapterCryptoCurrency interface {
	NewPrivateKey() (wifPriKey string)
	GetPublicKey(wifPriKey string) (pubKey string)
	GetAddress(wifPriKey string) (address string)
	CreateMultiSigAddress(params string) string

	GetUnspendUTXO(params string) string //

	RawTransactionGen(params string) string
	DecodeRawTransaction(params string) string
	SignTransaction(params string) string

	GetBalance(params string) string      //
	GetTransactions(params string) string //
	ImportMultisig(params string) string  //

	SendTransaction(params string) string
}

type adapterSmartContract interface {
	NewPrivateKey() (priKey string)
	GetPublicKey(priKey string) (pubKey string)
	GetAddress(priKey string) (address string)

	CreateMultiSigAddress(params string) string

	GetBalance(params string) string
	GetTransactionByHash(params string) string
	ContractDeployment(params string) string

	GetMultisigInfo(params string) string
	CalculateSig(params string) string
}
