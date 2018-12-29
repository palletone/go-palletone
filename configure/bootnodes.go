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
	"pnode://2660db8e0911786da20c87a90e2c619359234736b86007adf49fb2fb78ba03eccee4ee839d852dd68b24238411ba56537140e7202c11b85a8f83be74453a28ae@124.251.111.62:30303",
	"pnode://3a29b3c23f5dc09393c3284c30cd1d29b5a55826a6eae2d01eeddbe9c5b6d226894f5da21ceddf2304daca04918c9ba26f7712f9996c0c52637216b6593f6c04@124.251.111.62:30305",
	"pnode://2344a0a41550544939751470140431e00088cff266e000763f2f298b9a4e109529e752e55834e7c68c74231d69e277a02e3d201761e0624067ce0bfc92db6939@60.205.177.166:30306",
}
