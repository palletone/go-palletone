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
 *  * @date 2018
 *
 */
package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"github.com/palletone/digital-identity/config"
	"golang.org/x/crypto/sha3"
	"hash"
	"math/big"
	"net"
	"net/mail"
	"fmt"
)

// CryptSuite defines common interface for different crypto implementations.
type CryptoSuite interface {
	// GenerateKey returns PrivateKey.
	GenerateKey() (interface{}, error)
	// CreateCertificateRequest will create CSR request. It takes enrolmentId and Private key
	CreateCertificateRequest(enrollmentID string, key interface{}, hosts []string) ([]byte, error)
	// Sign signs message. It takes message to sign and Private key
	Sign(msg []byte, k interface{}) ([]byte, error)
	// Hash computes Hash value of provided data. Hash function will be different in different crypto implementations.
	Hash(data []byte) []byte
}

var (
	// precomputed curves half order values for efficiency
	ecCurveHalfOrders = map[elliptic.Curve]*big.Int{
		elliptic.P224(): new(big.Int).Rsh(elliptic.P224().Params().N, 1),
		elliptic.P256(): new(big.Int).Rsh(elliptic.P256().Params().N, 1),
		elliptic.P384(): new(big.Int).Rsh(elliptic.P384().Params().N, 1),
		elliptic.P521(): new(big.Int).Rsh(elliptic.P521().Params().N, 1),
	}
)

type ECCryptSuite struct {
	curve        elliptic.Curve
	sigAlgorithm x509.SignatureAlgorithm
	//key          *ecdsa.PrivateKey
	hashFunction func() hash.Hash
}

type eCDSASignature struct {
	R, S *big.Int
}

func (c *ECCryptSuite) GenerateKey() (interface{}, error) {
	key, err := ecdsa.GenerateKey(c.curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (c *ECCryptSuite) CreateCertificateRequest(enrollmentID string, key interface{}, hosts []string) ([]byte, error) {
	if enrollmentID == "" {
		return nil, config.ErrEnrollmentIDMissing
	}
	subj := pkix.Name{
		CommonName: enrollmentID,
	}
	rawSubj := subj.ToRDNSequence()

	asn1Subj, err := asn1.Marshal(rawSubj)
	if err != nil {
		return nil, err
	}

	ipAddr := make([]net.IP, 0)
	emailAddr := make([]string, 0)
	dnsAddr := make([]string, 0)

	for i := range hosts {
		if ip := net.ParseIP(hosts[i]); ip != nil {
			ipAddr = append(ipAddr, ip)
		} else if email, _ := mail.ParseAddress(hosts[i]); email != nil {
			emailAddr = append(emailAddr, email.Address)
		} else {
			dnsAddr = append(dnsAddr, hosts[i])
		}
	}

	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: c.sigAlgorithm,
		IPAddresses:        ipAddr,
		EmailAddresses:     emailAddr,
		DNSNames:           dnsAddr,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		return nil, err
	}
	csr := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csr, nil
}

func (c *ECCryptSuite) Sign(msg []byte, k interface{}) ([]byte, error) {
	key, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, config.ErrInvalidKeyType
	}

	h := c.Hash(msg)
	R, S, err := ecdsa.Sign(rand.Reader, key, h)
	if err != nil {
		return nil, err
	}
	c.preventMalleability(key, S)
	sig, err := asn1.Marshal(eCDSASignature{R: R, S: S})
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (c *ECCryptSuite) preventMalleability(k *ecdsa.PrivateKey, s *big.Int) {
	halfOrder := ecCurveHalfOrders[k.Curve]
	if s.Cmp(halfOrder) == 1 {
		s.Sub(k.Params().N, s)
	}
}
func (c *ECCryptSuite) Hash(data []byte) []byte {
	h := c.hashFunction()
	_, err := h.Write(data)
	if err != nil {
		fmt.Println(err)
	}
	return h.Sum(nil)
}

func NewECCryptoSuiteFromConfig(conf config.CryptoConfig) (CryptoSuite, error) {
	var suite *ECCryptSuite

	switch conf.Algorithm {
	case "P256-SHA256":
		suite = &ECCryptSuite{curve: elliptic.P256(), sigAlgorithm: x509.ECDSAWithSHA256}
	case "P384-SHA384":
		suite = &ECCryptSuite{curve: elliptic.P384(), sigAlgorithm: x509.ECDSAWithSHA384}
	case "P521-SHA512":
		suite = &ECCryptSuite{curve: elliptic.P521(), sigAlgorithm: x509.ECDSAWithSHA512}
	default:
		return nil, config.ErrInvalidAlgorithm
	}

	switch conf.Hash {
	case "SHA2-256":
		suite.hashFunction = sha256.New
	case "SHA2-384":
		suite.hashFunction = sha512.New384
	case "SHA3-256":
		suite.hashFunction = sha3.New256
	case "SHA3-384":
		suite.hashFunction = sha3.New384
	default:
		return nil, config.ErrInvalidHash
	}
	return suite, nil
}
