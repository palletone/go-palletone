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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// check the defaults
	assert.EqualValues(t, maxRecvMsgSize, MaxRecvMsgSize())
	assert.EqualValues(t, maxSendMsgSize, MaxSendMsgSize())
	assert.EqualValues(t, keepaliveOptions, DefaultKeepaliveOptions())
	assert.EqualValues(t, false, TLSEnabled())
	assert.EqualValues(t, true, configurationCached)

	// set send/recv msg sizes
	size := 10 * 1024 * 1024
	SetMaxRecvMsgSize(size)
	SetMaxSendMsgSize(size)
	assert.EqualValues(t, size, MaxRecvMsgSize())
	assert.EqualValues(t, size, MaxSendMsgSize())

	// reset cache
	configurationCached = false
	viper.Set("peer.tls.enabled", true)
	assert.EqualValues(t, true, TLSEnabled())
	// check that value is cached
	viper.Set("peer.tls.enabled", false)
	assert.NotEqual(t, false, TLSEnabled())
	// reset tls
	configurationCached = false
	viper.Set("peer.tls.enabled", false)
}
