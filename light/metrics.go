// Copyright 2016 The go-ethereum Authors
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

package light

import (
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/statistics/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	miscInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:les:misc:in:packets",
		Help: "les misc in packets",
	})
	miscInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:les:misc:in:traffic",
		Help: "les misc in traffic",
	})

	miscOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:les:misc:out:packets",
		Help: "les misc out packets",
	})
	miscOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:les:misc:out:traffic",
		Help: "les misc out traffic",
	})
)

func init() {
	prometheus.MustRegister(miscInPacketsPrometheus)
	prometheus.MustRegister(miscInTrafficPrometheus)

	prometheus.MustRegister(miscOutPacketsPrometheus)
	prometheus.MustRegister(miscOutTrafficPrometheus)
}

// meteredMsgReadWriter is a wrapper around a p2p.MsgReadWriter, capable of
// accumulating the above defined metrics based on the data stream contents.
type meteredMsgReadWriter struct {
	p2p.MsgReadWriter     // Wrapped message stream to meter
	version           int // Protocol version to select correct meters
}

// newMeteredMsgWriter wraps a p2p MsgReadWriter with metering support. If the
// metrics system is disabled, this function returns the original object.
func newMeteredMsgWriter(rw p2p.MsgReadWriter) p2p.MsgReadWriter {
	if !metrics.Enabled {
		return rw
	}
	return &meteredMsgReadWriter{MsgReadWriter: rw}
}

// Init sets the protocol version used by the stream to know which meters to
// increment in case of overlapping message ids between protocol versions.
func (rw *meteredMsgReadWriter) Init(version int) {
	rw.version = version
}

func (rw *meteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	// Read the message and short circuit in case of an error
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	// Account for the data traffic
	packets, traffic := miscInPacketsPrometheus, miscInTrafficPrometheus
	packets.Add(1)
	traffic.Add(float64(msg.Size))

	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsPrometheus, miscOutTrafficPrometheus
	packets.Add(1)
	traffic.Add(float64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
