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

package configure

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main PalletOne network.
var MainnetBootnodes = []string{}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"pnode://79d2bb2ebebe8480d57ee0c37c91be6a7debc9b6e057be002cff00d69515825aa341ff30c1cc9d957e8c3c9f93e1842f9535a79684308d888df54540ca5eb694@123.126.106.82:30303",
	"pnode://b6c1f975b69bd912bda382d250c76d0ede2ec7b203262a77896127c1e4309d80175a64c053c79a1722c024f3c9bf71bb7888c0a74d9291a74664861a29ea2351@123.126.106.83:30305",
	"pnode://afc4f47f4a8092fbe2d55a17918901813a90e1e0546e1cc6ddfb412cf65f390573cf309c787b21680ee700bd2394d7e4a7b4577db3105ee9e8cef150020cf196@123.126.106.84:30306",
	"pnode://ddaac2cb0a6a0720dd5994397c04ed96956e5a36ce4d0075efa0947fa9b663c01eed579dbd99d8a3275d421a93f6abfacb21ca9bbc6cff5d4a0da4dc68cd11b3@123.126.106.85:30307",
	"pnode://00a21897d0ef52eddc7cbdc600e1cde6ca05e38e1c0b29f10909e7e0a4177d0af93210205237f4bd9c2b37d2e02bcd5e6c785adeedab4d844344b97e8b90393d@60.205.177.166:30308",
}
