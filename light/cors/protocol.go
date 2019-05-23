package cors

import (
	"crypto/ecdsa"

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

// les protocol message codes
const (
	// Protocol messages belonging to LPV1
	StatusMsg           = 0x00
	CorsHeaderMsg       = 0x01
	CorsHeadersMsg      = 0x02
	GetCurrentHeaderMsg = 0x03
	CurrentHeaderMsg    = 0x04
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

// sign adds a signature to the block announcement by the given privKey
func (a *announceData) sign(privKey *ecdsa.PrivateKey) {
	//rlp, _ := rlp.EncodeToBytes(announceBlock{a.Hash, a.Number.Index /*, a.Td*/})
	//sig, _ := crypto.Sign(crypto.Keccak256(rlp), privKey)
	//a.Update = a.Update.add("sign", sig)
}

// checkSignature verifies if the block announcement has a valid signature by the given pubKey
func (a *announceData) checkSignature(pubKey *ecdsa.PublicKey) error {
	return nil

}
