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

package comm

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const defaultTimeout = time.Second * 3

//var log = flogging.MustGetLogger("comm")
var credSupport *CredentialSupport
var once sync.Once

// CASupport type manages certificate authorities scoped by channel
type CASupport struct {
	sync.RWMutex
	AppRootCAsByChain     map[string][][]byte
	OrdererRootCAsByChain map[string][][]byte
	ClientRootCAs         [][]byte
	ServerRootCAs         [][]byte
}

// CredentialSupport type manages credentials used for gRPC client connections
type CredentialSupport struct {
	*CASupport
	clientCert tls.Certificate
}

// GetCredentialSupport returns the singleton CredentialSupport instance
func GetCredentialSupport() *CredentialSupport {

	once.Do(func() {
		credSupport = &CredentialSupport{
			CASupport: &CASupport{
				AppRootCAsByChain:     make(map[string][][]byte),
				OrdererRootCAsByChain: make(map[string][][]byte),
			},
		}
	})
	return credSupport
}

// GetServerRootCAs returns the PEM-encoded root certificates for all of the
// application and orderer organizations defined for all chains.  The root
// certificates returned should be used to set the trusted server roots for
// TLS clients.
func (cas *CASupport) GetServerRootCAs() (appRootCAs, ordererRootCAs [][]byte) {
	cas.RLock()
	defer cas.RUnlock()

	appRootCAs = [][]byte{}
	ordererRootCAs = [][]byte{}

	for _, appRootCA := range cas.AppRootCAsByChain {
		appRootCAs = append(appRootCAs, appRootCA...)
	}

	for _, ordererRootCA := range cas.OrdererRootCAsByChain {
		ordererRootCAs = append(ordererRootCAs, ordererRootCA...)
	}

	// also need to append statically configured root certs
	appRootCAs = append(appRootCAs, cas.ServerRootCAs...)
	return appRootCAs, ordererRootCAs
}

// GetClientRootCAs returns the PEM-encoded root certificates for all of the
// application and orderer organizations defined for all chains.  The root
// certificates returned should be used to set the trusted client roots for
// TLS servers.
func (cas *CASupport) GetClientRootCAs() (appRootCAs, ordererRootCAs [][]byte) {
	cas.RLock()
	defer cas.RUnlock()

	appRootCAs = [][]byte{}
	ordererRootCAs = [][]byte{}

	for _, appRootCA := range cas.AppRootCAsByChain {
		appRootCAs = append(appRootCAs, appRootCA...)
	}

	for _, ordererRootCA := range cas.OrdererRootCAsByChain {
		ordererRootCAs = append(ordererRootCAs, ordererRootCA...)
	}

	// also need to append statically configured root certs
	appRootCAs = append(appRootCAs, cas.ClientRootCAs...)
	return appRootCAs, ordererRootCAs
}

// SetClientCertificate sets the tls.Certificate to use for gRPC client
// connections
func (cs *CredentialSupport) SetClientCertificate(cert tls.Certificate) {
	cs.clientCert = cert
}

// GetClientCertificate returns the client certificate of the CredentialSupport
func (cs *CredentialSupport) GetClientCertificate() tls.Certificate {
	return cs.clientCert
}

// GetDeliverServiceCredentials returns GRPC transport credentials for given channel to be used by GRPC
// clients which communicate with ordering service endpoints.
// If the channel isn't found, error is returned.
func (cs *CredentialSupport) GetDeliverServiceCredentials(channelID string) (credentials.TransportCredentials, error) {
	cs.RLock()
	defer cs.RUnlock()

	var creds credentials.TransportCredentials
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cs.clientCert},
	}
	certPool := x509.NewCertPool()

	rootCACerts, exists := cs.OrdererRootCAsByChain[channelID]
	if !exists {
		log.Errorf("Attempted to obtain root CA certs of a non existent channel: %s", channelID)
		return nil, fmt.Errorf("didn't find any root CA certs for channel %s", channelID)
	}

	for _, cert := range rootCACerts {
		block, _ := pem.Decode(cert)
		if block != nil {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				certPool.AddCert(cert)
			} else {
				log.Warnf("Failed to add root cert to credentials (%s)", err)
			}
		} else {
			log.Warn("Failed to add root cert to credentials")
		}
	}
	tlsConfig.RootCAs = certPool
	creds = credentials.NewTLS(tlsConfig)
	return creds, nil
}

// GetPeerCredentials returns GRPC transport credentials for use by GRPC
// clients which communicate with remote peer endpoints.
func (cs *CredentialSupport) GetPeerCredentials() credentials.TransportCredentials {
	var creds credentials.TransportCredentials
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cs.clientCert},
	}
	certPool := x509.NewCertPool()
	// loop through the server root CAs
	//glh
	/*
		roots, _ := cs.GetServerRootCAs()
		for _, root := range roots {
			err := AddPemToCertPool(root, certPool)
			if err != nil {
				log.Warningf("Failed adding certificates to peer's client TLS trust pool: %s", err)
			}
		}
	*/
	tlsConfig.RootCAs = certPool
	creds = credentials.NewTLS(tlsConfig)
	return creds
}

//func getEnv(key, def string) string {
//	val := os.Getenv(key)
//	if len(val) > 0 {
//		return val
//	} else {
//		return def
//	}
//}

// NewClientConnectionWithAddress Returns a new grpc.ClientConn to the given address
func NewClientConnectionWithAddress(peerAddress string, block bool, tslEnabled bool,
	creds credentials.TransportCredentials, ka *KeepaliveOptions) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if ka != nil {
		opts = ClientKeepaliveOptions(ka)
	} else {
		// set to the default options
		opts = ClientKeepaliveOptions(DefaultKeepaliveOptions())
	}

	if tslEnabled {
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	if block {
		opts = append(opts, grpc.WithBlock())
	}
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxRecvMsgSize()),
		grpc.MaxCallSendMsgSize(MaxSendMsgSize())))
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, defaultTimeout)
	conn, err := grpc.DialContext(ctx, peerAddress, opts...)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func InitTLSForShim(key, certStr string) credentials.TransportCredentials {
	var sn string
	priv, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Errorf("failed decoding private key from base64, string: %s, error: %v", key, err)
	}
	pub, err := base64.StdEncoding.DecodeString(certStr)
	if err != nil {
		log.Errorf("failed decoding public key from base64, string: %s, error: %v", certStr, err)
	}
	cert, err := tls.X509KeyPair(pub, priv)
	if err != nil {
		log.Errorf("failed loading certificate: %v", err)
	}
	b, err := ioutil.ReadFile("" /*config.GetPath("peer.tls.rootcert.file")*/)
	if err != nil {
		log.Errorf("failed loading root ca cert: %v", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		log.Errorf("failed to append certificates")
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      cp,
		ServerName:   sn,
	})
}
