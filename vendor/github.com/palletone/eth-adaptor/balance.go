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
	"context"
	"encoding/json"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/ethclient"

	"github.com/palletone/eth-adaptor/ethclient"
)

func GetClient(rpcParams *RPCParams) (*ethclient.Client, error) {
	client, err := ethclient.Dial(rpcParams.Rawurl)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type GetBalanceParams struct {
	Account string `json:"account"`
}
type GetBalanceResult struct {
	Balance float64 `json:"balance"`
}

func GetBalance(params string, rpcParams *RPCParams, netID int) string {
	//convert params from json format
	var getBalanceParams GetBalanceParams
	err := json.Unmarshal([]byte(params), &getBalanceParams)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}

	//call eth rpc method
	account := common.HexToAddress(getBalanceParams.Account)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}
	//	fmt.Println("balance : ", balance)

	//remove e+18
	bigFloat := new(big.Float)
	bigFloat.SetInt(balance)
	bigFloat.Mul(bigFloat, big.NewFloat(1e-18))
	strFloat := bigFloat.String()
	//fmt.Println(strFloat)

	//convert balance
	var result GetBalanceResult
	result.Balance, _ = strconv.ParseFloat(strFloat, 8)

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}

	return string(jsonResult)
}
