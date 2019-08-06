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
package adaptoreth

import (
	"github.com/palletone/adaptor"
)

type RPCParams struct {
	Rawurl string `json:"rawurl"`
}

type AdaptorETH struct {
	NetID int
	RPCParams
}

const (
	NETID_MAIN = iota
	NETID_TEST
)

func (aeth AdaptorETH) NewPrivateKey() (prikeyHex string) {
	return NewPrivateKey(aeth.NetID)
}
func (aeth AdaptorETH) GetPublicKey(prikeyHex string) (pubKey string) {
	return GetPublicKey(prikeyHex, aeth.NetID)
}
func (aeth AdaptorETH) GetAddress(prikeyHex string) (address string) {
	return GetAddress(prikeyHex, aeth.NetID)
}

func (aeth AdaptorETH) GetBalance(params string) string {
	return GetBalance(params, &aeth.RPCParams, aeth.NetID)
}
func (aeth AdaptorETH) GetTransactionByHash(params *adaptor.GetTransactionParams) (string, error) {
	return GetTransactionByHash(params, &aeth.RPCParams, aeth.NetID)
}
func (aeth AdaptorETH) GetErc20TxByHash(params *adaptor.GetErc20TxByHashParams) (string, error) {
	return GetErc20TxByHash(params, &aeth.RPCParams, aeth.NetID)
}
func (aeth AdaptorETH) CreateMultiSigAddress(params *adaptor.CreateMultiSigAddressParams) (string, error) {
	return CreateMultiSigAddress(params)
}

func (aeth AdaptorETH) CalculateSig(params string) string {
	return CalculateSig(params)
}
func (aeth AdaptorETH) Keccak256HashPackedSig(params *adaptor.Keccak256HashPackedSigParams) (string, error) {
	return Keccak256HashPackedSig(params)
}
func (aeth AdaptorETH) RecoverAddr(params *adaptor.RecoverParams) (string, error) {
	return RecoverAddr(params)
}

func (aeth AdaptorETH) SignTransaction(params *adaptor.ETHSignTransactionParams) (string, error) {
	return SignTransaction(params)
}
func (aeth AdaptorETH) SendTransaction(params *adaptor.SendTransactionParams) (string, error) {
	return SendTransaction(params, &aeth.RPCParams, aeth.NetID)
}

func (aeth AdaptorETH) QueryContract(params *adaptor.QueryContractParams) (string, error) {
	return QueryContract(params, &aeth.RPCParams, aeth.NetID)
}
func (aeth AdaptorETH) GenInvokeContractTX(params *adaptor.GenInvokeContractTXParams) (string, error) {
	return GenInvokeContractTX(params, &aeth.RPCParams, aeth.NetID)
}
func (aeth AdaptorETH) GenDeployContractTX(params *adaptor.GenDeployContractTXParams) (string, error) {
	return GenDeployContractTX(params, &aeth.RPCParams, aeth.NetID)
}

func (aeth AdaptorETH) GetEventByAddress(params *adaptor.GetEventByAddressParams) (string, error) {
	return GetEventByAddress(params, &aeth.RPCParams, aeth.NetID)
}

func (aeth AdaptorETH) GetBestHeader(params *adaptor.GetBestHeaderParams) (string, error) {
	return GetBestHeader(params, &aeth.RPCParams, aeth.NetID)
}
