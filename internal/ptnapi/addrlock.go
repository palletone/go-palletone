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
	"sync"
	"fmt"
        "encoding/hex"
	"github.com/palletone/go-palletone/common"
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
func rpcDecodeHexError(gotHex string) *ptnjson.RPCError {
	return ptnjson.NewRPCError(ptnjson.ErrRPCDecodeHexString,
		fmt.Sprintf("Argument must be hexadecimal string (not %q)",
			gotHex))
}
func internalRPCError(errStr, context string) *ptnjson.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	fmt.Println(logStr)
	//rpcsLog.Error(logStr)
	return ptnjson.NewRPCError(ptnjson.ErrRPCInternal.Code, errStr)
}
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
type response struct {
        result []byte
        err    error
}
type FutureGetTxOutResult chan *response
type SignatureError struct {
        InputIndex uint32
        Error      error
}
type SigHashType uint32
const (
        SigHashOld          SigHashType = 0x0
        SigHashAll          SigHashType = 0x1
        SigHashNone         SigHashType = 0x2
        SigHashSingle       SigHashType = 0x3
        SigHashAnyOneCanPay SigHashType = 0x80
        // sigHashMask defines the number of bits of the hash type which is used
        // to identify which outputs are signed.
        sigHashMask = 0x1f
)
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
