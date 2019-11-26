// Copyright 2017 The go-ethereum Authors
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

package prometheus

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Prometheus contains the dashboard internals.
type Prometheus struct {
	config Config
}

// New creates a new dashboard instance with the given configuration.
func New(config Config) (*Prometheus, error) {
	return &Prometheus{config: config}, nil
}

// Protocols is a meaningless implementation of node.Service.
func (db *Prometheus) Protocols() []p2p.Protocol { return nil }

func (db *Prometheus) GenesisHash() common.Hash      { return common.Hash{} }
func (db *Prometheus) CorsProtocols() []p2p.Protocol { return nil }

// APIs is a meaningless implementation of node.Service.
func (db *Prometheus) APIs() []rpc.API { return nil }

// Start implements node.Service, starting the data collection thread and the listening server of the dashboard.
func (db *Prometheus) Start(server *p2p.Server, corss *p2p.Server) error {
	log.Info("Starting Prometheus")

	http.Handle("/", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%d", db.config.Host, db.config.Port), nil)

	return nil
}

// Stop implements node.Service, stopping the data collection thread and the connection listener of the dashboard.
func (db *Prometheus) Stop() error {
	// Close the connection listener.

	log.Info("Prometheus stopped")

	return nil
}
