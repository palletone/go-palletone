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
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Datasource struct {
	ID                int    `json:"id"`
	OrgID             int    `json:"orgId"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	TypeLogoURL       string `json:"typeLogoUrl"`
	Access            string `json:"access"`
	URL               string `json:"url"`
	Password          string `json:"password"`
	User              string `json:"user"`
	Database          string `json:"database"`
	BasicAuth         bool   `json:"basicAuth"`
	BasicAuthUser     string `json:"basicAuthUser"`
	BasicAuthPassword string `json:"basicAuthPassword"`
	WithCredentials   bool   `json:"withCredentials"`
	IsDefault         bool   `json:"isDefault"`
	JSONData          struct {
		HTTPMethod  string        `json:"httpMethod"`
		KeepCookies []interface{} `json:"keepCookies"`
	} `json:"jsonData"`
	SecureJSONFields struct {
	} `json:"secureJsonFields"`
	Version  int  `json:"version"`
	ReadOnly bool `json:"readOnly"`
}

type DatasourceResp struct {
	Datasource Datasource `json:"datasource"`
	ID         int        `json:"id"`
	Message    string     `json:"message"`
	Name       string     `json:"name"`
}

type QueryResp struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string        `json:"resultType"`
		Result     []interface{} `json:"result"`
	} `json:"data"`
}

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
	go func() {
		http.Handle("/", promhttp.Handler())
		///api/v1/query
		http.Handle("/api/v1/query", db.QueryHandler())

		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", db.config.Host, db.config.Port), nil); err != nil {
			log.Error("Failed to starting prometheus", "err", err)
			os.Exit(1)
		}
	}()

	return nil
}

func (db *Prometheus) QueryHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, db.HandlerForQuery(prometheus.DefaultGatherer, promhttp.HandlerOpts{}),
	)
}

func (db *Prometheus) HandlerForQuery(reg prometheus.Gatherer, opts promhttp.HandlerOpts) http.Handler {
	h := http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		queryurl := req.URL.RawQuery
		log.Debug("======HandlerForQuery======", "query", queryurl)
		respurl, err := url.ParseQuery(queryurl)
		if err != nil {
			log.Error("HandlerForQuery ParseQuery", "err:", err)
			rsp.WriteHeader(400)
		} else {
			rsp.Header().Set("Content-Type", "application/json")
			rsp.Header().Set("Access-Control-Allow-Origin", "*")
			rsp.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Origin")
			rsp.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			rsp.Header().Set("Access-Control-Expose-Headers", "Date")

			resp := QueryResp{}
			resp.Status = "success"
			resp.Data.ResultType = "scalar"

			if r1, err := strconv.ParseFloat(respurl.Get("time"), 64); err != nil {
				rsp.WriteHeader(500)
			} else {
				resp.Data.Result = append(resp.Data.Result, r1)
				resp.Data.Result = append(resp.Data.Result, "2")

				if data, err := json.Marshal(resp); err != nil {
					rsp.WriteHeader(500)
				} else {
					rsp.WriteHeader(200)
					rsp.Write(data)
				}
			}
		}
	})
	return h
}

// Stop implements node.Service, stopping the data collection thread and the connection listener of the dashboard.
func (db *Prometheus) Stop() error {
	// Close the connection listener.

	log.Info("Prometheus stopped")

	return nil
}
