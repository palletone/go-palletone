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

// CertKeyPair denotes a TLS certificate and corresponding key,
// both PEM encoded
type CertKeyPair struct {
	// Cert is the certificate, PEM encoded
	Cert []byte
	// Key is the key corresponding to the certificate, PEM encoded
	Key []byte
}

// CA defines a certificate authority that can generate
// certificates signed by it
type CA interface {
	// CertBytes returns the certificate of the CA in PEM encoding
	CertBytes() []byte

	// newCertKeyPair returns a certificate and private key pair and nil,
	// or nil, error in case of failure
	// The certificate is signed by the CA and is used for TLS client authentication
	newClientCertKeyPair() (*certKeyPair, error)

	// NewServerCertKeyPair returns a CertKeyPair and nil,
	// with a given custom SAN.
	// The certificate is signed by the CA.
	// Returns nil, error in case of failure
	NewServerCertKeyPair(host string) (*CertKeyPair, error)
}

type ca struct {
	caCert *certKeyPair
}

func NewCA() (CA, error) {
	c := &ca{}
	var err error
	c.caCert, err = newCertKeyPair(true, false, "", nil, nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CertBytes returns the certificate of the CA in PEM encoding
func (c *ca) CertBytes() []byte {
	return c.caCert.Cert
}

// newClientCertKeyPair returns a certificate and private key pair and nil,
// or nil, error in case of failure
// The certificate is signed by the CA and is used as a client TLS certificate
func (c *ca) newClientCertKeyPair() (*certKeyPair, error) {
	return newCertKeyPair(false, false, "", c.caCert.Signer, c.caCert.cert)
}

// newServerCertKeyPair returns a certificate and private key pair and nil,
// or nil, error in case of failure
// The certificate is signed by the CA and is used as a server TLS certificate
func (c *ca) NewServerCertKeyPair(host string) (*CertKeyPair, error) {
	keypair, err := newCertKeyPair(false, true, host, c.caCert.Signer, c.caCert.cert)
	if err != nil {
		return nil, err
	}
	return keypair.CertKeyPair, nil
}
