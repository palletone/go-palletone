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


package peer

import (
	"fmt"

	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

//MockPeerCCSupport provides CC support for peer interfaces.
type MockPeerCCSupport struct {
	ccStream map[string]*MockCCComm
}

//NewMockPeerSupport getsa mock peer support
func NewMockPeerSupport() *MockPeerCCSupport {
	return &MockPeerCCSupport{ccStream: make(map[string]*MockCCComm)}
}

//AddCC adds a cc to the MockPeerCCSupport
func (mp *MockPeerCCSupport) AddCC(name string, recv chan *pb.ChaincodeMessage, send chan *pb.ChaincodeMessage) (*MockCCComm, error) {
	if mp.ccStream[name] != nil {
		return nil, fmt.Errorf("CC %s already added", name)
	}
	mcc := &MockCCComm{name: name, recvStream: recv, sendStream: send}
	mp.ccStream[name] = mcc
	return mcc, nil
}

//GetCC gets a cc from the MockPeerCCSupport
func (mp *MockPeerCCSupport) GetCC(name string) (*MockCCComm, error) {
	s := mp.ccStream[name]
	if s == nil {
		return nil, fmt.Errorf("CC %s not added", name)
	}
	return s, nil
}

//GetCCMirror creates a MockCCStream with streans switched
func (mp *MockPeerCCSupport) GetCCMirror(name string) *MockCCComm {
	s := mp.ccStream[name]
	if s == nil {
		return nil
	}

	return &MockCCComm{name: name, recvStream: s.sendStream, sendStream: s.recvStream}
}

//RemoveCC removes a cc
func (mp *MockPeerCCSupport) RemoveCC(name string) error {
	if mp.ccStream[name] == nil {
		return fmt.Errorf("CC %s not added", name)
	}
	delete(mp.ccStream, name)
	return nil
}

//RemoveAll removes all ccs
func (mp *MockPeerCCSupport) RemoveAll() error {
	mp.ccStream = make(map[string]*MockCCComm)
	return nil
}
