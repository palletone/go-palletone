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
	"pnode://556edb4145cbf9265c770d5e5a81a6747a421cdb7a692aaa85699c5ba98d4c1af9597bc639fdb83f610cd0b69ffd8ea9d15f7eb9c96e8b28525feaa07b01ea4e@123.126.106.82:30303",
	"pnode://99233d4edbdcc9aa827de23f6ac504e423bbbc20df524390032cd4ffcaae54166646cc7259c9dfa5a3079ac719bffb90b0c4629cf022484cd4c9ce933485af39@123.126.106.83:30305",
	"pnode://0028dd2530b8b60a824ed2ebff8189752d314e726edf68aad6b6a75970f6b0faa492f04517324e3b6833ed3a7c61c874268d16294a87679940888236b8d7197d@123.126.106.84:30306",
	"pnode://4bd20fda4a5d247043e68299a24fe5b285dc8f7f791bdf86402783dafe43ad1a0e4035001a318567480a71b62d1362c5821a85acfe82bb6baf91f71735beab15@123.126.106.85:30307",
	"pnode://cc19faee491b7624d69a8f191144fff5f98fcd0104629cfd9fda8c2d83ebc259773ce1204b365e7fb5a89f104711f7efcd0a5e0959346004af33694ec1216fe6@60.205.177.166:30308",
	"pnode://28d6a39df4303dfbb949f478072950c016c11d483a59f37433c801c5be19dd6f7d474b8c1fdf2ad71524c89656d8f869fb3db93c013a986bcf8406ca0e3b2eb7@47.74.209.46:30309",
	"pnode://0d8cfe1f9cf81f6ba3a4675e67e3931481c3a2c33adeea3f1fae2eae15465c1b0a4bb42380a4ef08a10d670246aff5bb5c359db7ac62270f7160b21e3c7f01e8@39.105.121.252:30310",
}
