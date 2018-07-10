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

package modules

// import (
// 	"crypto/ecdsa"
// 	"errors"
// 	"fmt"
// 	"math/big"
//
// 	"github.com/palletone/go-palletone/common"
// 	"github.com/palletone/go-palletone/common/crypto"
// )
//
// var (
// 	ErrInvalidChainId = errors.New("invalid chain id for signer")
// )
//
// // sigCache is used to cache the derived sender and contains
// // the signer used to derive it.
// type sigCache struct {
//	signer Signer
// 	from common.Address
// }
//
// // SignTx signs the transaction using the given signer and private key
// func SignTx(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
// 	h := s.Hash(tx)
// 	sig, err := crypto.Sign(h[:], prv)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tx.WithSignature(s, sig)
// }
//
// // Sender returns the address derived from the signature (V, R, S) using secp256k1
// // elliptic curve and an error if it failed deriving or upon an incorrect
// // signature.
////
// // Sender may cache the address, allowing it to be used regardless of
// // signing method. The cache is invalidated if the cached signer does
// // not match the signer used in the current call.
// func Sender(signer Signer, tx *Transaction) (common.Address, error) {
// 	if sc := tx.from.Load(); sc != nil {
// 		sigCache := sc.(sigCache)
// 		// If the signer used to derive from in a previous
// 		// call is not the same as used current, invalidate
// 		// the cache.
// 		if sigCache.signer.Equal(signer) {
// 			return sigCache.from, nil
// 		}
// 	}
//
// 	addr, err := signer.Sender(tx)
// 	if err != nil {
// 		return common.Address{}, err
// 	}
// 	tx.from.Store(sigCache{signer: signer, from: addr})
// 	return addr, nil
// }
//
// // Signer encapsulates transaction signature handling. Note that this interface is not a
// // stable API and may change at any time to accommodate new protocol rules.
// type Signer interface {
// 	// Sender returns the sender address of the transaction.
// 	Sender(tx *Transaction) (common.Address, error)
// 	// SignatureValues returns the raw R, S, V values corresponding to the
// 	// given signature.
// 	SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error)
// 	// Hash returns the UnitHash to be signed.
// 	Hash(tx *Transaction) common.Hash
// 	// Equal returns true if the given signer is the same as the receiver.
// 	Equal(Signer) bool
// }
//
// // EIP155Transaction implements Signer using the EIP155 rules.
// type EIP155Signer struct {
// 	chainId, chainIdMul *big.Int
// }
//
// func NewEIP155Signer(chainId *big.Int) EIP155Signer {
// 	if chainId == nil {
// 		chainId = new(big.Int)
// 	}
// 	return EIP155Signer{
// 		chainId:    chainId,
// 		chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2)),
// 	}
// }
//
// func (s EIP155Signer) Equal(s2 Signer) bool {
// 	eip155, ok := s2.(EIP155Signer)
// 	return ok && eip155.chainId.Cmp(s.chainId) == 0
// }
//
// var big8 = big.NewInt(8)
//
// func (s EIP155Signer) Sender(tx *Transaction) (common.Address, error) {
// 	if !tx.Protected() {
// 		return HomesteadSigner{}.Sender(tx)
// 	}
// 	if tx.ChainId().Cmp(s.chainId) != 0 {
// 		return common.Address{}, ErrInvalidChainId
// 	}
// 	V := new(big.Int).Sub(tx.data.V, s.chainIdMul)
// 	V.Sub(V, big8)
// 	return recoverPlain(s.Hash(tx), tx.data.R, tx.data.S, V, true)
// }
//
// // WithSignature returns a new transaction with the given signature. This signature
// // needs to be in the [R || S || V] format where V is 0 or 1.
// func (s EIP155Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
// 	R, S, V, err = HomesteadSigner{}.SignatureValues(tx, sig)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	if s.chainId.Sign() != 0 {
// 		V = big.NewInt(int64(sig[64] + 35))
// 		V.Add(V, s.chainIdMul)
// 	}
// 	return R, S, V, nil
// }
//
// // Hash returns the UnitHash to be signed by the sender.
// // It does not uniquely identify the transaction.
// func (s EIP155Signer) Hash(tx *Transaction) common.Hash {
// 	return rlpHash([]interface{}{
// 		tx.data.From,
// 		tx.data.Price,
// 		tx.data.Recipient,
// 		tx.data.Amount,
// 		tx.data.Payload,
// 		s.chainId, uint(0), uint(0),
// 	})
// }
//
// // HomesteadTransaction implements TransactionInterface using the
// // homestead rules.
// type HomesteadSigner struct{ FrontierSigner }
//
// func (s HomesteadSigner) Equal(s2 Signer) bool {
// 	_, ok := s2.(HomesteadSigner)
// 	return ok
// }
//
// // SignatureValues returns signature values. This signature
// // needs to be in the [R || S || V] format where V is 0 or 1.
// func (hs HomesteadSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
// 	return hs.FrontierSigner.SignatureValues(tx, sig)
// }
//
// func (hs HomesteadSigner) Sender(tx *Transaction) (common.Address, error) {
// 	return recoverPlain(hs.Hash(tx), tx.data.R, tx.data.S, tx.data.V, true)
// }
//
// type FrontierSigner struct{}
//
// func (s FrontierSigner) Equal(s2 Signer) bool {
// 	_, ok := s2.(FrontierSigner)
// 	return ok
// }
//
// // SignatureValues returns signature values. This signature
// // needs to be in the [R || S || V] format where V is 0 or 1.
// func (fs FrontierSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
// 	if len(sig) != 65 {
// 		panic(fmt.Sprintf("wrong UnitSize for signature: got %d, want 65", len(sig)))
// 	}
// 	r = new(big.Int).SetBytes(sig[:32])
// 	s = new(big.Int).SetBytes(sig[32:64])
// 	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
// 	return r, s, v, nil
// }
//
// // Hash returns the UnitHash to be signed by the sender.
// // It does not uniquely identify the transaction.
// func (fs FrontierSigner) Hash(tx *Transaction) common.Hash {
// 	return rlpHash([]interface{}{
// 		tx.data.From,
// 		tx.data.Price,
// 		tx.data.Recipient,
// 		tx.data.Amount,
// 		tx.data.Payload,
// 	})
// }
//
// func (fs FrontierSigner) Sender(tx *Transaction) (common.Address, error) {
// 	return recoverPlain(fs.Hash(tx), tx.data.R, tx.data.S, tx.data.V, false)
// }
//
// func recoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (common.Address, error) {
// 	if Vb.BitLen() > 8 {
// 		return common.Address{}, errors.New("invalid transaction v, r, s values")
// 	}
// 	V := byte(Vb.Uint64() - 27)
// 	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
// 		return common.Address{}, errors.New("invalid transaction v, r, s values")
// 	}
// 	// encode the snature in uncompressed format
// 	r, s := R.Bytes(), S.Bytes()
// 	sig := make([]byte, 65)
// 	copy(sig[32-len(r):32], r)
// 	copy(sig[64-len(s):64], s)
// 	sig[64] = V
// 	// recover the public key from the snature
// 	pub, err := crypto.Ecrecover(sighash[:], sig)
// 	if err != nil {
// 		return common.Address{}, err
// 	}
// 	if len(pub) == 0 || pub[0] != 4 {
// 		return common.Address{}, errors.New("invalid public key")
// 	}
// 	var addr common.Address
// 	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
// 	return addr, nil
// }
