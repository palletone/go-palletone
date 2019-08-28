// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ptnapi

import (
	"encoding/hex"
	"sync"

	"github.com/palletone/go-palletone/common"
	//"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/ptnjson"
)

type AddrLocker struct {
	mu    sync.Mutex
	locks map[common.Address]*sync.Mutex
}

// lock returns the lock of the given address.
func (l *AddrLocker) lock(address common.Address) *sync.Mutex {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.locks == nil {
		l.locks = make(map[common.Address]*sync.Mutex)
	}
	if _, ok := l.locks[address]; !ok {
		l.locks[address] = new(sync.Mutex)
	}
	return l.locks[address]
}

// LockAddr locks an account's mutex. This is used to prevent another tx getting the
// same nonce until the lock is released. The mutex prevents the (an identical nonce) from
// being read again during the time that the first transaction is being signed.
func (l *AddrLocker) LockAddr(address common.Address) {
	l.lock(address).Lock()
}

// UnlockAddr unlocks the mutex of the given account.
func (l *AddrLocker) UnlockAddr(address common.Address) {
	l.lock(address).Unlock()
}

//func rpcDecodeHexError(gotHex string) *ptnjson.RPCError {
//	return ptnjson.NewRPCError(ptnjson.ErrRPCDecodeHexString,
//		fmt.Sprintf("Argument must be hexadecimal string (not %q)",
//			gotHex))
//}

/*func internalRPCError(errStr, context string) *ptnjson.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log.Warn(logStr)
	//rpcsLog.Error(logStr)
	return ptnjson.NewRPCError(ptnjson.ErrRPCInternal.Code, errStr)
}*/
func decodeHexStr(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, &ptnjson.RPCError{
			Code:    ptnjson.ErrRPCDecodeHexString,
			Message: "Hex string decode failed: " + err.Error(),
		}
	}
	return decoded, nil
}

//type response struct {
//	result []byte
	//err    error
//}
//type FutureGetTxOutResult chan *response
type SignatureError struct {
	InputIndex uint32
	Error      error
}

//var (
	// zeroHash is the zero value for a chainhash.Hash and is defined as
	// a package level variable to avoid the need to create a new instance
	// every time a check is needed.
	//zeroHash common.Hash

	// block91842Hash is one of the two nodes which violate the rules
	// set forth in BIP0030.  It is defined as a package level variable to
	// avoid the need to create a new instance every time a check is needed.
	//block91842Hash ,err = common.NewHashFromStr("00000000000a4d0a398161ffc163c503763b1f4360639393e0e4c8e300e0caec")

	// block91880Hash is one of the two nodes which violate the rules
	// set forth in BIP0030.  It is defined as a package level variable to
	// avoid the need to create a new instance every time a check is needed.
	//block91880Hash ,err = common.NewHashFromStr("00000000000743f190a18c5577a3c2d2a1f610ae9601ac046a38084ccb7cd721")
//)

type (
	// DeserializationError describes a failed deserializaion due to bad
	// user input.  It corresponds to btcjson.ErrRPCDeserialization.
	DeserializationError struct {
		error
	}
	// InvalidParameterError describes an invalid parameter passed by
	// the user.  It corresponds to btcjson.ErrRPCInvalidParameter.
	InvalidParameterError struct {
		error
	}
	// ParseError describes a failed parse due to bad user input.  It
	// corresponds to btcjson.ErrRPCParse.
	ParseError struct {
		error
	}
)

type SignTransactionParams struct {
	RawTx  string `json:"rawtx"`
	Inputs []struct {
		Txid         string `json:"txid"`
		Vout         uint32 `json:"vout"`
		MessageIndex uint32 `json:"messageindex"`
		ScriptPubKey string `json:"scriptPubKey"`
		RedeemScript string `json:"redeemScript"`
	} `json:"rawtxinput"`
	PrivKeys []string `json:"privkeys"`
	Flags    string   `jsonrpcdefault:"\"ALL\""`
}

// isNullOutpoint determines whether or not a previous transaction output point
// is set.
//func isNullOutpoint(outpoint *modules.OutPoint) bool {
//	if outpoint.OutIndex == math.MaxUint32 && outpoint.TxHash == zeroHash {
//		return true
//	}
//	return false
//}

//type SignTransactionResult struct {
//        TransactionHex string `json:"transactionhex"`
//        Complete       bool   `json:"complete"`
//}
/*func CheckTransactionSanity(tx *modules.Transaction) error {
	// A transaction must have at least one input.
	if len(tx.TxMessages) == 0 {
		return  &ptnjson.RPCError{
			Code:    ptnjson.ErrRPCRawTxString,
			Message: "transaction has no inputs",
		}
	}
	// A transaction must not exceed the maximum allowed block payload when
	// serialized.
	serializedTxSize := tx.SerializeSizeStripped()
	if serializedTxSize > ptnjson.MaxBlockBaseSize {
		str := fmt.Sprintf("serialized transaction is too big - got "+
			"%d, max %d", serializedTxSize, ptnjson.MaxBlockBaseSize)
		return  &ptnjson.RPCError{
			Code:    ptnjson.ErrRPCRawTxString,
			Message: str,
		}
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value known
	// as a satoshi.  One bitcoin is a quantity of satoshi as defined by the
	// SatoshiPerBitcoin constant.
	var totalSatoshi uint64
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}
		for _, txOut := range payload.Output {
			satoshi := txOut.Value
			if satoshi < 0 {
				str := fmt.Sprintf("transaction output has negative "+
					"value of %v", satoshi)
				return  &ptnjson.RPCError{
					Code:    ptnjson.ErrBadTxOutValue,
					Message: str,
				}
			}
			if satoshi > ptnjson.MaxDao {
				str := fmt.Sprintf("transaction output value of %v is "+
					"higher than max allowed value of %v", satoshi,
					ptnjson.MaxDao)
				return  &ptnjson.RPCError{
					Code:    ptnjson.ErrBadTxOutValue,
					Message: str,
				}
			}

			// Two's complement int64 overflow guarantees that any overflow
			// is detected and reported.  This is impossible for Bitcoin, but
			// perhaps possible if an alt increases the total money supply.
			totalSatoshi += satoshi
			if totalSatoshi < 0 {
				str := fmt.Sprintf("total value of all transaction "+
					"outputs exceeds max allowed value of %v",
					ptnjson.MaxDao)
				return  &ptnjson.RPCError{
					Code:    ptnjson.ErrBadTxOutValue,
					Message: str,
				}
			}
			if totalSatoshi > ptnjson.MaxDao {
				str := fmt.Sprintf("total value of all transaction "+
					"outputs is %v which is higher than max "+
					"allowed value of %v", totalSatoshi,
					ptnjson.MaxDao)
				return  &ptnjson.RPCError{
					Code:    ptnjson.ErrBadTxOutValue,
					Message: str,
				}
			}
		}


	// Check for duplicate transaction inputs.
	existingTxOut := make(map[modules.OutPoint]struct{})
	for _, txIn := range payload.Input {
		if _, exists := existingTxOut[*txIn.PreviousOutPoint]; exists {
			return  &ptnjson.RPCError{
					Code:    ptnjson.ErrDuplicateTxInputs,
					Message:  "transaction "+"contains duplicate inputs",
				}
		}
		existingTxOut[*txIn.PreviousOutPoint] = struct{}{}
	}

	// Coinbase script length must be between min and max length.
	if dagcommon.IsCoinBase(tx) {
		slen := len(payload.Input[0].SignatureScript)
		if slen < ptnjson.MinCoinbaseScriptLen || slen > ptnjson.MaxCoinbaseScriptLen {
			str := fmt.Sprintf("coinbase transaction script length "+
				"of %d is out of range (min: %d, max: %d)",
				slen, ptnjson.MinCoinbaseScriptLen, ptnjson.MaxCoinbaseScriptLen)
			return  &ptnjson.RPCError{
					Code:    ptnjson.ErrBadCoinbaseScriptLen,
					Message:  str,
				}
		}
	} else {
		    // Previous transaction outputs referenced by the inputs to this
		    // transaction must not be null.
			for _, txIn := range payload.Input {
				if isNullOutpoint(txIn.PreviousOutPoint) {
					return  &ptnjson.RPCError{
					    Code:    ptnjson.ErrBadTxInput,
					    Message:  "transaction "+
						"input refers to previous output that "+
						"is null",
				    }
				}
			}
	    }
    }

	return nil
}*/
