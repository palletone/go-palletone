package vrfEs

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
)

func VrfProve(pri *ecdsa.PrivateKey, msg []byte) (proof []byte, err error) {
	h := crypto.Keccak256Hash(util.RHashBytes(msg))
	proof, err = Evaluate(pri, h, msg)
	if err != nil {
		log.Error("VrfProve Evaluate fail")
	}
	return
}

func VrfVerify(pk , msg, proof []byte) (bool, error) {
	if pk == nil || msg == nil || proof == nil{
		log.Error("VrfVerify param is nil")
		return false, errors.New("VrfVerify fail, param is nil")
	}
	return VerifyWithPK(proof, msg, pk), nil
}
