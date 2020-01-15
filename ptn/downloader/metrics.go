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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/palletone/go-palletone/statistics/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	headerReqTimer = metrics.NewRegisteredTimer("ptn/downloader/headers/req", nil)
	bodyReqTimer   = metrics.NewRegisteredTimer("ptn/downloader/bodies/req", nil)
)

var (
	headerInPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:downloader:headers:in",
		Help: "headers in",
	})
	headerDropPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:downloader:headers:drop",
		Help: "headers drop",
	})
	headerTimeoutPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:downloader:headers:timeout",
		Help: "headers timeout",
	})

	bodyInPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:downloader:bodies:in",
		Help: "bodies in",
	})
	bodyDropPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:downloader:bodies:drop",
		Help: "bodies drop",
	})
	bodyTimeoutPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:downloader:bodies:timeout",
		Help: "bodies timeout",
	})
)

func init() {
	prometheus.MustRegister(headerInPrometheus)
	prometheus.MustRegister(headerDropPrometheus)
	prometheus.MustRegister(headerTimeoutPrometheus)

	prometheus.MustRegister(bodyInPrometheus)
	prometheus.MustRegister(bodyDropPrometheus)
	prometheus.MustRegister(bodyTimeoutPrometheus)
}
