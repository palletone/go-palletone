package vrf

import (
	"crypto/ecdsa"
	"github.com/palletone/go-palletone/common/log"
	"github.com/btcsuite/btcd/btcec"
)

type Es struct {
}

func (e *Es) VrfProve(priKey interface{}, msg []byte) (proof ,selData []byte, err error) {
	siger, err := NewVRFSigner(priKey.(*ecdsa.PrivateKey))
	if err != nil {
		log.Errorf("VrfProve, NewVRFSigner err:%s", err.Error())
		return nil, nil,err
	}
	idx, proof := siger.Evaluate(msg)
	log.Debugf("VrfProve, msg[%v], idx[%v], proof[%v]", msg, idx, proof)

	return proof, idx[:],nil
}

func (e *Es) VrfVerify(pubKey, msg, proof []byte) (bool, []byte, error) {
	key, err := btcec.ParsePubKey(pubKey, btcec.S256())
	if err != nil {
		log.Errorf("VrfVerify, parsePubKey error:%s", err.Error())
		return false, nil, err
	}
	pk, err := NewVRFVerifier(key.ToECDSA())
	if err != nil {
		log.Errorf("VrfVerify, NewVRFVerifier error:%s", err.Error())
		return false, nil, err
	}
	idx, err := pk.ProofToHash(msg, proof)
	if err != nil {
		log.Errorf("VrfVerify, ProofToHash error:%s", err.Error())
		return false, nil, err
	}
	return true, idx[:], nil
}
