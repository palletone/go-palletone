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

//const StableIndex = 1672212

//var GenesisHash = []byte("fda8aea1b4b2920b1f4038fb10edb8fe510669ec1f574f5a13b20cad0f2294df")
var GenesisHash = []byte("")

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main PalletOne network.
var MainnetBootnodes = []string{
	"pnode://53d198682fd6e7335e52056952cbbeb77df0dc97909512266d64409dc1a4d9dfa331e6f5c2cd486bcb8a91a9abd4ae7989c5a1682a824ac67ce670e60707284e@hub1.pallet.one:30309",
	"pnode://fc541a39fe5443adaa797c42ecf05a0b0a1d29af9f4843531adee999dffdc19509a810ebc7c85502d90030f0497bdb3b29fb254d1259c9b57be25a7edb30b8aa@hub2.pallet.one:30303",
	"pnode://7acd09b54f40aacbbf082de5a3cd13157d67747f771d20857320ab0f8a8cc48174277a5a7c2da96914388dfaf0eaadadb9aae219274a31cd72f7eebe3f3c73fa@hub3.pallet.one:30303",
}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	"pnode://27108152fecd83368b40c3451b6eb774f829a89273ccdcd2c82887e2dc8afc6264d7a9797bb7abeec6be59b3c9b2f8ef6796f30e5f9b5d27b5880896de39a713@123.126.106.85:30309",
	"pnode://1d27062fc6656f4f7ff4360ffa6464f9e6ebcb27c3c766fba74ec8969a3b2d4c075a2e37d83dd9bb0e28cd761baffe97ee372f3141e5631fa3edd9f0c718633c@123.126.106.85:30310",
	"pnode://d6250d040c3569e2f15f1dcd07f3dc0c56bc92736ecab2a57fe16a3d800fd8bbc2e52de8dacba59a0248c5982a4ebc2ed35bd63ad506d80705be681a51ee8d7d@123.126.106.82:30303",
	"pnode://3d7b3db523418565cc119bb4a4813753ad5a714441193e5c89ade195f4666691df66f4dbf43595efa9b343f8ad7946d057a47fb10ad5b05d7b00e30aa9068712@60.205.177.166:30308",
}
