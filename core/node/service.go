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
	"reflect"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/accounts"
)

// ServiceContext is a collection of service independent options inherited from
// the protocol stack, that is passed to all constructors to be optionally used;
// as well as utility methods to operate on the service environment.
// ServiceContext 是一些服务的集合，这些服务是从结点（或者叫协议栈）继承过来的，和具体服务无关，
// 具体服务的构造函数都被传递ServiceContext类型的参数以供选择使用;
// ServiceContext 也是在服务环境上运行的实用方法。
type ServiceContext struct {
	config         *Config
	services       map[reflect.Type]Service // Index of the already constructed services
	EventMux       *event.TypeMux           // Event multiplexer used for decoupled notifications
	AccountManager *accounts.Manager        // Account manager created by the node.
}

// OpenDatabase opens an existing database with the given name (or creates one
// if no previous can be found) from within the node's data directory. If the
// node is an ephemeral one, a memory database is returned.
func (ctx *ServiceContext) OpenDatabase(name string, cache int, handles int) (ptndb.Database, error) {
	if ctx.config.DataDir == "" {
		return ptndb.NewMemDatabase()
	}
	db, err := ptndb.NewLDBDatabase(ctx.config.resolvePath(name), cache, handles)
	if err != nil {
		return nil, err
	}

	return db, nil
}

//Get leveldb file path for initdb storage Init
//func (ctx *ServiceContext) DatabasePath(name string) string {
//	return ctx.config.resolvePath(name)
//}

// ResolvePath resolves a user path into the data directory if that was relative
// and if the user actually uses persistent storage. It will return an empty string
// for emphemeral storage and the user's own input for absolute paths.
func (ctx *ServiceContext) ResolvePath(path string) string {
	return ctx.config.resolvePath(path)
}

// Service retrieves a currently running service registered of a specific type.
func (ctx *ServiceContext) Service(service interface{}) error {
	element := reflect.ValueOf(service).Elem()
	if running, ok := ctx.services[element.Type()]; ok {
		element.Set(reflect.ValueOf(running))
		return nil
	}
	return ErrServiceUnknown
}

// ServiceConstructor is the function signature of the constructors needed to be
// registered for service instantiation.
// ServiceConstructor是服务实例化时需要的构造函数的函数签名；
// 函数签名，如果两个函数的参数列表和返回值列表的变量类型能一一对应，那么这两个函数就有相同的签名。
type ServiceConstructor func(ctx *ServiceContext) (Service, error)

// Service is an individual protocol that can be registered into a node.
//
// Notes:
//
// • Service life-cycle management is delegated to the node. The service is allowed to
// initialize itself upon creation, but no goroutines should be spun up outside of the
// Start method.
//
// • Restart logic is not required as the node will create a fresh instance
// every time a service is started.
// Service是一个接口，定义了4个需要实现的函数。任何实现了这4个方法的类型，都可以称之为一个Service。
// 服务的生命周期管理已经代理给node管理。该服务允许在创建时自动初始化，但是在Start方法之外不应该启动goroutines。
// 重新启动逻辑不是必需的，因为节点将在每次启动服务时创建一个新的实例。
type Service interface {
	// Protocols retrieves the P2P protocols the service wishes to start.
	// 返回 service 要启动的 P2P 协议列表
	Protocols() []p2p.Protocol

	CorsProtocols() []p2p.Protocol

	// APIs retrieves the list of RPC descriptors the service provides
	// 返回本 service 能提供的 RPC API 接口
	APIs() []rpc.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	// start 方法是在所有服务构建之后和网络层初始化完后调用，
	// 在 start 方法用于开启大量 service 所需要的任何 goroutines
	Start(server *p2p.Server, corss *p2p.Server) error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	// 停止service所有的goroutines，并阻塞线程直到所有goroutines都终止
	Stop() error

	//GenesisHash
	GenesisHash() common.Hash
}
