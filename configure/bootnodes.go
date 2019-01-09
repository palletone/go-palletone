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
	"pnode://47df61a19b9657200ccf95967f08493c1b401d6c6b0238707b1ee435b621d26875a393abdb687f001303df26c606e6796857fd45b8a1ca42139a929ec1e0abd6@124.251.111.62:30303",
	"pnode://15182ec892612759d46195af0a09129ebdc19a488f0e2d9fb46de6687bde4a8cbb59e7679184ed00aa47e5b6bede7844d0a0527ed29b7f95e03a2d30e9c8427a@124.251.111.62:30305",
	"pnode://258b88772357bb102c5c836fdd3dc1baf7203eac237773c44b040130855cb6ffa8dadfb2b6a0d364229110f13f54650af3d20db463561946a7ca4a9486b4d1c8@60.205.177.166:30306",
}
