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
