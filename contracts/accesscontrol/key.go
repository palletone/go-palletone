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
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

type KeyGenFunc func() (*certKeyPair, error)

type certKeyPair struct {
	*CertKeyPair
	crypto.Signer
	cert *x509.Certificate
}

func (p *certKeyPair) privKeyString() string {
	return base64.StdEncoding.EncodeToString(p.Key)
}

func (p *certKeyPair) pubKeyString() string {
	return base64.StdEncoding.EncodeToString(p.Cert)
}

func newPrivKey() (*ecdsa.PrivateKey, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, privBytes, nil
}

func newCertTemplate() (x509.Certificate, error) {
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return x509.Certificate{}, err
	}
	return x509.Certificate{
		Subject:      pkix.Name{SerialNumber: sn.String()},
		NotBefore:    time.Now().Add(time.Hour * (-24)),
		NotAfter:     time.Now().Add(time.Hour * 24),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		SerialNumber: sn,
	}, nil
}

func newCertKeyPair(isCA bool, isServer bool, host string, certSigner crypto.Signer, parent *x509.Certificate) (*certKeyPair, error) {
	privateKey, privBytes, err := newPrivKey()
	if err != nil {
		return nil, err
	}

	template, err := newCertTemplate()
	if err != nil {
		return nil, err
	}

	tenYearsFromNow := time.Now().Add(time.Hour * 24 * 365 * 10)
	if isCA {
		template.NotAfter = tenYearsFromNow
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
		template.BasicConstraintsValid = true
	} else {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	if isServer {
		template.NotAfter = tenYearsFromNow
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		if ip := net.ParseIP(host); ip != nil {
			logger.Debug("Classified", host, "as an IP address, adding it as an IP SAN")
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			logger.Debug("Classified", host, "as a hostname, adding it as a DNS SAN")
			template.DNSNames = append(template.DNSNames, host)
		}
	}
	// If no parent cert, it's a self signed cert
	if parent == nil || certSigner == nil {
		parent = &template
		certSigner = privateKey
	}
	rawBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, &privateKey.PublicKey, certSigner)
	if err != nil {
		return nil, err
	}
	pubKey := encodePEM("CERTIFICATE", rawBytes)

	block, _ := pem.Decode(pubKey)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	privKey := encodePEM("EC PRIVATE KEY", privBytes)
	return &certKeyPair{
		CertKeyPair: &CertKeyPair{
			Key:  privKey,
			Cert: pubKey,
		},
		Signer: privateKey,
		cert:   cert,
	}, nil
}
