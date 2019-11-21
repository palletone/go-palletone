// +build !mainnet

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

const UdpVersion = 1070

var GenesisHash = []byte("")

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main PalletOne network.
var MainnetBootnodes = []string{
	"pnode://27108152fecd83368b40c3451b6eb774f829a89273ccdcd2c82887e2dc8afc6264d7a9797bb7abeec6be59b3c9b2f8ef6796f30e5f9b5d27b5880896de39a713@123.126.106.85:30309",
	"pnode://1d27062fc6656f4f7ff4360ffa6464f9e6ebcb27c3c766fba74ec8969a3b2d4c075a2e37d83dd9bb0e28cd761baffe97ee372f3141e5631fa3edd9f0c718633c@123.126.106.85:30310",
	//	"pnode://0b148c639802c7dadcbe9773870d301b078d28b6ae34b4a6e48f9d4dfdfe3d184e082fae4f70af23c403b3f1878678bdede9c45898f40b0b8d77d4e4856aa638@123.126.106.88:30309",
	//	"pnode://495747920c44a2873b76a0c0cd40579975b3dc9438925ce8b0effcefea1593af17449e2532533399a28e82bf1a51eb3ebe480b21a447b740dfb1d4b2b29386b9@123.126.106.89:30309",
}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"pnode://27108152fecd83368b40c3451b6eb774f829a89273ccdcd2c82887e2dc8afc6264d7a9797bb7abeec6be59b3c9b2f8ef6796f30e5f9b5d27b5880896de39a713@123.126.106.85:30309",
	"pnode://1d27062fc6656f4f7ff4360ffa6464f9e6ebcb27c3c766fba74ec8969a3b2d4c075a2e37d83dd9bb0e28cd761baffe97ee372f3141e5631fa3edd9f0c718633c@123.126.106.85:30310",
}
