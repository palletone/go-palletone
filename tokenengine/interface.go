package tokenengine

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type ITokenEngine interface {
	//根据公钥Hash，生成锁定脚本
	GenerateP2PKHLockScript(pubKeyHash []byte) []byte
	//根据赎回脚本Hash，生成锁定脚本
	GenerateP2SHLockScript(redeemScriptHash []byte) []byte
	//根据合约ID，生成锁定脚本
	GenerateP2CHLockScript(contractId common.Address) []byte
	//根据地址生成锁定脚本
	GenerateLockScript(address common.Address) []byte
	//根据锁定脚本，得出对应的地址
	GetAddressFromScript(lockScript []byte) (common.Address, error)
	//根据公钥列表和需要的签名数，获得赎回脚本
	GenerateRedeemScript(needed byte, pubKeys [][]byte) []byte
	//将一个脚本二进制解析为字符串形式
	DisasmString(script []byte) (string, error)
	//计算要对一个Tx的msgIdx和inputInx位置进行签名，对应的Hash
	CalcSignatureHash(tx *modules.Transaction, hashType uint32,
		msgIdx, inputIdx int, lockOrRedeemScript []byte) ([]byte, error)
	//根据tx，msgIdx和inputIdx，获得多签的解锁脚本，然后计算获得解锁脚本是哪些地址进行的该签名，主要用于计算手续费分摊
	GetScriptSigners(tx *modules.Transaction, msgIdx, inputIndex int) ([]common.Address, error)
	//对一个未签名的tx进行签名，将所有input的解锁脚本填充完毕
	SignTxAllPaymentInput(tx *modules.Transaction, hashType uint32, utxoLockScripts map[modules.OutPoint][]byte,
		redeemScript []byte, pubKeyFn AddressGetPubKey, hashFn AddressGetSign) ([]common.SignatureError, error)
	//对tx的某个多签input进行签名，如果已经有别人签名，则合并
	MultiSignOnePaymentInput(tx *modules.Transaction,
		hashType uint32, msgIdx, id int,
		utxoLockScript []byte, redeemScript []byte,
		pubKeyFn AddressGetPubKey, hashFn AddressGetSign, previousScript []byte) ([]byte, error)
	//当用户合约需要解锁Token时，收集到了jury签名列表和赎回脚本，组合成合约的解锁脚本
	MergeContractUnlockScript(signs [][]byte, redeemScript []byte) []byte
	//验证一个tx的指定input的解锁脚本是否正确
	ScriptValidate(utxoLockScript []byte,
		pickupJuryRedeemScript PickupJuryRedeemScript,
		tx *modules.Transaction,
		msgIdx, inputIndex int) error
	//验证tx中的某个Payment message的所有input的解锁脚本是否正确
	ScriptValidate1Msg(utxoLockScripts map[string][]byte,
		pickupJuryRedeemScript PickupJuryRedeemScript,
		tx *modules.Transaction, msgIdx int) error
}
