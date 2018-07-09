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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
)

type addr struct {
}

func (*addr) Network() string {
	return ""
}

func (*addr) String() string {
	return "1.2.3.4:5000"
}

func TestExtractAddress(t *testing.T) {
	ctx := context.Background()
	assert.Zero(t, ExtractRemoteAddress(ctx))

	ctx = peer.NewContext(ctx, &peer.Peer{
		Addr: &addr{},
	})
	assert.Equal(t, "1.2.3.4:5000", ExtractRemoteAddress(ctx))
}
