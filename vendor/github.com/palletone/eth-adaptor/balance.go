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
package ethadaptor

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/palletone/adaptor"

	"github.com/palletone/eth-adaptor/ethclient"
)

func GetClient(rpcParams *RPCParams) (*ethclient.Client, error) {
	client, err := ethclient.Dial(rpcParams.Rawurl)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetBalanceETH(input *adaptor.GetBalanceInput, rpcParams *RPCParams) (*adaptor.GetBalanceOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	//call eth rpc method
	account := common.HexToAddress(input.Address)

	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, err
	}
	//	fmt.Println("balance : ", balance)

	//convert balance
	var result adaptor.GetBalanceOutput
	result.Balance.Amount = balance
	result.Balance.Asset = input.Asset
	return &result, nil
}
