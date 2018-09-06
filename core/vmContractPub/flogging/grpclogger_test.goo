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

package flogging

import (
	"testing"

	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/grpclog"
)

// from go-logging memory_test.go
func MemoryRecordN(b *logging.MemoryBackend, n int) *logging.Record {
	node := b.Head()
	for i := 0; i < n; i++ {
		if node == nil {
			break
		}
		node = node.Next()
	}
	if node == nil {
		return nil
	}
	return node.Record
}

func TestGRPCLogger(t *testing.T) {
	initgrpclogger()
	backend := logging.NewMemoryBackend(3)
	logging.SetBackend(backend)
	logging.SetLevel(defaultLevel, "")
	SetModuleLevel(GRPCModuleID, "DEBUG")
	messages := []string{"print test", "printf test", "println test"}
	grpclog.Info(messages[0])
	grpclog.Info(messages[1])
	grpclog.Infoln(messages[2])

	for i, message := range messages {
		assert.Equal(t, message, MemoryRecordN(backend, i).Message())
		t.Log(MemoryRecordN(backend, i).Message())
	}

	// now make sure there's no logging at a level other than DEBUG
	SetModuleLevel(GRPCModuleID, "INFO")
	messages2 := []string{"print test2", "printf test2", "println test2"}
	grpclog.Info(messages2[0])
	grpclog.Info(messages2[1])
	grpclog.Infoln(messages2[2])

	// should still be messages not messages2
	for i, message := range messages {
		assert.Equal(t, message, MemoryRecordN(backend, i).Message())
		t.Log(MemoryRecordN(backend, i).Message())
	}
	// reset flogging
	Reset()
}
