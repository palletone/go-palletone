// +build mainnet

/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2019
 *
 */
package configure

const UdpVersion = 1076

//var GenesisHash = []byte("fda8aea1b4b2920b1f4038fb10edb8fe510669ec1f574f5a13b20cad0f2294df")
var GenesisHash = []byte("")

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main PalletOne network.
var MainnetBootnodes = []string{
	"pnode://53d198682fd6e7335e52056952cbbeb77df0dc97909512266d64409dc1a4d9dfa331e6f5c2cd486bcb8a91a9abd4ae7989c5a1682a824ac67ce670e60707284e@hub1.pallet.one:30309",
	"pnode://e54c8ebf2be0786f837dfcae09692ed49b7d25c6b7ffdccaca93f8c6c825d418c70cb39c706457ca19e0b472b0583f9fa35badc610cdd03958fe759e656f4ad7@hub2.pallet.one:30309",
	"pnode://7acd09b54f40aacbbf082de5a3cd13157d67747f771d20857320ab0f8a8cc48174277a5a7c2da96914388dfaf0eaadadb9aae219274a31cd72f7eebe3f3c73fa@hub3.pallet.one:30303",
}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	"pnode://27108152fecd83368b40c3451b6eb774f829a89273ccdcd2c82887e2dc8afc6264d7a9797bb7abeec6be59b3c9b2f8ef6796f30e5f9b5d27b5880896de39a713@123.126.106.85:30309",
	"pnode://1d27062fc6656f4f7ff4360ffa6464f9e6ebcb27c3c766fba74ec8969a3b2d4c075a2e37d83dd9bb0e28cd761baffe97ee372f3141e5631fa3edd9f0c718633c@123.126.106.85:30310",
}
