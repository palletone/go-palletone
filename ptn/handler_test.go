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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/dag/modules"
	"testing"

	"fmt"
	"github.com/palletone/go-palletone/ptn/downloader"
	"math"
)

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

// Tests that block headers can be retrieved from a remote chain based on user queries.
func TestGetBlockHeaders1(t *testing.T) { testGetBlockHeaders(t, 1) }
func testGetBlockHeaders(t *testing.T, protocol int) {
	//pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxHashFetch+15, nil)
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, downloader.MaxHashFetch+15, nil)
	//uu := pm.dag.CurrentUnit()
	//log.Printf("--------testGetBlockHeaders--CurrentUnit--unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//log.Printf("--------testGetBlockHeaders--CurrentUnit--unit.UnitHash-------%#v\n", uu.UnitHash)
	//log.Printf("--------testGetBlockHeaders--CurrentUnit--unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//log.Printf("--------testGetBlockHeaders--CurrentUnit--unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	peer, _ := newTestPeer("peer", protocol, pm, true)
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
		true,
		0,
	}
	index0 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit / 2,
	}
	index1 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 + 1,
	}
	index2 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 + 2,
	}
	index4 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 + 4,
	}
	index8 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 + 8,
	}
	index44 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		4,
	}

	index21 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 - 1,
	}
	index22 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 - 2,
	}
	index24 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 - 4,
	}
	index28 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		limit/2 - 8,
	}
	i := pm.dag.CurrentUnit().Number()
	jia1 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		i.Index+1,
	}
	in1 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		i.Index-1,
	}
	in4 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		i.Index-4,
	}
	i1 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		1,
	}
	i2 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		2,
	}
	i3 := modules.ChainIndex{
		modules.PTNCOIN,
		true,
		3,
	}
	tests := []struct {
		query  *getBlockHeadersData // The query to execute for header retrieval
		expect []common.Hash        // The hashes of the block whose headers are expected
	}{
		// A single random block should be retrievable by hash and number too
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: pm.dag.GetUnitByNumber(index0).Hash()}, Amount: 1},
			[]common.Hash{pm.dag.GetUnitByNumber(index0).Hash()},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index0}, Amount: 1},
			[]common.Hash{pm.dag.GetUnitByNumber(index0).Hash()},
		},
		// Multiple headers should be retrievable in both directions
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: index0}, Amount: 3},
			[]common.Hash{
				pm.dag.GetUnitByNumber(index0).Hash(),
				pm.dag.GetUnitByNumber(index1).Hash(),
				pm.dag.GetUnitByNumber(index2).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index0}, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.dag.GetUnitByNumber(index0).Hash(),
				pm.dag.GetUnitByNumber(index21).Hash(),
				pm.dag.GetUnitByNumber(index22).Hash(),
			},
		},
		// Multiple headers with skip lists should be retrievable
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: index0}, Skip: 3, Amount: 3},
			[]common.Hash{
				pm.dag.GetUnitByNumber(index0).Hash(),
				pm.dag.GetUnitByNumber(index4).Hash(),
				pm.dag.GetUnitByNumber(index8).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index0}, Skip: 3, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.dag.GetUnitByNumber(index0).Hash(),
				pm.dag.GetUnitByNumber(index24).Hash(),
				pm.dag.GetUnitByNumber(index28).Hash(),
			},
		},
		// The chain endpoints should be retrievable
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: index}, Amount: 1},
			[]common.Hash{pm.dag.GetUnitByNumber(index).Hash()},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: pm.dag.CurrentUnit().Number()}, Amount: 1},
			[]common.Hash{pm.dag.CurrentUnit().Hash()},
		},
		// Ensure protocol limits are honored
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: in1}, Amount: limit + 10, Reverse: true},
			pm.dag.GetUnitHashesFromHash(pm.dag.CurrentUnit().Hash(), limit),
		},
		// Check that requesting more than available is handled gracefully
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: in4}, Skip: 3, Amount: 3},
			[]common.Hash{
				pm.dag.GetUnitByNumber(in4).Hash(),
				pm.dag.GetUnitByNumber(pm.dag.CurrentUnit().Number()).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index44}, Skip: 3, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.dag.GetUnitByNumber(index44).Hash(),
				pm.dag.GetUnitByNumber(index).Hash(),
			},
		},
		// Check that requesting more than available is handled gracefully, even if mid skip
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: in4}, Skip: 2, Amount: 3},
			[]common.Hash{
				pm.dag.GetUnitByNumber(in4).Hash(),
				pm.dag.GetUnitByNumber(in1).Hash(),
			},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: index44}, Skip: 2, Amount: 3, Reverse: true},
			[]common.Hash{
				pm.dag.GetUnitByNumber(index44).Hash(),
				pm.dag.GetUnitByNumber(i1).Hash(),
			},
		},
		// Check a corner case where requesting more can iterate past the endpoints
		{
			&getBlockHeadersData{Origin: hashOrNumber{Number: i2}, Amount: 5, Reverse: true},
			[]common.Hash{
				pm.dag.GetUnitByNumber(i2).Hash(),
				pm.dag.GetUnitByNumber(i1).Hash(),
				pm.dag.GetUnitByNumber(index).Hash(),
			},
		},
		// Check a corner case where skipping overflow loops back into the chain start
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: pm.dag.GetUnitByNumber(i3).Hash()}, Amount: 2, Reverse: false, Skip: math.MaxUint64 - 1},
			[]common.Hash{
				pm.dag.GetUnitByNumber(i3).Hash(),
			},
		},
		// Check a corner case where skipping overflow loops back to the same header
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: pm.dag.GetUnitByNumber(i1).Hash()}, Amount: 2, Reverse: false, Skip: math.MaxUint64},
			[]common.Hash{
				pm.dag.GetUnitByNumber(i1).Hash(),
			},
		},
		// Check that non existing headers aren't returned
		{
			&getBlockHeadersData{Origin: hashOrNumber{Hash: unknown}, Amount: 1},
			[]common.Hash{},
		}, {
			&getBlockHeadersData{Origin: hashOrNumber{Number: jia1}, Amount: 1},
			[]common.Hash{},
		},
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the headers to expect in the response
		headers := []*modules.Header{}
		for _, hash := range tt.expect {
			//headers = append(headers, pm.dag.GetBlockByHash(hash).Header())
			//hash = hash
			//headers = append(headers, pm.dag.CurrentUnit().UnitHeader)
			//fmt.Println("========================pm.dag.GetUnitByHash(hash)=",pm.dag.GetUnitByHash(hash))
			headers = append(headers, pm.dag.GetUnitByHash(hash).Header())
		}
		// Send the hash request and verify the response
		p2p.Send(peer.app, 0x03, tt.query)
		//fmt.Println(len(headers))
		if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
			t.Errorf("test %d: headers mismatch: %v", i, err)
		}
		// If the test used number origins, repeat with hashes as the too
		if tt.query.Origin.Hash == (common.Hash{}) {
			if origin := pm.dag.GetUnitByNumber(tt.query.Origin.Number); origin != nil {
				index := modules.ChainIndex{
						AssetID:modules.PTNCOIN,
						IsMain: true,
						Index:  uint64(0),
					}
				tt.query.Origin.Hash, tt.query.Origin.Number = origin.Hash(), index
				p2p.Send(peer.app, 0x03, tt.query)
				if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
					t.Errorf("test %d: headers mismatch: %v", i, err)
				}
			}
			//index := modules.ChainIndex{
			//	IsMain: true,
			//	Index:  uint64(0),
			//}
			//tt.query.Origin.Hash, tt.query.Origin.Number = common.Hash{}, index
			//p2p.Send(peer.app, 0x03, tt.query)
			//if err := p2p.ExpectMsg(peer.app, 0x04, headers); err != nil {
			//	t.Errorf("test %d: headers mismatch: %v", i, err)
			//}
		}
	}
}

// Tests that block contents can be retrieved from a remote chain based on their hashes.
//func TestGetBlockBodies1(t *testing.T) { testGetBlockBodies(t, 1) }
func testGetBlockBodies(t *testing.T, protocol int) {
	pm, _ := newTestProtocolManagerMust(t, downloader.FullSync, 11, nil)
	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()
	// Create a batch of tests for various scenarios
	//limit := downloader.MaxBlockFetch
	tests := []struct {
		random    int           // Number of blocks to fetch randomly from the chain
		explicit  []common.Hash // Explicitly requested blocks
		available []bool        // Availability of explicitly requested blocks
		expected  int           // Total number of existing blocks to expect
	}{
		{1, nil, nil, 1}, // A single random block should be retrievable
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Collect the hashes to request, and the response to expect
		hashes, seen := []common.Hash{}, make(map[int64]bool)
		bodies := []*blockBody{}
		for j := 0; j < tt.random; j++ {
			for {
				//num := rand.Int63n(int64(pm.dag.CurrentUnit().UnitHeader.Number.Index))
				if !seen[0] {
					seen[0] = true
					chain := modules.ChainIndex{
						Index: 0,
					}
					block := pm.dag.GetUnitByNumber(chain)
					fmt.Println("block===>", block)
					hashes = append(hashes, block.Hash())
					if len(bodies) < tt.expected {
						bodies = append(bodies, &blockBody{Transactions: block.Transactions()})
					}
					break
				}
			}
		}
		for j, hash := range tt.explicit {
			hashes = append(hashes, hash)
			if tt.available[j] && len(bodies) < tt.expected {
				block := pm.dag.GetUnitByHash(hash)
				bodies = append(bodies, &blockBody{Transactions: block.Transactions()})
			}
		}

		pay := modules.PaymentPayload{
			Input:  []*modules.Input{},
			Output: []*modules.Output{},
		}
		msg0 := &modules.Message{
			App:     modules.APP_PAYMENT,
			Payload: pay,
		}
		tx := &modules.Transaction{
			TxMessages: []*modules.Message{msg0},
		}
		bodies = append(bodies, &blockBody{Transactions: []*modules.Transaction{tx}})

		// Send the hash request and verify the response
		p2p.Send(peer.app, 0x00, hashes)
		if err := p2p.ExpectMsg(peer.app, 0x00, nil); err != nil {
			t.Errorf("test %d: bodies mismatch: %v", i, err)
		}
	}
}
