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

package inproccontroller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

func TestSend(t *testing.T) {
	ch := make(chan *pb.ChaincodeMessage)

	stream := newInProcStream(ch, ch)

	//good send (non-blocking send and receive)
	msg := &pb.ChaincodeMessage{}
	go stream.Send(msg)
	msg2, _ := stream.Recv()
	assert.Equal(t, msg, msg2, "send != recv")

	//close the channel
	close(ch)

	//bad send, should panic, unblock and return error
	err := stream.Send(msg)
	assert.NotNil(t, err, "should have errored on panic")
}
