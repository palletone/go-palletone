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

//go:generate yarn --cwd ./assets install
//go:generate yarn --cwd ./assets build
//go:generate go-bindata -nometadata -o assets.go -prefix assets -nocompress -pkg dashboard assets/index.html assets/bundle.js
//go:generate sh -c "sed 's#var _bundleJs#//nolint:misspell\\\n&#' assets.go > assets.go.tmp && mv assets.go.tmp assets.go"
//go:generate sh -c "sed 's#var _indexHtml#//nolint:misspell\\\n&#' assets.go > assets.go.tmp && mv assets.go.tmp assets.go"
//go:generate gofmt -w -s assets.go

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/websocket"
	"net/http"
)

var nextID uint32 // Next connection id

// Prometheus contains the dashboard internals.
type Prometheus struct {
	url string
}

// New creates a new dashboard instance with the given configuration.
func New(url string) (*Prometheus, error) {
	return &Prometheus{url: url}, nil
}

// Protocols is a meaningless implementation of node.Service.
func (db *Prometheus) Protocols() []p2p.Protocol { return nil }

func (db *Prometheus) GenesisHash() common.Hash      { return common.Hash{} }
func (db *Prometheus) CorsProtocols() []p2p.Protocol { return nil }

// APIs is a meaningless implementation of node.Service.
func (db *Prometheus) APIs() []rpc.API { return nil }

// Start implements node.Service, starting the data collection thread and the listening server of the dashboard.
func (db *Prometheus) Start(server *p2p.Server, corss *p2p.Server) error {
	log.Info("Starting dashboard")

	http.Handle("/", promhttp.Handler())
	http.ListenAndServe(":8080", nil)

	return nil
}

// Stop implements node.Service, stopping the data collection thread and the connection listener of the dashboard.
func (db *Prometheus) Stop() error {
	// Close the connection listener.

	log.Info("Prometheus stopped")

	return nil
}

// webHandler handles all non-api requests, simply flattening and returning the dashboard website.
func (db *Prometheus) webHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Request", "URL", r.URL)
}

// apiHandler handles requests for the dashboard.
func (db *Prometheus) apiHandler(conn *websocket.Conn) {

}

// collectData collects the required data to plot on the dashboard.
func (db *Prometheus) collectData() {

}

// collectLogs collects and sends the logs to the active dashboards.
func (db *Prometheus) collectLogs() {

}
