package vrfEs

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
)

type Ess struct {
}

func (e *Ess) VrfProve(priKey interface{}, msg []byte) (proof, selData []byte, err error) {
	proof, err = Evaluate(priKey.(*ecdsa.PrivateKey), msg)
	if err != nil {
		log.Error("VrfProve Evaluate fail")
	}
	return proof, proof, nil
}

func (e *Ess) VrfVerify(pk, msg, proof []byte) (bool, []byte, error) {
	if pk == nil || msg == nil || proof == nil {
		log.Error("VrfVerify param is nil")
		return false, nil, errors.New("VrfVerify fail, param is nil")
	}
	return VerifyWithPK(proof, msg, pk), proof, nil
}
