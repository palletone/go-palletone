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

package ptn

import (
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/statistics/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	propTxnInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:txns:in:packets",
		Help: "ptn prop txns in packets",
	})
	propTxnInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:txns:in:traffic",
		Help: "ptn prop txns in traffic",
	})
	propTxnOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:txns:out:packets",
		Help: "ptn prop txns out packets",
	})
	propTxnOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:txns:out:traffic",
		Help: "ptn prop txns out traffic",
	})

	propHashInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:hashes:in:packets",
		Help: "ptn prop hashes in packets",
	})
	propHashInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:hashes:in:traffic",
		Help: "ptn prop hashes in traffic",
	})
	propHashOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:hashes:out:packets",
		Help: "ptn prop hashes out packets",
	})
	propHashOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:hashes:out:traffic",
		Help: "ptn prop hashes out traffic",
	})

	propBlockInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:blocks:in:packets",
		Help: "ptn prop blocks in packets",
	})
	propBlockInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:blocks:in:traffic",
		Help: "ptn prop blocks in traffic",
	})
	propBlockOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:blocks:out:packets",
		Help: "ptn prop blocks out packets",
	})
	propBlockOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:prop:blocks:out:traffic",
		Help: "ptn prop blocks out traffic",
	})

	reqHeaderInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:headers:in:packets",
		Help: "ptn req headers in packets",
	})
	reqHeaderInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:headers:in:traffic",
		Help: "ptn req headers in traffic",
	})
	reqHeaderOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:headers:out:packets",
		Help: "ptn req headers out packets",
	})
	reqHeaderOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:headers:out:traffic",
		Help: "ptn req headers out traffic",
	})

	reqBodyInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:bodies:in:packets",
		Help: "ptn req bodies in packets",
	})
	reqBodyInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:bodies:in:traffic",
		Help: "ptn req bodies in traffic",
	})
	reqBodyOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:bodies:out:packets",
		Help: "ptn req bodies out packets",
	})
	reqBodyOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:req:bodies:out:traffic",
		Help: "ptn req bodies out traffic",
	})

	//ptn/misc/in/packets
	miscInPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:misc:in:packets",
		Help: "ptn misc in packets",
	})
	miscInTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:misc:in:traffic",
		Help: "ptn misc in traffic",
	})
	miscOutPacketsPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:misc:out:packets",
		Help: "ptn misc out packets",
	})
	miscOutTrafficPrometheus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prometheus:ptn:misc:out:traffic",
		Help: "ptn misc out traffic",
	})
)

func init() {
	prometheus.MustRegister(propTxnInPacketsPrometheus)
	prometheus.MustRegister(propTxnInTrafficPrometheus)
	prometheus.MustRegister(propTxnOutPacketsPrometheus)
	prometheus.MustRegister(propTxnOutTrafficPrometheus)
	prometheus.MustRegister(propHashInPacketsPrometheus)
	prometheus.MustRegister(propHashInTrafficPrometheus)
	prometheus.MustRegister(propHashOutPacketsPrometheus)
	prometheus.MustRegister(propHashOutTrafficPrometheus)
	prometheus.MustRegister(propBlockInPacketsPrometheus)
	prometheus.MustRegister(propBlockInTrafficPrometheus)
	prometheus.MustRegister(propBlockOutPacketsPrometheus)
	prometheus.MustRegister(propBlockOutTrafficPrometheus)
	prometheus.MustRegister(reqHeaderInPacketsPrometheus)
	prometheus.MustRegister(reqHeaderInTrafficPrometheus)
	prometheus.MustRegister(reqHeaderOutPacketsPrometheus)
	prometheus.MustRegister(reqHeaderOutTrafficPrometheus)
	prometheus.MustRegister(reqBodyInPacketsPrometheus)
	prometheus.MustRegister(reqBodyInTrafficPrometheus)
	prometheus.MustRegister(reqBodyOutPacketsPrometheus)
	prometheus.MustRegister(reqBodyOutTrafficPrometheus)

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

// newPrometheusedMsgWriter wraps a p2p MsgReadWriter with metering support. If the
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
	switch {
	case msg.Code == BlockHeadersMsg:
		packets, traffic = reqHeaderInPacketsPrometheus, reqHeaderInTrafficPrometheus
	case msg.Code == BlockBodiesMsg:
		packets, traffic = reqBodyInPacketsPrometheus, reqBodyInTrafficPrometheus

	//case msg.Code == NodeDataMsg:
	//packets, traffic = reqStateInPacketsPrometheus, reqStateInTrafficPrometheus
	//case msg.Code == ReceiptsMsg:
	//	packets, traffic = reqReceiptInPacketsPrometheus, reqReceiptInTrafficPrometheus

	case msg.Code == NewBlockHashesMsg:
		packets, traffic = propHashInPacketsPrometheus, propHashInTrafficPrometheus
	case msg.Code == NewBlockMsg:
		packets, traffic = propBlockInPacketsPrometheus, propBlockInTrafficPrometheus
	case msg.Code == TxMsg:
		packets, traffic = propTxnInPacketsPrometheus, propTxnInTrafficPrometheus
	}
	packets.Add(1)
	traffic.Add(float64(msg.Size))

	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsPrometheus, miscOutTrafficPrometheus
	switch {
	case msg.Code == BlockHeadersMsg:
		packets, traffic = reqHeaderOutPacketsPrometheus, reqHeaderOutTrafficPrometheus
	case msg.Code == BlockBodiesMsg:
		packets, traffic = reqBodyOutPacketsPrometheus, reqBodyOutTrafficPrometheus

	//case msg.Code == NodeDataMsg:
	//	packets, traffic = reqStateOutPacketsPrometheus, reqStateOutTrafficPrometheus
	//case msg.Code == ReceiptsMsg:
	//	packets, traffic = reqReceiptOutPacketsPrometheus, reqReceiptOutTrafficPrometheus

	case msg.Code == NewBlockHashesMsg:
		packets, traffic = propHashOutPacketsPrometheus, propHashOutTrafficPrometheus
	case msg.Code == NewBlockMsg:
		packets, traffic = propBlockOutPacketsPrometheus, propBlockOutTrafficPrometheus
	case msg.Code == TxMsg:
		packets, traffic = propTxnOutPacketsPrometheus, propTxnOutTrafficPrometheus
	}
	packets.Add(1)
	traffic.Add(float64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
