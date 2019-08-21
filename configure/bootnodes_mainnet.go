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

var GenesisHash = []byte("fda8aea1b4b2920b1f4038fb10edb8fe510669ec1f574f5a13b20cad0f2294df")

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main PalletOne network.
var MainnetBootnodes = []string{
	"pnode://53d198682fd6e7335e52056952cbbeb77df0dc97909512266d64409dc1a4d9dfa331e6f5c2cd486bcb8a91a9abd4ae7989c5a1682a824ac67ce670e60707284e@hub1.pallet.one:30309",
	"pnode://e54c8ebf2be0786f837dfcae09692ed49b7d25c6b7ffdccaca93f8c6c825d418c70cb39c706457ca19e0b472b0583f9fa35badc610cdd03958fe759e656f4ad7@hub2.pallet.one:30309",
	"pnode://7acd09b54f40aacbbf082de5a3cd13157d67747f771d20857320ab0f8a8cc48174277a5a7c2da96914388dfaf0eaadadb9aae219274a31cd72f7eebe3f3c73fa@hub3.pallet.one:30303",
	//"pnode://98627f6b5fa0f549ff644e66fc2b9801ea039614784ea65d1e9e942fed573004a65d0b2e85d4a949c28844efe6f357d97c1198e1194eb436fb0a8bbfafb8e6d2@123.126.106.88:30303",
	//"pnode://eaef728cb3bb7d96f5efb377b2a21d134061eb4ae0276a9573055f998b7fde4fece4aac1d93c5d9bb57d8ddf054f922720dd6188f30c8f85cc68ee9e37465958@123.126.106.88:30305",
	//"pnode://d9d50a943c836e1576589e948c5f0b92429173e617906bcecbe16b528f3bd924f831d86b1abdb646c9ecfed466760f18423f35ef26a5ef73ac102418da6dc3c9@123.126.106.88:30306",
	//"pnode://beacea1636d1ee955247199d09cdc39db38db115616e22dd6c50939b44d40a876cbda4f7cedf86a2398ff67f8d67c63fbc88be1e86277377c75250ae023aa15b@123.126.106.89:30307",
	//"pnode://5e93e974f036fa917d66277dafb020c2c705321c82fefde81b186a934a42867324f966a83f51884a0b1e8694bd241405fbdac3131f644de3d1d4b8f88a52886a@123.126.106.89:30308",
}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	"pnode://27108152fecd83368b40c3451b6eb774f829a89273ccdcd2c82887e2dc8afc6264d7a9797bb7abeec6be59b3c9b2f8ef6796f30e5f9b5d27b5880896de39a713@123.126.106.85:30309",
	"pnode://1d27062fc6656f4f7ff4360ffa6464f9e6ebcb27c3c766fba74ec8969a3b2d4c075a2e37d83dd9bb0e28cd761baffe97ee372f3141e5631fa3edd9f0c718633c@123.126.106.85:30310",
}

//var TestnetBootnodes = []string{
//	"pnode://d6250d040c3569e2f15f1dcd07f3dc0c56bc92736ecab2a57fe16a3d800fd8bbc2e52de8dacba59a0248c5982a4ebc2ed35bd63ad506d80705be681a51ee8d7d@123.126.106.82:30303",
//	"pnode://de7babb5d2b14bb780a1a5fc0f3c17e03d70ddfa847cd704f2347772997cd0ae7cf016e0d899748e0291d301805a263d2e96a15361ce3f26d47893ad167adbf0@123.126.106.83:30305",
//	"pnode://0096833853449ce46a1720638d044531a4f6de7cd3c22c90a96266384c9fa3cf550b14d552fbcb28e7635e0de4737ca5d985ff4954750eddbe1501ead5c461c5@123.126.106.84:30306",
//	"pnode://79054cfedfdbf2bedef3c9ee865a1a9843a65eaf0939be496f5b9560e42fbbfc09ea71197601834a2e68d61eb88792a2e9c143ad6c59c343b7eb1c36e461e62d@123.126.106.85:30307",
//	"pnode://3d7b3db523418565cc119bb4a4813753ad5a714441193e5c89ade195f4666691df66f4dbf43595efa9b343f8ad7946d057a47fb10ad5b05d7b00e30aa9068712@60.205.177.166:30308",
//}
