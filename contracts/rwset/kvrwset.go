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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package rwset

import (
	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/core/vmContractPub/flogging"
	"github.com/palletone/go-palletone/dag/modules"
)

var logger = flogging.MustGetLogger("rwset")

type KVRWSet struct {
	Reads            []*KVRead         `protobuf:"bytes,1,rep,name=reads" json:"reads,omitempty"`
	Writes           []*KVWrite        `protobuf:"bytes,3,rep,name=writes" json:"writes,omitempty"`
}

func (m *KVRWSet) Reset()                    { *m = KVRWSet{} }
func (m *KVRWSet) String() string            { return proto.CompactTextString(m) }
func (*KVRWSet) ProtoMessage()               {}
func (*KVRWSet) Descriptor() ([]byte, []int) { return nil, nil }

func (m *KVRWSet) GetReads() []*KVRead {
	if m != nil {
		return m.Reads
	}
	return nil
}

func (m *KVRWSet) GetWrites() []*KVWrite {
	if m != nil {
		return m.Writes
	}
	return nil
}

type KVRead struct {
	Key     string   `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Version *modules.StateVersion `protobuf:"bytes,2,opt,name=version" json:"version,omitempty"`
}

func (m *KVRead) Reset()                    { *m = KVRead{} }
func (m *KVRead) String() string            { return proto.CompactTextString(m) }
func (*KVRead) ProtoMessage()               {}
func (*KVRead) Descriptor() ([]byte, []int) { return nil, nil }

func (m *KVRead) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KVRead) GetVersion() *modules.StateVersion {
	if m != nil {
		return m.Version
	}
	return nil
}

// KVWrite captures a write (update/delete) operation performed during transaction simulation
type KVWrite struct {
	Key      string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	IsDelete bool   `protobuf:"varint,2,opt,name=is_delete,json=isDelete" json:"is_delete,omitempty"`
	Value    []byte `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *KVWrite) Reset()                    { *m = KVWrite{} }
func (m *KVWrite) String() string            { return proto.CompactTextString(m) }
func (*KVWrite) ProtoMessage()               {}
func (*KVWrite) Descriptor() ([]byte, []int) { return nil, nil }

func (m *KVWrite) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KVWrite) GetIsDelete() bool {
	if m != nil {
		return m.IsDelete
	}
	return false
}

func (m *KVWrite) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type Version struct {
	ChainId uint64 `protobuf:"varint,1,opt,name=block_num,json=blockNum" json:"block_num,omitempty"`
	TxNum    uint64 `protobuf:"varint,2,opt,name=tx_num,json=txNum" json:"tx_num,omitempty"`
}

// NewKVRead helps constructing proto message kvrwset.KVRead
func NewKVRead(key string, version *modules.StateVersion) *KVRead {
	return &KVRead{Key: key, Version: version}
}

func newKVWrite(key string, value []byte) *KVWrite {
	return &KVWrite{Key: key, IsDelete: value == nil, Value: value}
}

