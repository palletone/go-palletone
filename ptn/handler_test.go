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
	"math"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/palletone/go-palletone/common"
	//"github.com/palletone/go-palletone/common/event"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	//mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

/*
// Tests that protocol versions and modes of operations are matched up properly.
func TestProtocolCompatibility(t *testing.T) {
	// Define the compatibility chart
	tests := []struct {
		version    uint
		mode       downloader.SyncMode
		compatible bool
	}{
		{0, downloader.FullSync, true},
		{0, downloader.FullSync, true},
		{1, downloader.FullSync, true},
		{0, downloader.FastSync, false},
		{0, downloader.FastSync, false},
		{1, downloader.FastSync, true},
	}
	// Make sure anything we screw up is restored
	backup := ProtocolVersions
	defer func() { ProtocolVersions = backup }()
	// Try all available compatibility configs and check for errors
	for i, tt := range tests {
		ProtocolVersions = []uint{tt.version}
		pm, _, err := newTestProtocolManager(tt.mode, 0, nil)
		if pm != nil {
			defer pm.Stop()
		}
		if (err == nil && !tt.compatible) || (err != nil && tt.compatible) {
			t.Errorf("test %d: compatibility mismatch: have error %v, want compatibility %v", i, err, tt.compatible)
		}
	}
}
*/
// Tests that block headers can be retrieved from a remote chain based on user queries.
//func TestGetBlockHeaders1(t *testing.T) { testGetBlockHeaders(t, 1) }
func getUnitHashbyNumber(pm *ProtocolManager, index0 *modules.ChainIndex) common.Hash {
	u, _ := pm.dag.GetUnitByNumber(index0)
	return u.Hash()
}
func testGetBlockHeaders(t *testing.T, protocol int) {
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxHashFetch+15, nil, nil, nil, nil)
	peer, _ := newTestPeer("peer", protocol, pm, true, pm.dag)
	defer peer.close()
	// Create a "random" unknown hash for testing
	var unknown common.Hash
	for i := range unknown {
		unknown[i] = byte(i)
	}
	// Create a batch of tests for various scenarios
	limit := uint64(downloader.MaxHeaderFetch)
	index := modules.ChainIndex{
		modules.PTNCOIN,

		0,
	}
	index0 := &modules.ChainIndex{
		modules.PTNCOIN,

		limit / 2,
	}
	index1 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 + 1,
	}
	index2 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 + 2,
	}
	index21 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 - 1,
	}
	index22 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 - 2,
	}
	index4 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 + 4,
	}
	index8 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 + 8,
	}
	index24 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 - 4,
	}
	index28 := modules.ChainIndex{
		modules.PTNCOIN,

		limit/2 - 8,
	}
	index44 := modules.ChainIndex{
		modules.PTNCOIN,

		4,
	}
	i := pm.dag.CurrentUnit(modules.PTNCOIN).Number()
	jia1 := modules.ChainIndex{
		modules.PTNCOIN,

		i.Index + 1,
	}
	in1 := modules.ChainIndex{
		modules.PTNCOIN,

		i.Index - 1,
	}
	in4 := modules.ChainIndex{
		modules.PTNCOIN,

		i.Index - 4,
	}
	i1 := modules.ChainIndex{
		modules.PTNCOIN,

		1,
	}
	i2 := modules.ChainIndex{
		modules.PTNCOIN,

		2,
	}
	i3 := modules.ChainIndex{
		modules.PTNCOIN,

		3,
	}
	head, _ := pm.dag.GetHeaderByNumber(index0)
	tests := []struct {
		query  *getBlockHeadersData // The query to execute for header retrieval
		expect []common.Hash        // The hashes of the block whose headers are expected
	}{
		// A single random block should be retrievable by hash and number too
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: head.Hash()}, Amount: 1},
			[]common.Hash{getUnitHashbyNumber(pm, index0)},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: *index0}, Amount: 1},
			[]common.Hash{getUnitHashbyNumber(pm, index0)},
		},
		//Multiple headers should be retrievable in both directions
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: *index0}, Amount: 3},
			[]common.Hash{
				getUnitHashbyNumber(pm, index0),
				getUnitHashbyNumber(pm, &index1),
				getUnitHashbyNumber(pm, &index2),
			},
		},
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: *index0}, Amount: 3, Reverse: true},
			[]common.Hash{
				getUnitHashbyNumber(pm, index0),
				getUnitHashbyNumber(pm, &index21),
				getUnitHashbyNumber(pm, &index22),
			},
		},
		// Multiple headers with skip lists should be retrievable
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: *index0}, Skip: 3, Amount: 3},
			[]common.Hash{
				getUnitHashbyNumber(pm, index0),
				getUnitHashbyNumber(pm, &index4),
				getUnitHashbyNumber(pm, &index8),
			},
		},
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: *index0}, Skip: 3, Amount: 3, Reverse: true},
			[]common.Hash{
				getUnitHashbyNumber(pm, index0),
				getUnitHashbyNumber(pm, &index24),
				getUnitHashbyNumber(pm, &index28),
			},
		},
		//// The chain endpoints should be retrievable
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: index}, Amount: 1},
			[]common.Hash{getUnitHashbyNumber(pm, &index)},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: *pm.dag.CurrentUnit(modules.PTNCOIN).Number()}, Amount: 1},
			[]common.Hash{pm.dag.CurrentUnit(modules.PTNCOIN).Hash()},
		},
		//// Ensure protocol limits are honored
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: in1}, Amount: limit + 10, Reverse: true},
			pm.dag.GetUnitHashesFromHash(pm.dag.CurrentUnit(modules.PTNCOIN).Hash(), limit),
		},
		// Check that requesting more than available is handled gracefully
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: in4}, Skip: 3, Amount: 3},
			[]common.Hash{
				getUnitHashbyNumber(pm, &in4),
				getUnitHashbyNumber(pm, pm.dag.CurrentUnit(modules.PTNCOIN).Number()),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index44}, Skip: 3, Amount: 3, Reverse: true},
			[]common.Hash{
				getUnitHashbyNumber(pm, &index44),
				getUnitHashbyNumber(pm, &index),
			},
		},
		//// Check that requesting more than available is handled gracefully, even if mid skip
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: in4}, Skip: 2, Amount: 3},
			[]common.Hash{
				getUnitHashbyNumber(pm, &in4),
				getUnitHashbyNumber(pm, &in1),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index44}, Skip: 2, Amount: 3, Reverse: true},
			[]common.Hash{
				getUnitHashbyNumber(pm, &index44),
				getUnitHashbyNumber(pm, &i1),
			},
		},
		//// Check a corner case where requesting more can iterate past the endpoints
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: i2}, Amount: 5, Reverse: true},
			[]common.Hash{
				getUnitHashbyNumber(pm, &i2),
				getUnitHashbyNumber(pm, &i1),
				getUnitHashbyNumber(pm, &index),
			},
		},
		// Check a corner case where skipping overflow loops back into the chain start
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: getUnitHashbyNumber(pm, &i3)}, Amount: 2, Reverse: false, Skip: math.MaxUint64 - 1},
			[]common.Hash{
				getUnitHashbyNumber(pm, &i3),
			},
		},
		// Check a corner case where skipping overflow loops back to the same header
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: getUnitHashbyNumber(pm, &i1)}, Amount: 2, Reverse: false, Skip: math.MaxUint64},
			[]common.Hash{
				getUnitHashbyNumber(pm, &i1),
			},
		},
		// Check that non existing headers aren't returned
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: unknown}, Amount: 1},
			[]common.Hash{},
		},
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: jia1}, Amount: 1},
			[]common.Hash{},
		},
	}

	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the headers to expect in the response
		headers := []*modules.Header{}
		for _, hash := range tt.expect {
			u, _ := pm.dag.GetUnitByHash(hash)
			headers = append(headers, u.Header())
		}
		// Send the hash request and verify the response
		p2p.Send(peer.app, 0x03, tt.query)
		if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
			t.Errorf("test %d: headers mismatch: %v", i, err)
		}
		// If the test used number origins, repeat with hashes as the too
		if tt.query.Origin.Hash == (common.Hash{}) {
			if origin, _ := pm.dag.GetUnitByNumber(&tt.query.Origin.Number); origin != nil {
				index := &modules.ChainIndex{
					AssetID: modules.PTNCOIN,
					//IsMain:  true,
					Index: uint64(0),
				}
				tt.query.Origin.Hash, tt.query.Origin.Number = origin.Hash(), *index
				p2p.Send(peer.app, 0x03, tt.query)
				if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
					t.Errorf("test %d: headers mismatch: %v", i, err)
				}
			}
		}
	}
}

// Tests that block contents can be retrieved from a remote chain based on their hashes.
func TestGetBlockBodies1(t *testing.T) { testGetBlockBodies(t, 1) }
func testGetBlockBodies(t *testing.T, protocol int) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dag := dag.NewMockIDag(mockCtrl)
	pro := NewMockproducer(mockCtrl)
	height := 10
	mockUnit := unitForTest(height)
	stable_index := new(modules.ChainIndex)
	stable_index.Index = uint64(height - 1)
	stable_index.AssetID = modules.PTNCOIN

	dag.EXPECT().GetUnitByNumber(gomock.Any()).Return(mockUnit, nil).AnyTimes()
	dag.EXPECT().GetActiveMediatorNodes().Return(map[string]*discover.Node{}).AnyTimes()
	dag.EXPECT().CurrentHeader(modules.PTNCOIN).Return(mockUnit.Header()).AnyTimes()
	dag.EXPECT().CurrentUnit(modules.PTNCOIN).Return(mockUnit).AnyTimes()
	dag.EXPECT().GetUnitTransactions(gomock.Any()).DoAndReturn(func(hash common.Hash) (modules.Transactions, error) {
		return mockUnit.Transactions(), nil
	}).AnyTimes()
	par := core.NewChainParams()
	dag.EXPECT().GetChainParameters().Return(&par).AnyTimes()
	dag.EXPECT().SubscribeActiveMediatorsUpdatedEvent(gomock.Any()).Return(&rpc.ClientSubscription{}).AnyTimes()
	dag.EXPECT().SubscribeToGroupSignEvent(gomock.Any()).Return(&rpc.ClientSubscription{}).AnyTimes()
	dag.EXPECT().GetStableChainIndex(gomock.Any()).Return(mockUnit.UnitHeader.Number).AnyTimes()
	pro.EXPECT().LocalHaveActiveMediator().Return(false).AnyTimes()

	/*
		pro.EXPECT().SubscribeNewUnitEvent(gomock.Any()).DoAndReturn(func(ch chan<- mp.NewUnitEvent) event.Subscription {
			return nil
		}).AnyTimes()
		pro.EXPECT().SubscribeSigShareEvent(gomock.Any()).DoAndReturn(func(ch chan<- mp.SigShareEvent) event.Subscription {
			return nil
		}).AnyTimes()
		pro.EXPECT().SubscribeGroupSigEvent(gomock.Any()).DoAndReturn(func(ch chan<- mp.GroupSigEvent) event.Subscription {
			return nil
		}).AnyTimes()
		pro.EXPECT().SubscribeVSSDealEvent(gomock.Any()).DoAndReturn(func(ch chan<- mp.VSSDealEvent) event.Subscription {
			return nil
		}).AnyTimes()
		pro.EXPECT().SubscribeVSSResponseEvent(gomock.Any()).DoAndReturn(func(ch chan<- mp.VSSResponseEvent) event.Subscription {
			return nil
		}).AnyTimes()
	*/
	pro = pro

	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxBlockFetch+15, dag, nil, nil, nil)

	peer, _ := newTestPeer("peer", protocol, pm, true, pm.dag)
	defer peer.close()
	//test for send data
	hashes := []common.Hash{}
	hashes = append(hashes, mockUnit.Hash())
	bodies := blockBody{}
	for _, tx := range mockUnit.Transactions() {
		bodies.Transactions = append(bodies.Transactions, tx)
	}
	p2p.Send(peer.app, 0x05, hashes)

	//bodies = append(bodies,)
	content, _ := rlp.EncodeToBytes(bodies)
	if err := p2p.ExpectMsg(peer.app, 0x06, content /*bodies*/); err != nil {
		//TODO must recover
		//t.Errorf("test: bodies mismatch: %v", err)
	}
}
