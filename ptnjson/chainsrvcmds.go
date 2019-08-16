package ptnjson

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"hash"

	//"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"github.com/palletone/go-palletone/common"
	"golang.org/x/crypto/ripemd160"
)

// Standard JSON-RPC 2.0 errors.
var (
	ErrRPCInvalidRequest = &RPCError{
		Code:    -32600,
		Message: "Invalid request",
	}
	ErrRPCMethodNotFound = &RPCError{
		Code:    -32601,
		Message: "Method not found",
	}
	ErrRPCInvalidParams = &RPCError{
		Code:    -32602,
		Message: "Invalid parameters",
	}
	ErrRPCInternal = &RPCError{
		Code:    -32603,
		Message: "Internal error",
	}
	ErrRPCParse = &RPCError{
		Code:    -32700,
		Message: "Parse error",
	}
)

// General application defined JSON errors.
const (
	ErrRPCMisc                RPCErrorCode = -1
	ErrRPCForbiddenBySafeMode RPCErrorCode = -2
	ErrRPCType                RPCErrorCode = -3
	ErrRPCInvalidAddressOrKey RPCErrorCode = -5
	ErrRPCOutOfMemory         RPCErrorCode = -7
	ErrRPCInvalidParameter    RPCErrorCode = -8
	ErrRPCDatabase            RPCErrorCode = -20
	ErrRPCDeserialization     RPCErrorCode = -22
	ErrRPCVerify              RPCErrorCode = -25
)

// Peer-to-peer client errors.
const (
	ErrRPCClientNotConnected      RPCErrorCode = -9
	ErrRPCClientInInitialDownload RPCErrorCode = -10
	ErrRPCClientNodeNotAdded      RPCErrorCode = -24
)

const (
	// pubkeyCompressed     byte = 0x2 // y_bit + x coord
	// pubkeyUncompressed   byte = 0x4 // x coord + y coord
	// pubkeyHybrid         byte = 0x6 // y_bit + x coord + y coord
	MinCoinbaseScriptLen      = 2
	MaxCoinbaseScriptLen      = 100
)

// Wallet JSON errors
const (
	ErrRPCWallet                    RPCErrorCode = -4
	ErrRPCWalletInsufficientFunds   RPCErrorCode = -6
	ErrRPCWalletInvalidAccountName  RPCErrorCode = -11
	ErrRPCWalletKeypoolRanOut       RPCErrorCode = -12
	ErrRPCWalletUnlockNeeded        RPCErrorCode = -13
	ErrRPCWalletPassphraseIncorrect RPCErrorCode = -14
	ErrRPCWalletWrongEncState       RPCErrorCode = -15
	ErrRPCWalletEncryptionFailed    RPCErrorCode = -16
	ErrRPCWalletAlreadyUnlocked     RPCErrorCode = -17
)
const (
	ErrRPCBlockNotFound     RPCErrorCode = -5
	ErrRPCBlockCount        RPCErrorCode = -5
	ErrRPCBestBlockHash     RPCErrorCode = -5
	ErrRPCDifficulty        RPCErrorCode = -5
	ErrRPCOutOfRange        RPCErrorCode = -1
	ErrRPCNoTxInfo          RPCErrorCode = -5
	ErrRPCNoCFIndex         RPCErrorCode = -5
	ErrRPCNoNewestBlockInfo RPCErrorCode = -5
	ErrRPCInvalidTxVout     RPCErrorCode = -5
	ErrRPCRawTxString       RPCErrorCode = -32602
	ErrRPCDecodeHexString   RPCErrorCode = -22
	ErrBadTxOutValue        RPCErrorCode = -23
	ErrDuplicateTxInputs    RPCErrorCode = -24
	ErrBadCoinbaseScriptLen RPCErrorCode = -25
	ErrBadTxInput           RPCErrorCode = -26
)
const (
	// MaxBlockWeight defines the maximum block weight, where "block
	// weight" is interpreted as defined in BIP0141. A block's weight is
	// calculated as the sum of the of bytes in the existing transactions
	// and header, plus the weight of each byte within a transaction. The
	// weight of a "base" byte is 4, while the weight of a witness byte is
	// 1. As a result, for a block to be valid, the BlockWeight MUST be
	// less than, or equal to MaxBlockWeight.
	MaxBlockWeight = 4000000

	// MaxBlockBaseSize is the maximum number of bytes within a block
	// which can be allocated to non-witness data.
	MaxBlockBaseSize = 1000000

	// MaxBlockSigOpsCost is the maximum number of signature operations
	// allowed for a block. It is calculated via a weighted algorithm which
	// weights segregated witness sig ops lower than regular sig ops.
	MaxBlockSigOpsCost = 80000
)
const (
	// SatoshiPerBitcent is the number of satoshi in one bitcoin cent.
	DaoPerPtncent = 1e6

	// SatoshiPerBitcoin is the number of satoshi in one bitcoin (1 BTC).
	DaoPerPtn = 1e8

	// MaxSatoshi is the maximum transaction amount allowed in satoshi.
	MaxDao = 10e8 * DaoPerPtn
)

// These constants define the lengths of serialized public keys.
const (
	PubKeyBytesLenCompressed   = 33
	PubKeyBytesLenUncompressed = 65
	PubKeyBytesLenHybrid       = 65
)

type ErrorCode int

// These constants are used to identify a specific RuleError.
const (
	// ErrDuplicateBlock indicates a block with the same hash already
	// exists.
	ErrDuplicateBlock ErrorCode = iota

	// ErrBlockTooBig indicates the serialized block size exceeds the
	// maximum allowed size.
	ErrBlockTooBig

	// ErrBlockWeightTooHigh indicates that the block's computed weight
	// metric exceeds the maximum allowed value.
	ErrBlockWeightTooHigh

	// ErrBlockVersionTooOld indicates the block version is too old and is
	// no longer accepted since the majority of the network has upgraded
	// to a newer version.
	ErrBlockVersionTooOld

	// ErrInvalidTime indicates the time in the passed block has a precision
	// that is more than one second.  The chain consensus rules require
	// timestamps to have a maximum precision of one second.
	ErrInvalidTime

	// ErrTimeTooOld indicates the time is either before the median time of
	// the last several blocks per the chain consensus rules or prior to the
	// most recent checkpoint.
	ErrTimeTooOld

	// ErrTimeTooNew indicates the time is too far in the future as compared
	// the current time.
	ErrTimeTooNew

	// ErrDifficultyTooLow indicates the difficulty for the block is lower
	// than the difficulty required by the most recent checkpoint.
	ErrDifficultyTooLow

	// ErrUnexpectedDifficulty indicates specified bits do not align with
	// the expected value either because it doesn't match the calculated
	// valued based on difficulty regarted rules or it is out of the valid
	// range.
	ErrUnexpectedDifficulty

	// ErrHighHash indicates the block does not hash to a value which is
	// lower than the required target difficultly.
	ErrHighHash

	// ErrBadMerkleRoot indicates the calculated merkle root does not match
	// the expected value.
	ErrBadMerkleRoot

	// ErrBadCheckpoint indicates a block that is expected to be at a
	// checkpoint height does not match the expected one.
	ErrBadCheckpoint

	// ErrForkTooOld indicates a block is attempting to fork the block chain
	// before the most recent checkpoint.
	ErrForkTooOld

	// ErrCheckpointTimeTooOld indicates a block has a timestamp before the
	// most recent checkpoint.
	ErrCheckpointTimeTooOld

	// ErrNoTransactions indicates the block does not have a least one
	// transaction.  A valid block must have at least the coinbase
	// transaction.
	ErrNoTransactions

	// ErrNoTxInputs indicates a transaction does not have any inputs.  A
	// valid transaction must have at least one input.
	ErrNoTxInputs

	// ErrNoTxOutputs indicates a transaction does not have any outputs.  A
	// valid transaction must have at least one output.
	ErrNoTxOutputs

	// ErrTxTooBig indicates a transaction exceeds the maximum allowed size
	// when serialized.
	ErrTxTooBig

	// ErrDuplicateTxInputs indicates a transaction references the same
	// input more than once.

	// ErrBadTxInput indicates a transaction input is invalid in some way
	// such as referencing a previous transaction outpoint which is out of
	// range or not referencing one at all.
	// ErrMissingTxOut indicates a transaction output referenced by an input
	// either does not exist or has already been spent.
	ErrMissingTxOut

	// ErrUnfinalizedTx indicates a transaction has not been finalized.
	// A valid block may only contain finalized transactions.
	ErrUnfinalizedTx

	// ErrDuplicateTx indicates a block contains an identical transaction
	// (or at least two transactions which hash to the same value).  A
	// valid block may only contain unique transactions.
	ErrDuplicateTx

	// ErrOverwriteTx indicates a block contains a transaction that has
	// the same hash as a previous transaction which has not been fully
	// spent.
	ErrOverwriteTx

	// ErrImmatureSpend indicates a transaction is attempting to spend a
	// coinbase that has not yet reached the required maturity.
	ErrImmatureSpend

	// ErrSpendTooHigh indicates a transaction is attempting to spend more
	// value than the sum of all of its inputs.
	ErrSpendTooHigh

	// ErrBadFees indicates the total fees for a block are invalid due to
	// exceeding the maximum possible value.
	ErrBadFees

	// ErrTooManySigOps indicates the total number of signature operations
	// for a transaction or block exceed the maximum allowed limits.
	ErrTooManySigOps

	// ErrFirstTxNotCoinbase indicates the first transaction in a block
	// is not a coinbase transaction.
	ErrFirstTxNotCoinbase

	// ErrMultipleCoinbases indicates a block contains more than one
	// coinbase transaction.
	ErrMultipleCoinbases

	// ErrBadCoinbaseScriptLen indicates the length of the signature script
	// for a coinbase transaction is not within the valid range.

	// ErrBadCoinbaseValue indicates the amount of a coinbase value does
	// not match the expected value of the subsidy plus the sum of all fees.
	ErrBadCoinbaseValue

	// ErrMissingCoinbaseHeight indicates the coinbase transaction for a
	// block does not start with the serialized block block height as
	// required for version 2 and higher blocks.
	ErrMissingCoinbaseHeight

	// ErrBadCoinbaseHeight indicates the serialized block height in the
	// coinbase transaction for version 2 and higher blocks does not match
	// the expected value.
	ErrBadCoinbaseHeight

	// ErrScriptMalformed indicates a transaction script is malformed in
	// some way.  For example, it might be longer than the maximum allowed
	// length or fail to parse.
	ErrScriptMalformed

	// ErrScriptValidation indicates the result of executing transaction
	// script failed.  The error covers any failure when executing scripts
	// such signature verification failures and execution past the end of
	// the stack.
	ErrScriptValidation

	// ErrUnexpectedWitness indicates that a block includes transactions
	// with witness data, but doesn't also have a witness commitment within
	// the coinbase transaction.
	ErrUnexpectedWitness

	// ErrInvalidWitnessCommitment indicates that a block's witness
	// commitment is not well formed.
	ErrInvalidWitnessCommitment

	// ErrWitnessCommitmentMismatch indicates that the witness commitment
	// included in the block's coinbase transaction doesn't match the
	// manually computed witness commitment.
	ErrWitnessCommitmentMismatch

	// ErrPreviousBlockUnknown indicates that the previous block is not known.
	ErrPreviousBlockUnknown

	// ErrInvalidAncestorBlock indicates that an ancestor of this block has
	// already failed validation.
	ErrInvalidAncestorBlock

	// ErrPrevBlockNotBest indicates that the block's previous block is not the
	// current chain tip. This is not a block validation rule, but is required
	// for block proposals submitted via getblocktemplate RPC.
	ErrPrevBlockNotBest
)

type RawTxInput struct {
	Txid         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	MessageIndex uint32 `json:"messageindex"`
	ScriptPubKey string `json:"scriptPubKey"`
	RedeemScript string `json:"redeemScript"`
}

// SignRawTransactionCmd defines the signrawtransaction JSON-RPC command.
type SignRawTransactionCmd struct {
	RawTx    string
	Inputs   *[]RawTxInput
	PrivKeys *[]string
	Flags    *string `jsonrpcdefault:"\"ALL\""`
}

func NewSignRawTransactionCmd(hexEncodedTx string, inputs *[]RawTxInput, privKeys *[]string, flags *string) *SignRawTransactionCmd {
	return &SignRawTransactionCmd{
		RawTx:    hexEncodedTx,
		Inputs:   inputs,
		PrivKeys: privKeys,
		Flags:    flags,
	}
}

type SignRawTransactionResult struct {
	Hex      string                    `json:"hex"`
	Txid     string                    `json:"txid"`
	Complete bool                      `json:"complete"`
	Errors   []SignRawTransactionError `json:"errors,omitempty"`
}

type SignRawTransactionError struct {
	TxID      string `json:"txid"`
	ScriptSig string `json:"scriptSig"`
	Error     string `json:"error"`
	Vout      uint32 `json:"vout"`
	Sequence  uint32 `json:"sequence"`
}

// RPCErrorCode represents an error code to be used as a part of an RPCError
// which is in turn used in a JSON-RPC Response object.
//
// A specific type is used to help ensure the wrong errors aren't used.
type RPCErrorCode int

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

// Guarantee RPCError satisifies the builtin error interface.
var _, _ error = RPCError{}, (*RPCError)(nil)

// Error returns a string describing the RPC error.  This satisifies the
// builtin error interface.
func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// NewRPCError constructs and returns a new JSON-RPC error that is suitable
// for use in a JSON-RPC Response object.
func NewRPCError(code RPCErrorCode, message string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
	}
}

// AddressPubKeyHash is an Address for a pay-to-pubkey-hash (P2PKH)
// transaction.
type AddressPubKeyHash struct {
	hash  [ripemd160.Size]byte
	netID byte
}

// NewAddressPubKeyHash returns a new AddressPubKeyHash.  pkHash mustbe 20
// bytes.
func NewAddressPubKeyHash(pkHash []byte, netID byte) (*AddressPubKeyHash, error) {
	return newAddressPubKeyHash(pkHash, netID)
}
func newAddressPubKeyHash(pkHash []byte, netID byte) (*AddressPubKeyHash, error) {
	// Check for a valid pubkey hash length.
	if len(pkHash) != ripemd160.Size {
		return nil, errors.New("pkHash must be 20 bytes")
	}

	addr := &AddressPubKeyHash{netID: netID}
	copy(addr.hash[:], pkHash)
	return addr, nil
}
func encodeAddress(hash160 []byte, netID byte) string {
	// Format is 1 byte for a network and address class (i.e. P2PKH vs
	// P2SH), 20 bytes for a RIPEMD160 hash, and 4 bytes of checksum.
	return base58.CheckEncode(hash160[:ripemd160.Size], netID)
}

// EncodeAddress returns the string encoding of a pay-to-pubkey-hash
// address.  Part of the Address interface.
func (a *AddressPubKeyHash) EncodeAddress() string {
	return encodeAddress(a.hash[:], a.netID)
}

// ScriptAddress returns the bytes to be included in a txout script to pay
// to a pubkey hash.  Part of the Address interface.
func (a *AddressPubKeyHash) ScriptAddress() []byte {
	return a.hash[:]
}

// AddressScriptHash is an Address for a pay-to-script-hash (P2SH)
// transaction.
type AddressScriptHash struct {
	hash  [ripemd160.Size]byte
	netID byte
}

// Amount represents the base bitcoin monetary unit (colloquially referred
// to as a `Satoshi').  A single Amount is equal to 1e-8 of a bitcoin.
type Amount int64

// round converts a floating point number, which may or may not be representable
// as an integer, to the Amount integer type by rounding to the nearest integer.
// This is performed by adding or subtracting 0.5 depending on the sign, and
// relying on integer truncation to round the value to the nearest Amount.
// func round(f float64) Amount {
// 	if f < 0 {
// 		return Amount(f - 0.5)
// 	}
// 	return Amount(f + 0.5)
// }

//
//func NewAmount(f decimal.Decimal) (Amount, error) {
//	// The amount is only considered invalid if it cannot be represented
//	// as an integer type.  This may happen if f is NaN or +-Infinity.
//	switch {
//	case math.IsNaN(f):
//		fallthrough
//	case math.IsInf(f, 1):
//		fallthrough
//	case math.IsInf(f, -1):
//		return 0, errors.New("invalid bitcoin amount")
//	}
//
//	return round(f * DaoPerPtn), nil
//}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}

// NewAddressScriptHash returns a new AddressScriptHash.
func NewAddressScriptHash(serializedScript []byte, netScriptHashAddrID byte) (*AddressScriptHash, error) {
	scriptHash := Hash160(serializedScript)
	return newAddressScriptHashFromHash(scriptHash, netScriptHashAddrID)
}

// newAddressScriptHashFromHash is the internal API to create a script hash
// address with a known leading identifier byte for a network, rather than
// looking it up through its parameters.  This is useful when creating a new
// address structure from a string encoding where the identifier byte is already
// known.
func newAddressScriptHashFromHash(scriptHash []byte, netID byte) (*AddressScriptHash, error) {
	// Check for a valid script hash length.
	if len(scriptHash) != ripemd160.Size {
		return nil, errors.New("scriptHash must be 20 bytes")
	}

	addr := &AddressScriptHash{netID: netID}
	copy(addr.hash[:], scriptHash)
	return addr, nil
}

// String returns a human-readable string for the pay-to-script-hash address.
// This is equivalent to calling EncodeAddress, but is provided so the type can
// be used as a fmt.Stringer.
func (a *AddressScriptHash) String() string {
	return a.EncodeAddress()
}

// EncodeAddress returns the string encoding of a pay-to-script-hash
// address.  Part of the Address interface.
func (a *AddressScriptHash) EncodeAddress() string {
	return encodeAddress(a.hash[:], a.netID)
}

// PubKeyFormat describes what format to use for a pay-to-pubkey address.
type PubKeyFormat int

// AddressPubKey is an Address for a pay-to-pubkey transaction.
type AddressPubKey struct {
	pubKeyFormat PubKeyFormat
	pubKey       *btcec.PublicKey
	pubKeyHashID byte
}

const (
	// PKFUncompressed indicates the pay-to-pubkey address format is an
	// uncompressed public key.
	PKFUncompressed PubKeyFormat = iota

	// PKFCompressed indicates the pay-to-pubkey address format is a
	// compressed public key.
	PKFCompressed

	// PKFHybrid indicates the pay-to-pubkey address format is a hybrid
	// public key.
	PKFHybrid
)

var (
	// ErrChecksumMismatch describes an error where decoding failed due
	// to a bad checksum.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrUnknownAddressType describes an error where an address can not
	// decoded as a specific address type due to the string encoding
	// beginning with an identifier byte unknown to any standard or
	// registered (via chaincfg.Register) network.
	ErrUnknownAddressType = errors.New("unknown address type")

	// ErrAddressCollision describes an error where an address can not
	// be uniquely determined as either a pay-to-pubkey-hash or
	// pay-to-script-hash address since the leading identifier is used for
	// describing both address kinds, but for different networks.  Rather
	// than assuming or defaulting to one or the other, this error is
	// returned and the caller must decide how to decode the address.
	ErrAddressCollision = errors.New("address collision")
)

// EncodeAddress returns the string encoding of a pay-to-pubkey-hash
// address.  Part of the Address interface.

func (a *AddressPubKey) EncodeAddress() string {
	return encodeAddress(Hash160(a.serialize()), a.pubKeyHashID)
}
func NewAddressPubKey(serializedPubKey []byte) (*AddressPubKey, error) {
	pubKey, err := btcec.ParsePubKey(serializedPubKey, btcec.S256())
	if err != nil {
		return nil, err
	}

	// Set the format of the pubkey.  This probably should be returned
	// from btcec, but do it here to avoid API churn.  We already know the
	// pubkey is valid since it parsed above, so it's safe to simply examine
	// the leading byte to get the format.
	pkFormat := PKFUncompressed
	switch serializedPubKey[0] {
	case 0x02, 0x03:
		pkFormat = PKFCompressed
	case 0x06, 0x07:
		pkFormat = PKFHybrid
	}

	return &AddressPubKey{
		pubKeyFormat: pkFormat,
		pubKey:       pubKey,
		pubKeyHashID: 0x00,
		//pubKeyHashID: netID,
	}, nil
}

//serialize returns the serialization of the public key according to the
// format associated with the address.
func (a *AddressPubKey) serialize() []byte {
	switch a.pubKeyFormat {
	default:
		fallthrough
	case PKFUncompressed:
		return a.pubKey.SerializeUncompressed()

	case PKFCompressed:
		return a.pubKey.SerializeCompressed()

	case PKFHybrid:
		return a.pubKey.SerializeHybrid()
	}
}

const compressMagic byte = 0x01

// PrivKeyBytesLen defines the length in bytes of a serialized private key.
const PrivKeyBytesLen = 32

var ErrMalformedPrivateKey = errors.New("malformed private key")

/*func DecodeWIF(wif string) (*WIF, error) {
	decoded := base58.Decode(wif)
	decodedLen := len(decoded)
	var compress bool

	// Length of base58 decoded WIF must be 32 bytes + an optional 1 byte
	// (0x01) if compressed, plus 1 byte for netID + 4 bytes of checksum.
	switch decodedLen {
	case 1 + PrivKeyBytesLen + 1 + 4:
		if decoded[33] != compressMagic {
			return nil, ErrMalformedPrivateKey
		}
		compress = true
	case 1 + PrivKeyBytesLen + 4:
		compress = false
	default:
		return nil, ErrMalformedPrivateKey
	}

	// Checksum is first four bytes of double SHA256 of the identifier byte
	// and privKey.  Verify this matches the final 4 bytes of the decoded
	// private key.
	var tosum []byte
	if compress {
		tosum = decoded[:1+PrivKeyBytesLen+1]
	} else {
		tosum = decoded[:1+PrivKeyBytesLen]
	}
	cksum := common.DoubleHashB(tosum)[:4]
	if !bytes.Equal(cksum, decoded[decodedLen-4:]) {
		return nil, ErrChecksumMismatch
	}

	netID := decoded[0]
	privKeyBytes := decoded[1 : 1+PrivKeyBytesLen]
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	return &WIF{privKey, compress, netID}, nil
}*/

type WIF struct {
	// PrivKey is the private key being imported or exported.
	PrivKey *ecdsa.PrivateKey

	// CompressPubKey specifies whether the address controlled by the
	// imported or exported private key was created by hashing a
	// compressed (33-byte) serialized public key, rather than an
	// uncompressed (65-byte) one.
	CompressPubKey bool

	// netID is the bitcoin network identifier byte used when
	// WIF encoding the private key.
	netID byte
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

// SerializePubKey serializes the associated public key of the imported or
// exported private key in either a compressed or uncompressed format.  The
// serialization format chosen depends on the value of w.CompressPubKey.
func (w *WIF) SerializePubKey() []byte {
	pk := (*btcec.PublicKey)(&w.PrivKey.PublicKey)
	if w.CompressPubKey {
		return pk.SerializeCompressed()
	}
	return pk.SerializeUncompressed()
}

// String creates the Wallet Import Format string encoding of a WIF structure.
// See DecodeWIF for a detailed breakdown of the format and requirements of
// a valid WIF string.
func (w *WIF) String() string {
	// Precalculate size.  Maximum number of bytes before base58 encoding
	// is one byte for the network, 32 bytes of private key, possibly one
	// extra byte if the pubkey is to be compressed, and finally four
	// bytes of checksum.
	encodeLen := 1 + btcec.PrivKeyBytesLen + 4
	if w.CompressPubKey {
		encodeLen++
	}

	a := make([]byte, 0, encodeLen)
	a = append(a, w.netID)
	// Pad and append bytes manually, instead of using Serialize, to
	// avoid another call to make.
	a = paddedAppend(btcec.PrivKeyBytesLen, a, w.PrivKey.D.Bytes())
	if w.CompressPubKey {
		a = append(a, compressMagic)
	}
	cksum := common.DoubleHashB(a)[:4]
	a = append(a, cksum...)
	return base58.Encode(a)
}

// String is a helper routine that allocates a new string value to store v and
// returns a pointer to it.  This is useful when assigning optional parameters.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// NewAddressScriptHash returns a new AddressScriptHash.
// NewAddressScriptHashFromHash returns a new AddressScriptHash.  scriptHash
// must be 20 bytes.
func NewAddressScriptHashFromHash(scriptHash []byte, netScriptHashAddrID byte) (*AddressScriptHash, error) {
	return newAddressScriptHashFromHash(scriptHash, netScriptHashAddrID)
}

// GetTxOutResult models the data from the getTransactionsByTxid command.
type GetTxIdResult struct {
	Txid     string `json:"txid"`
	Apptype  string `json:"apptype"`
	Content  []byte `json:"content"`
	Coinbase bool   `json:"coinbase"`
	UnitHash string `json:"unit_hash"`
}

func NewWIF(privKey *ecdsa.PrivateKey, netid byte, compress bool) (*WIF, error) {
	return &WIF{privKey, compress, netid}, nil
}
