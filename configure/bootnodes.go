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
	"pnode://7fcd4fc797e5004b576f253a2ac94d1f3c5be51ea4ed931796174e1ff91f504a8810b0d5d914f7b831b3e422d11d800d561dc9a4b5da74cc3374061b95fe9f5a@124.251.111.62:30303",
	"pnode://28015c784ddc194764226f17bffdd3829cf8fd2df537b11c56dde2d47cf1dd44c06281dabe3d60861b80c0bb2f7a332410805fc8beaa3ed2b2cb4cf275797e2a@124.251.111.62:30305",
	"pnode://3c1d2e5d262d4672d41ec54b9e890de55f6f26967ee3c7f2a071ece9d3fcfd2ae92f00a2e67b8d67f0ce2aaadc64bab1cc410c495ae7198482124f3b206c9e00@60.205.177.166:30306",
}
