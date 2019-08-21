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
package certficate

import (
	"crypto"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path"

	"github.com/palletone/go-palletone/common/crypto/gmsm/sm2"
	"gopkg.in/errgo.v1"
)

type GmCertInfo struct {
	Address         string
	Country         []string
	Locality        []string
	Organization    []string
	EmailAddresses  namesVar
	IPAddresses     ipsVar
	// subjectAltEmail namesVar
}

func (c *GmCertInfo) GenSMCert() ([]byte, error) {
	key, err := sm2.GenerateKey()
	if err != nil {
		return nil, err
	}

	err = SaveSM2Key(key)
	if err != nil {
		return nil, err
	}
	subj := pkix.Name{
		Country:      c.Country,
		Organization: c.Organization,
		Locality:     c.Locality,
	}

	template := &sm2.Certificate{
		Subject:        subj,
		DNSNames:       DNSNames(),
		EmailAddresses: c.EmailAddresses,
		IPAddresses:    c.IPAddresses,
		IsCA:           false,
	}

	crt, err := SelfSignCertificate(template, key, c.Address)
	if err != nil {
		return nil, err
	}
	certbyte, err := WriteCertificate(os.Stdout, crt)
	if err != nil {
		return nil, err
	}

	return certbyte, nil
}

func SelfSignCertificate(params *sm2.Certificate, key crypto.Signer, address string) (*sm2.Certificate, error) {
	template := *params
	if err := generateCertificateValues(&template, key.Public(), address); err != nil {
		return nil, errgo.Mask(err)
	}

	data, err := sm2.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
	if err != nil {
		return nil, errgo.Mask(err)
	}
	crt, err := sm2.ParseCertificate(data)
	if err != nil {
		// If we can't parse a certificate that we've just created something is very wrong.
		panic(err)
	}
	return crt, nil
}

func generateCertificateValues(template *sm2.Certificate, publicKey interface{}, address string) error {
	if template.SerialNumber == nil {
		max := big.NewInt(1)
		max.Lsh(max, 20*8)
		var err error
		template.SerialNumber, err = rand.Int(rand.Reader, max)
		if err != nil {
			return errgo.Notef(err, "cannot generate serial number")
		}
	}
	if template.SubjectKeyId == nil {
		data, err := sm2.MarshalPKIXPublicKey(publicKey)
		if err == nil {
			sum := sha1.Sum(data)
			template.SubjectKeyId = sum[:]
		}
	}
	template.Subject.CommonName = address
	return nil
}

func SaveSM2Key(key *sm2.PrivateKey) error {
	keyPath := path.Join("./", "/keystore")
	err := os.MkdirAll(keyPath, os.ModePerm)
	if err != nil {
		return err
	}
	keyFile := path.Join(keyPath, "key.pem")
	keyByte, err := sm2.MarshalSm2EcryptedPrivateKey(key, nil)
	if err != nil {
		return err
	}
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "SM PRIVATE KEY",
			Bytes: keyByte,
		},
	)
	err = ioutil.WriteFile(keyFile, keyPem, 0644)
	if err != nil {
		return nil
	}
	return nil
}

func WriteCertificate(w io.Writer, crt *sm2.Certificate) ([]byte, error) {
	caPath := path.Join("./", "/certs")
	err := os.MkdirAll(caPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	caFile := path.Join(caPath, "cert.pem")
	if crt.Raw == nil {
		return nil, errgo.New("invalid certificate")
	}
	caPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: crt.Raw,
		},
	)
	err = ioutil.WriteFile(caFile, caPem, 0644)
	if err != nil {
		return nil, err
	}
	return caPem, nil
}
