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
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/palletone/adaptor"
)

func writeBytes(buf io.Writer, appendBytes []byte) {
	lenBytes := len(appendBytes)
	if lenBytes == 32 {
		buf.Write(appendBytes)
	} else {
		zeroBytes := make([]byte, 32-lenBytes)
		buf.Write(zeroBytes)
		buf.Write(appendBytes)
	}
}

func SignTransaction(input *adaptor.SignTransactionInput) (*adaptor.SignTransactionOutput, error) {
	var tx types.Transaction
	err := rlp.DecodeBytes(input.Transaction, &tx)
	if err != nil {
		return nil, err
	}

	priKey, err := crypto.ToECDSA(input.PrivateKey)
	if err != nil {
		return nil, err
	}

	//
	signedTx, err := types.SignTx(&tx, types.HomesteadSigner{}, priKey)
	if err != nil {
		return nil, err
	}
	//
	rlpTXBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, err
	}

	//signedTx.WithSignature()
	v, r, s := signedTx.RawSignatureValues()
	var buf bytes.Buffer
	writeBytes(&buf, r.Bytes())
	writeBytes(&buf, s.Bytes())
	buf.WriteByte(v.Bytes()[0] - 27)

	//save result
	var result adaptor.SignTransactionOutput
	result.Signature = buf.Bytes()
	result.Extra = rlpTXBytes

	return &result, nil
}

func BindETHTxAndSignature(input *adaptor.BindTxAndSignatureInput) (*adaptor.BindTxAndSignatureOutput, error) {
	var tx types.Transaction
	err := rlp.DecodeBytes(input.Transaction, &tx)
	if err != nil {
		return nil, err
	}

	signedTx, err := tx.WithSignature(types.HomesteadSigner{}, input.Signs[0])
	if err != nil {
		return nil, err
	}

	rlpTXBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, err
	}

	//save result
	var result adaptor.BindTxAndSignatureOutput
	result.SignedTx = rlpTXBytes

	return &result, nil
}

func BindTxAndSignature(input *adaptor.BindTxAndSignatureInput) (*adaptor.BindTxAndSignatureOutput, error) {
	data := make([]byte, 0)
	//method ID, example, 0xa9059cbb withdraw(address,uint,string,bytes,bytes,bytes)
	hash := crypto.Keccak256Hash(input.Extra)
	data = append(data, hash[:4]...)
	//receiver amount token extra
	data = append(data, input.Transaction[33:]...) //m+from=33
	//signatures
	for i := range input.Signs {
		sigPadded := common.LeftPadBytes(input.Signs[i], 32)
		data = append(data, sigPadded...)
	}

	//save result
	var result adaptor.BindTxAndSignatureOutput
	result.SignedTx = data

	return &result, nil
}

func CalcTxHash(input *adaptor.CalcTxHashInput) (*adaptor.CalcTxHashOutput, error) {
	hash := crypto.Keccak256Hash(input.Transaction)

	//save result
	var result adaptor.CalcTxHashOutput
	result.Hash = hash.Bytes()

	return &result, nil
}

func SendTransaction(input *adaptor.SendTransactionInput, rpcParams *RPCParams) (
	*adaptor.SendTransactionOutput, error) {
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	var tx types.Transaction
	err = rlp.DecodeBytes(input.Transaction, &tx)
	if err != nil {
		return nil, err
	}

	//
	err = client.SendTransaction(context.Background(), &tx)
	if err != nil {
		//fmt.Println("client.SendTransaction failed:", err)
		return nil, err
	}

	//save result
	var result adaptor.SendTransactionOutput
	result.TxID = tx.Hash().Bytes()

	return &result, nil

}

func CreateETHTx(input *adaptor.CreateTransferTokenTxInput, rpcParams *RPCParams) (
	*adaptor.CreateTransferTokenTxOutput, error) {
	if input.Amount == nil {
		return nil, errors.New("input's Amount is nil")
	}
	if input.Fee == nil {
		return nil, errors.New("input's Fee is nil")
	}
	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}

	fromAddress := common.HexToAddress(input.FromAddress)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(21000) //in units
	toAddress := common.HexToAddress(input.ToAddress)

	tx := types.NewTransaction(nonce, toAddress,
		input.Amount.Amount, //in wei
		gasLimit,
		input.Fee.Amount, //in wei
		nil)

	rlpTXBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("unsigned tx: %x\n", rlpTXBytes)
	//save result
	var result adaptor.CreateTransferTokenTxOutput
	result.Transaction = append(result.Transaction, rlpTXBytes...)

	return &result, nil
}

//func CreateContractMsg(input *adaptor.CreateTransferTokenTxInput, client *ethclient.Client,
//	fromAddress common.Address) (*adaptor.CreateTransferTokenTxOutput, error) {
//	gasLimit := uint64(21000) //in units
//	toAddress := common.HexToAddress(input.ToAddress)
//
//	var extra []byte
//	extra = append(extra, fromAddress.Bytes()...)
//	//extra = append(extra, input.Extra...)
//
//	tx := types.NewTransaction(0, toAddress,
//		&input.Amount.Amount, //in wei
//		gasLimit,
//		&input.Fee.Amount, //in wei
//		extra)
//
//	rlpTXBytes, err := rlp.EncodeToBytes(tx)
//	if err != nil {
//		return nil, err
//	}
//	//fmt.Printf("unsigned tx: %x\n", rlpTXBytes)
//	//save result
//	var result adaptor.CreateTransferTokenTxOutput
//	result.Transaction = append(result.Transaction, 'm')
//	result.Transaction = append(result.Transaction, rlpTXBytes...)
//
//	return &result, nil
//}

func CreateTx(input *adaptor.CreateTransferTokenTxInput) (*adaptor.CreateTransferTokenTxOutput, error) {
	if input.Amount == nil {
		return nil, errors.New("input's Amount is nil")
	}
	//if input.Fee == nil {
	//	return nil, errors.New("input's Fee is nil")
	//}

	var data []byte
	data = append(data, []byte("msg")...)
	//from
	fromAddress := common.HexToAddress(input.FromAddress)
	fromAddressPadded := common.LeftPadBytes(fromAddress.Bytes(), 32)
	data = append(data, fromAddressPadded...)
	//to
	toAddress := common.HexToAddress(input.ToAddress)
	toAddressPadded := common.LeftPadBytes(toAddress.Bytes(), 32)
	data = append(data, toAddressPadded...)
	//amount
	amountPadded := common.LeftPadBytes(input.Amount.Amount.Bytes(), 32)
	data = append(data, amountPadded...)
	//asset empty is eth, other is erc20
	if input.Amount.Asset != "" {
		token := common.HexToAddress(input.Amount.Asset)
		tokenPadded := common.LeftPadBytes(token.Bytes(), 32)
		data = append(data, tokenPadded...)
	}
	//extra, example: reqid
	extraPadded := common.LeftPadBytes(input.Extra, 32)
	data = append(data, extraPadded...)

	var result adaptor.CreateTransferTokenTxOutput
	result.Transaction = data

	return &result, nil

	////get rpc client
	//client, err := GetClient(rpcParams)
	//if err != nil {
	//	return nil, err
	//}
	//fromAddress := common.HexToAddress(input.FromAddress)
	//code, err := client.CodeAt(context.Background(), fromAddress, nil)
	//if len(code) > 0 {
	//	return CreateContractMsg(input, client, fromAddress)
	//} else {
	//	return CreateETHTx(input, client, fromAddress)
	//}
}

func SignMessage(input *adaptor.SignMessageInput) (*adaptor.SignMessageOutput, error) {
	priKey, err := crypto.ToECDSA(input.PrivateKey)
	if err != nil {
		return nil, err
	}

	hash := crypto.Keccak256Hash(input.Message)
	//fmt.Printf("%x\n", hash.Bytes())

	sig, err := crypto.Sign(hash.Bytes(), priKey)
	if err != nil {
		return nil, err
	}

	var result adaptor.SignMessageOutput
	result.Signature = sig

	return &result, nil
}

func VerifySignature(input *adaptor.VerifySignatureInput) (*adaptor.VerifySignatureOutput, error) {
	//
	hash := crypto.Keccak256Hash(input.Message)
	//fmt.Printf("%x\n", hash.Bytes())

	//
	pubkeyByte, err := crypto.Ecrecover(hash.Bytes(), input.Signature)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%x\n", pubkeyByte)

	var result adaptor.VerifySignatureOutput
	if len(pubkeyByte) == len(input.PublicKey) {
		result.Pass = bytes.Equal(pubkeyByte, input.PublicKey)
	} else {
		pubkey, err := crypto.UnmarshalPubkey(pubkeyByte)
		if err != nil {
			return nil, err
		}
		result.Pass = bytes.Equal(crypto.CompressPubkey(pubkey), input.PublicKey)
	}

	return &result, nil
}
