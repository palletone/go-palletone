package vrfEs

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
)

func VrfProve(pri *ecdsa.PrivateKey, msg []byte) (proof []byte, err error) {
	h := crypto.Keccak256Hash(util.RHashBytes(msg))
	if err != nil {
		return nil, err
	}
	proof, err = Evaluate(pri, h, msg)
	if err != nil {
		log.Error("VrfProve Evaluate fail")
	}
	return
}

func VrfVerify(pk []byte, msg, proof []byte) (bool, error) {
	return VerifyWithPK(proof, msg, pk), nil
}
