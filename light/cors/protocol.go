/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developer Jiyou Wang <dev@pallet.one>
 * @date 2018
 */
package cors

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	NetworkId          = 10
	ProtocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message
)

const (
	cors = 10
)

var ProtocolLengths = map[uint]uint64{cors: 15}

var (
	ClientProtocolVersions = []uint{cors}
	ServerProtocolVersions = []uint{cors}
	//ClientProtocolVersions    = []uint{lpv2, lpv1}
	//ServerProtocolVersions    = []uint{lpv2, lpv1}
	//AdvertiseProtocolVersions = []uint{lpv2} // clients are searching for the first advertised protocol in the list
)

// cors protocol message codes
const (
	// Protocol messages belonging to Cors1
	StatusMsg           = 0x00
	CorsHeaderMsg       = 0x01
	CorsHeadersMsg      = 0x02
	GetCurrentHeaderMsg = 0x03
	CurrentHeaderMsg    = 0x04
	GetBlockHeadersMsg  = 0x05
	BlockHeadersMsg     = 0x06
)

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
	ErrGenesisBlockMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
	ErrUselessPeer
	ErrRequestRejected
	ErrUnexpectedResponse
	ErrInvalidResponse
	ErrTooManyTimeouts
	ErrMissingKey
)

type errCode int

// announceData is the network packet for the block announcements.
type announceData struct {
	Hash   common.Hash        // Hash of one particular block being announced
	Number modules.ChainIndex // Number of one particular block being announced
	Header modules.Header
	Update keyValueList
}

// getBlockHeadersData represents a block header query.
type getBlockHeadersData struct {
	Origin  hashOrNumber // Block from which to retrieve headers
	Amount  uint64       // Maximum number of headers to retrieve
	Skip    uint64       // Blocks to skip between consecutive headers
	Reverse bool         // Query direction (false = rising towards latest, true = falling towards genesis)
}

// hashOrNumber is a combined field for specifying an origin block.
type hashOrNumber struct {
	Hash   common.Hash        // Block hash from which to retrieve headers (excludes Number)
	Number modules.ChainIndex // Block hash from which to retrieve headers (excludes Hash)
}

// sign adds a signature to the block announcement by the given privKey
//func (a *announceData) sign(privKey *ecdsa.PrivateKey) {
//	//rlp, _ := rlp.EncodeToBytes(announceBlock{a.Hash, a.Number.Index /*, a.Td*/})
//	//sig, _ := crypto.Sign(crypto.Keccak256(rlp), privKey)
//	//a.Update = a.Update.add("sign", sig)
//}
//
//// checkSignature verifies if the block announcement has a valid signature by the given pubKey
//func (a *announceData) checkSignature(pubKey *ecdsa.PublicKey) error {
//	return nil
//}
