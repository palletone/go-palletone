/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package crypto

import "hash"

type ICrypto interface {

	// KeyGen generates a key using opts.
	KeyGen() (privKey []byte, err error)
	PrivateKeyToPubKey(privKey []byte) (pubKey []byte, err error)
	PrivateKeyToInstance(privKey []byte) (interface{}, error)
	// KeyDeriv derives a key from k using opts.
	// The opts argument should be appropriate for the primitive used.
	//KeyDeriv(k Key, opts KeyDerivOpts) (dk Key, err error)

	// KeyImport imports a key from its raw representation using opts.
	// The opts argument should be appropriate for the primitive used.
	//KeyImport(raw interface{}, opts KeyImportOpts) (k Key, err error)

	// GetKey returns the key this CSP associates to
	// the Subject Key Identifier ski.
	//GetKey(ski []byte) (k Key, err error)

	// Hash hashes messages msg using options opts.
	// If opts is nil, the default hash function will be used.
	Hash(msg []byte) (hash []byte, err error)

	// GetHash returns and instance of hash.Hash using options opts.
	// If opts is nil, the default hash function will be returned.
	GetHash() (h hash.Hash, err error)

	// Sign signs digest using key k.
	// The opts argument should be appropriate for the algorithm used.
	//
	// Note that when a signature of a hash of a larger message is needed,
	// the caller is responsible for hashing the larger message and passing
	// the hash (as digest).
	Sign(privKey, message []byte) (signature []byte, err error)

	// Verify verifies signature against key k and digest
	// The opts argument should be appropriate for the algorithm used.
	Verify(pubKey, signature, message []byte) (valid bool, err error)

	// Encrypt encrypts plaintext using key k.
	// The opts argument should be appropriate for the algorithm used.
	Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error)

	// Decrypt decrypts ciphertext using key k.
	// The opts argument should be appropriate for the algorithm used.
	Decrypt(key, ciphertext []byte) (plaintext []byte, err error)
}
