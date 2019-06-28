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
	"pnode://d6250d040c3569e2f15f1dcd07f3dc0c56bc92736ecab2a57fe16a3d800fd8bbc2e52de8dacba59a0248c5982a4ebc2ed35bd63ad506d80705be681a51ee8d7d@123.126.106.82:30303",
	"pnode://de7babb5d2b14bb780a1a5fc0f3c17e03d70ddfa847cd704f2347772997cd0ae7cf016e0d899748e0291d301805a263d2e96a15361ce3f26d47893ad167adbf0@123.126.106.83:30305",
	"pnode://0096833853449ce46a1720638d044531a4f6de7cd3c22c90a96266384c9fa3cf550b14d552fbcb28e7635e0de4737ca5d985ff4954750eddbe1501ead5c461c5@123.126.106.84:30306",
	"pnode://79054cfedfdbf2bedef3c9ee865a1a9843a65eaf0939be496f5b9560e42fbbfc09ea71197601834a2e68d61eb88792a2e9c143ad6c59c343b7eb1c36e461e62d@123.126.106.85:30307",
	"pnode://3d7b3db523418565cc119bb4a4813753ad5a714441193e5c89ade195f4666691df66f4dbf43595efa9b343f8ad7946d057a47fb10ad5b05d7b00e30aa9068712@60.205.177.166:30308"}
