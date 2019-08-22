package tokenengine

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
)

type ITokenEngine interface {
	GenerateP2PKHLockScript(pubKeyHash []byte) []byte
	GenerateP2SHLockScript(redeemScriptHash []byte) []byte
	GenerateP2CHLockScript(contractId common.Address) []byte
	GenerateLockScript(address common.Address) []byte
	GetAddressFromScript(lockScript []byte) (common.Address, error)
	GenerateRedeemScript(needed byte, pubKeys [][]byte) []byte
	ScriptValidate(utxoLockScript []byte,
		pickupJuryRedeemScript txscript.PickupJuryRedeemScript,
		tx *modules.Transaction,
		msgIdx, inputIndex int) error
	ScriptValidate1Msg(utxoLockScripts map[string][]byte,
		pickupJuryRedeemScript txscript.PickupJuryRedeemScript,
		tx *modules.Transaction, msgIdx int) error
	GetScriptSigners(tx *modules.Transaction, msgIdx, inputIndex int) ([]common.Address, error)
	MultiSignOnePaymentInput(tx *modules.Transaction,
		hashType uint32, msgIdx, id int,
		utxoLockScript []byte, redeemScript []byte,
		pubKeyFn AddressGetPubKey, hashFn AddressGetSign, previousScript []byte) ([]byte, error)
	CalcSignatureHash(tx *modules.Transaction, hashType uint32,
		msgIdx, inputIdx int, lockOrRedeemScript []byte) ([]byte, error)
	SignTxAllPaymentInput(tx *modules.Transaction, hashType uint32, utxoLockScripts map[modules.OutPoint][]byte,
		redeemScript []byte, pubKeyFn AddressGetPubKey, hashFn AddressGetSign) ([]common.SignatureError, error)
	DisasmString(script []byte) (string, error)
	MergeContractUnlockScript(signs [][]byte, redeemScript []byte) []byte
}
