package vrfEs

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/util"
)

func Evaluate(pri *ecdsa.PrivateKey,  msg []byte) (proof []byte, err error) {
	h := crypto.Keccak256Hash(util.RHashBytes(msg))
	sign, err := crypto.Sign(h.Bytes(), pri)
	if err != nil {
		return nil, err
	}
	return sign, nil
}

func VerifyWithPK(sign []byte, msg interface{}, publicKey []byte) bool {
	hash := crypto.Keccak256Hash(util.RHashBytes(msg))
	// sig := sign[:len(sign)-1] // remove recovery id
	return crypto.VerifySignature(publicKey, hash.Bytes(), sign)
}
