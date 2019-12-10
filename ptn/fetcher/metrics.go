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

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/palletone/go-palletone/statistics/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	propAnnounceOutTimer  = metrics.NewRegisteredTimer("ptn/fetcher/prop/announces/out", nil)
	propBroadcastOutTimer = metrics.NewRegisteredTimer("ptn/fetcher/prop/broadcasts/out", nil)
)

var (
	propAnnounceInPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:prop:announces:in",
		Help: "fetcher prop announces in",
	})
	propAnnounceDropPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:prop:announces:drop",
		Help: "fetcher prop announces drop",
	})
	propAnnounceDOSPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:prop:announces:dos",
		Help: "fetcher prop announces dos",
	})

	propBroadcastInPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:prop:broadcasts:in",
		Help: "fetcher prop broadcasts in",
	})
	propBroadcastDropPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:prop:broadcasts:drop",
		Help: "fetcher prop broadcasts drop",
	})
	propBroadcastDOSPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:prop:broadcasts:dos",
		Help: "fetcher prop broadcasts dos",
	})

	headerFetchPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:headers",
		Help: "fetcher headers",
	})
	bodyFetchPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:bodies",
		Help: "fetcher bodies",
	})

	headerFilterInPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:filter:headers:in",
		Help: "fetcher filter headers in",
	})
	headerFilterOutPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:filter:headers:out",
		Help: "fetcher filter headers out",
	})
	bodyFilterInPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:filter:bodies:in",
		Help: "fetcher filter bodies in",
	})
	bodyFilterOutPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:fetcher:filter:bodies:out",
		Help: "fetcher filter bodies out",
	})
)

func init() {
	prometheus.MustRegister(propAnnounceInPrometheus)
	prometheus.MustRegister(propAnnounceDropPrometheus)
	prometheus.MustRegister(propAnnounceDOSPrometheus)

	prometheus.MustRegister(propBroadcastInPrometheus)
	prometheus.MustRegister(propBroadcastDropPrometheus)
	prometheus.MustRegister(propBroadcastDOSPrometheus)

	prometheus.MustRegister(headerFetchPrometheus)
	prometheus.MustRegister(bodyFetchPrometheus)

	prometheus.MustRegister(headerFilterInPrometheus)
	prometheus.MustRegister(headerFilterOutPrometheus)
	prometheus.MustRegister(bodyFilterInPrometheus)
	prometheus.MustRegister(bodyFilterOutPrometheus)
}
