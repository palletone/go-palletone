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
	"pnode://eb6523c0220620a63c1b2d1375d8c04a35e88d8ef05344331bad187c1fdc37a0c65eeff91963cd576b160753f58680109cc181514ebfb0831d42ca39cac2ee15@60.205.177.166:30303",
	"pnode://aaa3a9f08a2ac9e566b5ab1c9e34655c5214b55a205c7ae0526ce59d5e841c6cafdbd23a98a8ee2181b113a0855ae007a02a559365fb0f21e0a72c9237940794@123.126.106.83:30305",
	"pnode://d18ea6e8cb09984b19fbf9c47636aaa952f8aef3fabdef75756d946990091510fd3b4fbff5db5860e3b395e5bc6159454a3f5b9b87a27f6efa8a081a037a4d6c@123.126.106.84:30306",
	"pnode://28f45690b29282b3ed19373f20099c6e00e275183c795d2082820cafd54d91e4009610effc7249feeaaf826ad21c3a5fa983567d190f9d3b4305338cc5a847b6@123.126.106.85:30307",
	"pnode://ef2b94a2305bbbe7a9445647cf73b6aeae96e5b75f16eb78f2605f2b85f6099842cec6a01e7f995caf813b61072b99090d37141ec31fc338054bb6500925ace5@123.126.106.85:30308",
}
