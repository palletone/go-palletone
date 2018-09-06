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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func createTLSService(t *testing.T, ca CA, host string) *grpc.Server {
	keyPair, err := ca.NewServerCertKeyPair(host)
	cert, err := tls.X509KeyPair(keyPair.Cert, keyPair.Key)
	assert.NoError(t, err)
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    x509.NewCertPool(),
	}
	tlsConf.ClientCAs.AppendCertsFromPEM(ca.CertBytes())
	return grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConf)))
}

func TestTLSCA(t *testing.T) {
	// This test checks that the CA can create certificates
	// and corresponding keys that are signed by itself

	rand.Seed(time.Now().UnixNano())
	randomPort := 1234 + rand.Intn(1234) // some random port

	ca, err := NewCA()
	assert.NoError(t, err)
	assert.NotNil(t, ca)

	endpoint := fmt.Sprintf("127.0.0.1:%d", randomPort)
	srv := createTLSService(t, ca, "127.0.0.1")
	l, err := net.Listen("tcp", endpoint)
	assert.NoError(t, err)
	go srv.Serve(l)
	defer srv.Stop()
	defer l.Close()

	probeTLS := func(kp *certKeyPair) error {
		keyBytes, err := base64.StdEncoding.DecodeString(kp.privKeyString())
		assert.NoError(t, err)
		certBytes, err := base64.StdEncoding.DecodeString(kp.pubKeyString())
		assert.NoError(t, err)
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		tlsCfg := &tls.Config{
			RootCAs:      x509.NewCertPool(),
			Certificates: []tls.Certificate{cert},
		}
		tlsCfg.RootCAs.AppendCertsFromPEM(ca.CertBytes())
		tlsOpts := grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, time.Second)
		conn, err := grpc.DialContext(ctx, fmt.Sprintf("127.0.0.1:%d", randomPort), tlsOpts, grpc.WithBlock())
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}

	// Good path - use a cert key pair generated from the CA
	// that the TLS server started with
	kp, err := ca.newClientCertKeyPair()
	assert.NoError(t, err)
	err = probeTLS(kp)
	assert.NoError(t, err)

	// Bad path - use a cert key pair generated from a foreign CA
	foreignCA, _ := NewCA()
	kp, err = foreignCA.newClientCertKeyPair()
	assert.NoError(t, err)
	err = probeTLS(kp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tls: bad certificate")
}
