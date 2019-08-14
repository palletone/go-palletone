// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/internal/debug"
	flock "github.com/prometheus/tsdb/fileutil"
)

// Node is a container on which services can be registered.
// 在区块链网络中，节点被定义为一个可以注册多种服务的容器。
// node也可以理解为一个进程，整个区块链网络就是由运行在世界各地的很多中类型的node组成。
// 服务是一套独立的协议，可以基于P2P网络和PRC通信来提供具体的服务内容。
// 一个典型的node就是一个p2p的节点。根据节点的类型，运行了不同p2p网络协议，
// 运行了不同的业务层协议(以区别网络层协议。 参考p2p peer中的Protocol接口)。
type Node struct {
	eventmux *event.TypeMux    // Event multiplexer used between the services of a stack
	config   *Config           // 节点的配置信息，副本
	accman   *accounts.Manager // 账户管理器

	ephemeralKeystore string         // if non-empty, the key directory that will be removed by Stop
	instanceDirLock   flock.Releaser // prevents concurrent use of instance directory

	serverConfig p2p.Config
	server       *p2p.Server // Currently running P2P networking layer
	corsserver   *p2p.Server

	serviceFuncs []ServiceConstructor     // Service constructors (in dependency order)
	services     map[reflect.Type]Service // Currently running services

	// --- RPC 相关对象 -- Start
	// RPC 提供的 API
	rpcAPIs []rpc.API // List of APIs currently provided by the node
	// InProc RPC 消息处理
	inprocHandler *rpc.Server // In-process RPC request handler to process the API requests

	// IPC 端点
	ipcEndpoint string // IPC endpoint to listen at (empty = IPC disabled)
	// IPC API 服务监听
	ipcListener net.Listener // IPC RPC listener socket to serve API requests
	// IPC API 消息处理
	ipcHandler *rpc.Server // IPC RPC request handler to process the API requests

	// HTTP 端点
	httpEndpoint string // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	// HTTP 白名单
	httpWhitelist []string // HTTP RPC modules to allow through this endpoint
	// HTTP API 服务监听
	httpListener net.Listener // HTTP RPC listener socket to server API requests
	// HTTP API 消息处理
	httpHandler *rpc.Server // HTTP RPC request handler to process the API requests

	// Websocket 端点
	wsEndpoint string // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	// Websocket API 服务监听
	wsListener net.Listener // Websocket RPC listener socket to server API requests
	// Websocket API 消息处理
	wsHandler *rpc.Server // Websocket RPC request handler to process the API requests
	// --- RPC 相关对象 -- End

	// 节点的等待终止通知的channel, node.New()时不创建，node.Start()时创建
	stop    chan struct{} // Channel to wait for termination notifications
	lock    sync.RWMutex
	CacheDb palletcache.ICache // global cache for use by other modules
	//log log.ILogger
	//for genesis 2018-8-14
	//dbpath string
}

// New creates a new P2P node, ready for protocol registration.
// 创建一个 p2p 节点, 为协议(即service)注册做准备；
// 节点的初始化并不依赖其他的外部组件， 只依赖一个Config对象。
func New(conf *Config) (*Node, error) {
	// Copy config and resolve the datadir so future changes to the current
	// working directory don't affect the node.
	confCopy := *conf
	conf = &confCopy
	// 把datadir转成绝对路径
	//if conf.DataDir != "" {
	//	absdatadir, err := filepath.Abs(conf.DataDir)
	//	if err != nil {
	//		return nil, err
	//	}
	//	conf.DataDir = absdatadir
	//}
	// Ensure that the instance name doesn't cause weird conflicts with
	// other files in the data directory.
	if strings.ContainsAny(conf.Name, `/\`) {
		return nil, errors.New(`Config.Name must not contain '/' or '\'`)
	}
	if conf.Name == datadirDefaultKeyStore {
		return nil, errors.New(`Config.Name cannot be "` + datadirDefaultKeyStore + `"`)
	}
	if strings.HasSuffix(conf.Name, ".ipc") {
		return nil, errors.New(`Config.Name cannot end in ".ipc"`)
	}
	// Ensure that the AccountManager method works before the node has started.
	// We rely on this in cmd/gptn.
	// 调用makeAccountManager()初始化账号管理器和临时Keystore
	am, ephemeralKeystore, err := makeAccountManager(conf)
	if err != nil {
		return nil, err
	}

	// Note: any interaction with Config that would create/touch files
	// in the data directory or instance directory is delayed until Start.
	return &Node{
		accman:            am,
		ephemeralKeystore: ephemeralKeystore,
		config:            conf,
		serviceFuncs:      []ServiceConstructor{},
		ipcEndpoint:       conf.IPCEndpoint(),
		httpEndpoint:      conf.HTTPEndpoint(),
		wsEndpoint:        conf.WSEndpoint(),
		eventmux:          new(event.TypeMux),
	}, nil
}

// Register injects a new service into the node's stack. The service created by
// the passed constructor must be unique in its type with regard to sibling ones.
func (n *Node) Register(constructor ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.server != nil {
		return ErrNodeRunning
	}
	// 把Service的构造函数放进node的serviceFuncs数组。等到启动结点的时候才真正调用构造函数创建Service。
	n.serviceFuncs = append(n.serviceFuncs, constructor)
	return nil
}

// Start create a live P2P node and starts running it.
// 启动P2P服务，并且依次启动Register的各个serviceFuncs相关服务。
func (n *Node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Short circuit if the node's already running
	if n.server != nil {
		return ErrNodeRunning
	}
	// 创建一个文件夹用于存储节点相关数据
	if err := n.openDataDir(); err != nil {
		return err
	}

	// Initialize the p2p server. This creates the node key and
	// discovery databases.
	n.serverConfig = n.config.P2P
	n.serverConfig.PrivateKey = n.config.NodeKey()
	n.serverConfig.Name = n.config.NodeName()

	if n.serverConfig.StaticNodes == nil {
		n.serverConfig.StaticNodes = n.config.StaticNodes()
	}

	if n.serverConfig.TrustedNodes == nil {
		n.serverConfig.TrustedNodes = n.config.TrustedNodes()
	}
	if n.serverConfig.NodeDatabase == "" {
		n.serverConfig.NodeDatabase = n.config.NodeDB()
	}

	corss := &p2p.Server{Config: p2p.GetCorsConfig(n.serverConfig)}

	running := &p2p.Server{Config: n.serverConfig}
	log.Info("Starting peer-to-peer node", "instance", n.serverConfig.Name)

	// Otherwise copy and specialize the P2P configuration
	services := make(map[reflect.Type]Service)
	for _, constructor := range n.serviceFuncs {
		// Create a new context for the particular service
		// 为每个service 分别新建一个ServiceContext 结构
		ctx := &ServiceContext{
			config:         n.config,
			services:       make(map[reflect.Type]Service),
			EventMux:       n.eventmux,
			AccountManager: n.accman,
		}
		//重新拷贝一下services 变量的所有成员给ctx.services
		for kind, s := range services { // copy needed for threaded access
			ctx.services[kind] = s
		}
		// Construct and save the service
		service, err := constructor(ctx)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := services[kind]; exists {
			return &DuplicateServiceError{Kind: kind}
		}
		// 记录一下，保存到services map里面。
		services[kind] = service
	}

	// Gather the protocols and start the freshly assembled P2P server
	for _, service := range services {
		running.Protocols = append(running.Protocols, service.Protocols()...)
		corss.Protocols = append(corss.Protocols, service.CorsProtocols()...)

		if !common.EmptyHash(service.GenesisHash()) && len(configure.GenesisHash) == 0 {
			log.Debug("Node Start", "service.GenesisHash", service.GenesisHash())
			configure.GenesisHash = service.GenesisHash().Bytes()
		}

	}
	log.Debug("Node Start", "len(running.Protocols)", len(running.Protocols), "len(corss.Protocols)", len(corss.Protocols))

	//for /*kind*/ _, service := range services {
	//	if pallet, ok := service.(*ptn.PalletOne); ok {
	//		pallet.Dag().GetGenesisUnit()
	//		if unit, err := pallet.Dag().GetGenesisUnit(); err != nil {
	//			log.Debug("===GetGenesisUnit===", "err", err)
	//		} else {
	//			log.Debug("===GetGenesisUnit===", "genesis hash", unit.Hash())
	//		}
	//	}
	//}

	if err := corss.Start(); err != nil {
		return convertFileLockError(err)
	}

	if err := running.Start(); err != nil {
		return convertFileLockError(err)
	}

	// Start each of the services
	started := []reflect.Type{}
	for kind, service := range services {
		// Start the next service, stopping all previous upon failure
		if err := service.Start(running, corss); err != nil {
			for _, kind := range started {
				services[kind].Stop()
			}
			running.Stop()
			corss.Stop()
			return err
		}
		// Mark the service started for potential cleanup
		started = append(started, kind)
	}

	// Lastly start the configured RPC interfaces
	if err := n.startRPC(services); err != nil {
		for _, service := range services {
			service.Stop()
		}
		running.Stop()
		corss.Stop()
		return err
	}
	// Finish initializing the startup
	n.services = services
	n.server = running
	n.corsserver = corss
	n.stop = make(chan struct{})

	return nil
}

func (n *Node) openDataDir() error {
	if n.config.DataDir == "" {
		return nil // ephemeral
	}

	instdir := filepath.Join(n.config.DataDir, n.config.name())
	if err := os.MkdirAll(instdir, 0700); err != nil {
		return err
	}
	// Lock the instance directory to prevent concurrent use by another instance as well as
	// accidental use of the instance directory as a database.
	release, _, err := flock.Flock(filepath.Join(instdir, "LOCK"))
	if err != nil {
		return convertFileLockError(err)
	}
	n.instanceDirLock = release
	return nil
}

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
// 启动4项RPC服务. 每种RPC服务都需要提供一个handler，
// 如果任何一个RPC启动失败，结束所有RPC endpoint，并返回err。
// 另外除了InProc之外，其他3种服务还需要启动一个server来监听外部连接请求。
func (n *Node) startRPC(services map[reflect.Type]Service) error {
	// Gather all the possible APIs to surface
	// 收集所有可能的API接口, 并把收集的APIs传给 RPC endpoint。
	apis := n.apis()
	for _, service := range services {
		apis = append(apis, service.APIs()...)
	}

	// Start the various API endpoints, terminating all in case of errors
	// 1. 启动 InProc，用于进程内部的通信，严格来说这种不能算是RPC, 出于架构上的统一
	if err := n.startInProc(apis); err != nil {
		log.Error("startRPC startInProc err:", err.Error())
		return err
	}

	// 2. 启动 IPC，用于节点内进程间的通信
	if err := n.startIPC(apis); err != nil {
		log.Error("startRPC startIPC err:", err.Error())
		n.stopInProc()
		return err
	}
	// 3. 启动 HTTP，用于 HTTP 的交互通信
	if err := n.startHTTP(n.httpEndpoint, apis, n.config.HTTPModules, n.config.HTTPCors,
		n.config.HTTPVirtualHosts); err != nil {
		log.Error("startRPC startHTTP err:", err.Error())
		n.stopIPC()
		n.stopInProc()
		return err
	}
	// 4. 启动 WebSocket，用于浏览器与服务器的 TCP 全双工通信
	if err := n.startWS(n.wsEndpoint, apis, n.config.WSModules, n.config.WSOrigins, n.config.WSExposeAll); err != nil {
		n.stopHTTP()
		n.stopIPC()
		n.stopInProc()
		return err
	}
	// All API endpoints started successfully
	n.rpcAPIs = apis
	return nil
}

// startInProc initializes an in-process RPC endpoint.
func (n *Node) startInProc(apis []rpc.API) error {
	// Register all the APIs exposed by the services
	// 严格来说 InProc 不能算是RPC，不过出于架构上的统一，也为这种调用方式配置一个handler
	handler := rpc.NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
		log.Debug("InProc registered", "service", api.Service, "namespace", api.Namespace)
	}
	n.inprocHandler = handler
	return nil
}

// stopInProc terminates the in-process RPC endpoint.
func (n *Node) stopInProc() {
	if n.inprocHandler != nil {
		n.inprocHandler.Stop()
		n.inprocHandler = nil
	}
}

// startIPC initializes and starts the IPC RPC endpoint.
func (n *Node) startIPC(apis []rpc.API) error {
	if n.ipcEndpoint == "" {
		log.Info("startIPC ipcEndpoint is null")
		return nil // IPC disabled.
	}
	listener, handler, err := rpc.StartIPCEndpoint(n.ipcEndpoint, apis)
	if err != nil {
		log.Info("startIPC StartIPCEndpoint err:", err.Error())
		return err
	}
	n.ipcListener = listener
	n.ipcHandler = handler
	log.Info("IPC endpoint opened", "url", n.ipcEndpoint)
	return nil
}

// stopIPC terminates the IPC RPC endpoint.
func (n *Node) stopIPC() {
	if n.ipcListener != nil {
		n.ipcListener.Close()
		n.ipcListener = nil

		log.Info("IPC endpoint closed", "endpoint", n.ipcEndpoint)
	}
	if n.ipcHandler != nil {
		n.ipcHandler.Stop()
		n.ipcHandler = nil
	}
}

// startHTTP initializes and starts the HTTP RPC endpoint.
func (n *Node) startHTTP(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		log.Info("HTTP endpoint is null")
		return nil
	}
	listener, handler, err := rpc.StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts)
	if err != nil {
		log.Info("HTTP endpoint StartHTTPEndpoint err:", err)
		return err
	}
	log.Info("HTTP endpoint opened", "url", fmt.Sprintf("http://%s", endpoint), "cors",
		strings.Join(cors, ","), "vhosts", strings.Join(vhosts, ","))
	// All listeners booted successfully
	n.httpEndpoint = endpoint
	n.httpListener = listener
	n.httpHandler = handler

	return nil
}

// stopHTTP terminates the HTTP RPC endpoint.
func (n *Node) stopHTTP() {
	if n.httpListener != nil {
		n.httpListener.Close()
		n.httpListener = nil

		log.Info("HTTP endpoint closed", "url", fmt.Sprintf("http://%s", n.httpEndpoint))
	}
	if n.httpHandler != nil {
		n.httpHandler.Stop()
		n.httpHandler = nil
	}
}

// startWS initializes and starts the websocket RPC endpoint.
func (n *Node) startWS(endpoint string, apis []rpc.API, modules []string, wsOrigins []string, exposeAll bool) error {
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := rpc.StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
	if err != nil {
		return err
	}
	log.Info("WebSocket endpoint opened", "url", fmt.Sprintf("ws://%s", listener.Addr()))
	// All listeners booted successfully
	n.wsEndpoint = endpoint
	n.wsListener = listener
	n.wsHandler = handler

	return nil
}

// stopWS terminates the websocket RPC endpoint.
func (n *Node) stopWS() {
	if n.wsListener != nil {
		n.wsListener.Close()
		n.wsListener = nil

		log.Info("WebSocket endpoint closed", "url", fmt.Sprintf("ws://%s", n.wsEndpoint))
	}
	if n.wsHandler != nil {
		n.wsHandler.Stop()
		n.wsHandler = nil
	}
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (n *Node) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Short circuit if the node's not running
	if n.server == nil {
		return ErrNodeStopped
	}

	// Terminate the API, services and the p2p server.
	n.stopWS()
	n.stopHTTP()
	n.stopIPC()
	n.rpcAPIs = nil
	failure := &StopError{
		Services: make(map[reflect.Type]error),
	}
	for kind, service := range n.services {
		if err := service.Stop(); err != nil {
			failure.Services[kind] = err
		}
	}

	n.corsserver.Stop()
	n.corsserver = nil
	n.server.Stop()
	n.services = nil
	n.server = nil

	// Release instance directory lock.
	if n.instanceDirLock != nil {
		if err := n.instanceDirLock.Release(); err != nil {
			log.Error("Can't release datadir lock", "err", err)
		}
		n.instanceDirLock = nil
	}

	// unblock n.Wait
	close(n.stop)

	// Remove the keystore if it was created ephemerally.
	var keystoreErr error
	if n.ephemeralKeystore != "" {
		keystoreErr = os.RemoveAll(n.ephemeralKeystore)
	}

	if len(failure.Services) > 0 {
		return failure
	}
	if keystoreErr != nil {
		return keystoreErr
	}
	return nil
}

// Wait blocks the thread until the node is stopped. If the node is not running
// at the time of invocation, the method immediately returns.
// 让主线程进入阻塞状态，保持进程不退出，直到从channel中收到stop消息。
func (n *Node) Wait() {
	n.lock.RLock()
	if n.server == nil && n.corsserver == nil {
		n.lock.RUnlock()
		return
	}
	stop := n.stop
	n.lock.RUnlock()

	<-stop
}

// Restart terminates a running node and boots up a new one in its place. If the
// node isn't running, an error is returned.
func (n *Node) Restart() error {
	if err := n.Stop(); err != nil {
		return err
	}
	if err := n.Start(); err != nil {
		return err
	}
	return nil
}

// Attach creates an RPC client attached to an in-process API handler.
func (n *Node) Attach() (*rpc.Client, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	if n.server == nil {
		return nil, ErrNodeStopped
	}
	return rpc.DialInProc(n.inprocHandler), nil
}

// RPCHandler returns the in-process RPC request handler.
func (n *Node) RPCHandler() (*rpc.Server, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	if n.inprocHandler == nil {
		return nil, ErrNodeStopped
	}
	return n.inprocHandler, nil
}

// Server retrieves the currently running P2P network layer. This method is meant
// only to inspect fields of the currently running server, life cycle management
// should be left to this Node entity.
func (n *Node) Server() *p2p.Server {
	n.lock.RLock()
	defer n.lock.RUnlock()

	return n.server
}

func (n *Node) CorsServer() *p2p.Server {
	return n.corsserver
}

// Service retrieves a currently running service registered of a specific type.
func (n *Node) Service(service interface{}) error {
	n.lock.RLock()
	defer n.lock.RUnlock()

	// Short circuit if the node's not running
	if n.server == nil {
		return ErrNodeStopped
	}
	// Otherwise try to find the service to return
	element := reflect.ValueOf(service).Elem()
	if running, ok := n.services[element.Type()]; ok {
		element.Set(reflect.ValueOf(running))
		return nil
	}
	return ErrServiceUnknown
}

// DataDir retrieves the current datadir used by the protocol stack.
func (n *Node) DataDir() string {
	return n.config.DataDir
}

// author Albert·Gou
func (n *Node) ListenAddr() string {
	return n.config.P2P.ListenAddr
}

// author Albert·Gou
func (n *Node) Config() *Config {
	return n.config
}

// InstanceDir retrieves the instance directory used by the protocol stack.
func (n *Node) InstanceDir() string {
	return n.config.instanceDir()
}

// AccountManager retrieves the account manager used by the protocol stack.
func (n *Node) AccountManager() *accounts.Manager {
	return n.accman
}

// IPCEndpoint retrieves the current IPC endpoint used by the protocol stack.
func (n *Node) IPCEndpoint() string {
	return n.ipcEndpoint
}

// HTTPEndpoint retrieves the current HTTP endpoint used by the protocol stack.
func (n *Node) HTTPEndpoint() string {
	return n.httpEndpoint
}

// WSEndpoint retrieves the current WS endpoint used by the protocol stack.
func (n *Node) WSEndpoint() string {
	return n.wsEndpoint
}

// EventMux retrieves the event multiplexer used by all the network services in
// the current protocol stack.
func (n *Node) EventMux() *event.TypeMux {
	return n.eventmux
}

// OpenDatabase opens an existing database with the given name (or creates one if no
// previous can be found) from within the node's instance directory. If the node is
// ephemeral, a memory database is returned.
func (n *Node) OpenDatabase(path string, cache, handles int) (ptndb.Database, error) {
	if n.config.DataDir == "" {
		log.Debug("Open a memery database.")
		return ptndb.NewMemDatabase()
	}
	log.Debug("Open a leveldb, path:", "info", path)
	//return ptndb.NewLDBDatabase(n.config.resolvePath(name), cache, handles)
	return storage.Init(path, cache, handles)
}

// ResolvePath returns the absolute path of a resource in the instance directory.
func (n *Node) ResolvePath(x string) string {
	return n.config.resolvePath(x)
}

// apis returns the collection of RPC descriptors this node offers.
func (n *Node) apis() []rpc.API {
	return []rpc.API{
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(n),
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPublicAdminAPI(n),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   debug.Handler,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(n),
			Public:    true,
		}, {
			Namespace: "web3",
			Version:   "1.0",
			Service:   NewPublicWeb3API(n),
			Public:    true,
		},
	}
}

// @author Albert·Gou
func (n *Node) GetKeyStore() *keystore.KeyStore {
	return n.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}
