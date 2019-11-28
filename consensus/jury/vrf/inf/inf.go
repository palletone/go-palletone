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
package vrf

import "crypto"

// PrivateKey supports evaluating the VRF function.
type PrivateKey interface {
	// Evaluate returns the output of H(f_k(m)) and its proof.
	Evaluate(m []byte) (index [32]byte, proof []byte)
	// Public returns the corresponding public key.
	Public() crypto.PublicKey
}

// PublicKey supports verifying output from the VRF function.
type PublicKey interface {
	// ProofToHash verifies the NP-proof supplied by Proof and outputs Index.
	ProofToHash(m, proof []byte) (index [32]byte, err error)
}
