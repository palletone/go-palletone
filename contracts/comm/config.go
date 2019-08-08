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
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	// Is the configuration cached?
	configurationCached = false
	// Is TLS enabled
	tlsEnabled bool
	// Max send and receive bytes for grpc clients and servers
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
	// Default peer keepalive options
	keepaliveOptions = &KeepaliveOptions{
		ClientInterval:    time.Duration(1) * time.Minute,  // 1 min
		ClientTimeout:     time.Duration(20) * time.Second, // 20 sec - gRPC default
		ServerInterval:    time.Duration(2) * time.Hour,    // 2 hours - gRPC default
		ServerTimeout:     time.Duration(20) * time.Second, // 20 sec - gRPC default
		ServerMinInterval: time.Duration(10) * time.Minute, //1 match ClientInterval
	}
	// strong TLS cipher suites
	//tlsCipherSuites = []uint16{
	//	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	//	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	//	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	//	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	//	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	//	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	//}
)

// ServerConfig defines the parameters for configuring a GRPCServer instance
type ServerConfig struct {
	// SecOpts defines the security parameters
	SecOpts *SecureOptions
	// KaOpts defines the keepalive parameters
	KaOpts *KeepaliveOptions
}

// ClientConfig defines the parameters for configuring a GRPCClient instance
type ClientConfig struct {
	// SecOpts defines the security parameters
	SecOpts *SecureOptions
	// KaOpts defines the keepalive parameters
	KaOpts *KeepaliveOptions
	// Timeout specifies how long the client will block when attempting to
	// establish a connection
	Timeout time.Duration
}

// SecureOptions defines the security parameters (e.g. TLS) for a
// GRPCServer instance
type SecureOptions struct {
	// PEM-encoded X509 public key to be used for TLS communication
	Certificate []byte
	// PEM-encoded private key to be used for TLS communication
	Key []byte
	// Set of PEM-encoded X509 certificate authorities used by clients to
	// verify server certificates
	ServerRootCAs [][]byte
	// Set of PEM-encoded X509 certificate authorities used by servers to
	// verify client certificates
	ClientRootCAs [][]byte
	// Whether or not to use TLS for communication
	UseTLS bool
	// Whether or not TLS client must present certificates for authentication
	RequireClientCert bool
	// CipherSuites is a list of supported cipher suites for TLS
	CipherSuites []uint16
}

// KeepAliveOptions is used to set the gRPC keepalive settings for both
// clients and servers
type KeepaliveOptions struct {
	// ClientInterval is the duration after which if the client does not see
	// any activity from the server it pings the server to see if it is alive
	//ClientInterval是持续时间之后的持续时间，如果客户端没有看到来自服务器的任何活动，
	// 它会ping服务器以查看它是否处于活动状态
	ClientInterval time.Duration
	// ClientTimeout is the duration the client waits for a response
	// from the server after sending a ping before closing the connection
	//ClientTimeout是客户端在关闭连接之前发送ping之后等待服务器响应的持续时间
	ClientTimeout time.Duration
	// ServerInterval is the duration after which if the server does not see
	// any activity from the client it pings the client to see if it is alive
	//ServerInterval是持续时间之后的持续时间，如果服务器没有看到来自客户端的任何活动，
	// 它会ping客户端以查看它是否处于活动状态
	ServerInterval time.Duration
	// ServerTimeout is the duration the server waits for a response
	// from the client after sending a ping before closing the connection
	//ServerTimeout是服务器在关闭连接之前发送ping之后等待客户端响应的持续时间
	ServerTimeout time.Duration
	// ServerMinInterval is the minimum permitted time between client pings.
	// If clients send pings more frequently, the server will disconnect them
	//ServerMinInterval是客户端ping之间允许的最短时间。 如果客户端更频繁地发送ping，服务器将断开它们
	ServerMinInterval time.Duration
}

// cacheConfiguration caches common package scoped variables
func cacheConfiguration() {
	if !configurationCached {
		tlsEnabled = viper.GetBool("peer.tls.enabled")
		configurationCached = true
	}
}

// TLSEnabled return cached value for "peer.tls.enabled" configuration value
func TLSEnabled() bool {
	if !configurationCached {
		cacheConfiguration()
	}
	return tlsEnabled
}

// MaxRecvMsgSize returns the maximum message size in bytes that gRPC clients
// and servers can receive
func MaxRecvMsgSize() int {
	return maxRecvMsgSize
}

// SetMaxRecvMsgSize sets the maximum message size in bytes that gRPC clients
// and servers can receive
func SetMaxRecvMsgSize(size int) {
	maxRecvMsgSize = size
}

// MaxSendMsgSize returns the maximum message size in bytes that gRPC clients
// and servers can send
func MaxSendMsgSize() int {
	return maxSendMsgSize
}

// SetMaxSendMsgSize sets the maximum message size in bytes that gRPC clients
// and servers can send
func SetMaxSendMsgSize(size int) {
	maxSendMsgSize = size
}

// DefaultKeepaliveOptions returns sane default keepalive settings for gRPC
// servers and clients
func DefaultKeepaliveOptions() *KeepaliveOptions {
	return keepaliveOptions
}

// ServerKeepaliveOptions returns gRPC keepalive options for server.  If
// opts is nil, the default keepalive options are returned
func ServerKeepaliveOptions(ka *KeepaliveOptions) []grpc.ServerOption {
	// use default keepalive options if nil
	if ka == nil {
		ka = keepaliveOptions
	}
	var serverOpts []grpc.ServerOption
	kap := keepalive.ServerParameters{
		Time:    ka.ServerInterval,
		Timeout: ka.ServerTimeout,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveParams(kap))
	kep := keepalive.EnforcementPolicy{
		MinTime: ka.ServerMinInterval,
		// allow keepalive w/o rpc
		PermitWithoutStream: true, //true,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveEnforcementPolicy(kep))
	return serverOpts
}

// ClientKeepaliveOptions returns gRPC keepalive options for clients.  If
// opts is nil, the default keepalive options are returned
func ClientKeepaliveOptions(ka *KeepaliveOptions) []grpc.DialOption {
	// use default keepalive options if nil
	if ka == nil {
		ka = keepaliveOptions
	}

	var dialOpts []grpc.DialOption
	kap := keepalive.ClientParameters{
		Time:                ka.ClientInterval,
		Timeout:             ka.ClientTimeout,
		PermitWithoutStream: true, //true,
	}
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(kap))
	return dialOpts
}
