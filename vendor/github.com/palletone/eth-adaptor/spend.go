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

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/palletone/adaptor"
)

func SignTransaction(signTransactionParams *adaptor.ETHSignTransactionParams) (string, error) {
	rlpTx, err := hexutil.Decode(signTransactionParams.TransactionHex)
	if err != nil {
		return "", err
	}

	var tx types.Transaction
	err = rlp.DecodeBytes(rlpTx, &tx)
	if err != nil {
		return "", err
	}
	//fmt.Printf("tx hash : 0x%x\n\n", tx.Hash())

	//hex private key to ecdsa private key

	if "0x" == signTransactionParams.PrivateKeyHex[0:2] {
		signTransactionParams.PrivateKeyHex = signTransactionParams.PrivateKeyHex[2:]
	}
	priKey, err := crypto.HexToECDSA(signTransactionParams.PrivateKeyHex)
	if err != nil {
		return "", err
	}

	//
	signedTx, err := types.SignTx(&tx, types.HomesteadSigner{}, priKey)
	if err != nil {
		return "", err
	}
	//fmt.Printf("signedTx hash : 0x%x\n\n", signedTx.Hash())

	//
	rlpTXBytes, err := rlp.EncodeToBytes(signedTx)

	//save result
	var result adaptor.ETHSignTransactionResult
	result.TransactionHex = hexutil.Encode(rlpTXBytes)

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func SendTransaction(sendTransactionParams *adaptor.SendTransactionParams, rpcParams *RPCParams, netID int) (string, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return "", err
	}

	rlpTx, err := hexutil.Decode(sendTransactionParams.TransactionHex)
	if err != nil {
		return "", err
	}

	var tx types.Transaction
	err = rlp.DecodeBytes(rlpTx, &tx)
	if err != nil {
		return "", err
	}

	//
	err = client.SendTransaction(context.Background(), &tx)
	if err != nil {
		return "", err
	}

	//save result
	var result adaptor.SendTransactionResult
	result.TransactionHah = tx.Hash().Hex()

	//
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil

}
