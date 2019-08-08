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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package accesscontrol

import (
	"encoding/pem"
	"sync"
	"time"

	"github.com/palletone/go-palletone/core/vmContractPub/util"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

var ttl = time.Minute * 10

type certHash string

type certMapper struct {
	keyGen KeyGenFunc
	sync.RWMutex
	m   map[certHash]string
	tls bool
}

func newCertMapper(keyGen KeyGenFunc) *certMapper {
	return &certMapper{
		keyGen: keyGen,
		tls:    viper.GetBool("peer.tls.enabled"),
		m:      make(map[certHash]string),
	}
}

func (r *certMapper) lookup(h certHash) string {
	r.RLock()
	defer r.RUnlock()
	return r.m[h]
}

func (r *certMapper) register(hash certHash, name string) {
	r.Lock()
	defer r.Unlock()
	r.m[hash] = name
	time.AfterFunc(ttl, func() {
		r.purge(hash)
	})
}

func (r *certMapper) purge(hash certHash) {
	r.Lock()
	defer r.Unlock()
	delete(r.m, hash)
}

//func certKeyPairFromString(privKey string, pubKey string) (*certKeyPair, error) {
//	priv, err := base64.StdEncoding.DecodeString(privKey)
//	if err != nil {
//		return nil, err
//	}
//	pub, err := base64.StdEncoding.DecodeString(pubKey)
//	if err != nil {
//		return nil, err
//	}
//	return &certKeyPair{
//		CertKeyPair: &CertKeyPair{
//			Key:  priv,
//			Cert: pub,
//		},
//	}, nil
//}

func (r *certMapper) genCert(name string) (*certKeyPair, error) {
	keyPair, err := r.keyGen()
	if err != nil {
		return nil, err
	}
	hash := util.ComputeSHA256(keyPair.cert.Raw)
	r.register(certHash(hash), name)
	return keyPair, nil
}

func encodePEM(keyType string, data []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: keyType, Bytes: data})
}

// ExtractCertificateHash extracts the hash of the certificate from the stream
func extractCertificateHashFromContext(ctx context.Context) []byte {
	pr, extracted := peer.FromContext(ctx)
	if !extracted {
		return nil
	}

	authInfo := pr.AuthInfo
	if authInfo == nil {
		return nil
	}

	tlsInfo, isTLSConn := authInfo.(credentials.TLSInfo)
	if !isTLSConn {
		return nil
	}
	certs := tlsInfo.State.PeerCertificates
	if len(certs) == 0 {
		return nil
	}
	raw := certs[0].Raw
	if len(raw) == 0 {
		return nil
	}
	return util.ComputeSHA256(raw)
}
