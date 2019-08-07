// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"errors"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

// RawTxInSignature returns the serialized ECDSA signature for the input idx of
// the given transaction, with hashType appended to it.
func RawTxInSignature(tx *modules.Transaction, msgIdx, idx int, subScript []byte,
	hashType SigHashType, crypto ICrypto, addr common.Address) ([]byte, error) {

	parsedScript, err := parseScript(subScript)
	if err != nil {
		return nil, fmt.Errorf("cannot parse output script: %v", err)
	}
	data := calcSignatureData(parsedScript, hashType, tx, msgIdx, idx)
	sign, err := crypto.Sign(addr, data)
	if err != nil {
		return nil, fmt.Errorf("cannot sign tx input: %s", err)
	}
	return append(sign, byte(hashType)), nil
	//signature, err := key.Sign(hash)
	//if err != nil {
	//	return nil, fmt.Errorf("cannot sign tx input: %s", err)
	//}
	//
	//return append(signature.Serialize(), byte(hashType)), nil
}

// SignatureScript creates an input signature script for tx to spend BTC sent
// from a previous output to the owner of privKey. tx must include all
// transaction inputs and outputs, however txin scripts are allowed to be filled
// or empty. The returned script is calculated to be used as the idx'th txin
// sigscript for tx. subscript is the PkScript of the previous output being used
// as the idx'th input. privKey is serialized in either a compressed or
// uncompressed format based on compress. This format must match the same format
// used to generate the payment address, or the script validation will fail.
func SignatureScript(tx *modules.Transaction, msgIdx, idx int, 
	subscript []byte, hashType SigHashType, pubKey []byte, 
	crypto ICrypto, addr common.Address) ([]byte, error) {
	sig, err := RawTxInSignature(tx, msgIdx, idx, subscript, hashType, crypto, addr)
	if err != nil {
		return nil, err
	}

	//pk := (*btcec.PublicKey)(&privKey.PublicKey)
	//var pkData []byte
	//if compress {
	//	pkData = pk.SerializeCompressed()
	//} else {
	//	pkData = pk.SerializeUncompressed()
	//}

	return NewScriptBuilder().AddData(sig).AddData(pubKey).Script()
}

func p2pkSignatureScript(tx *modules.Transaction, msgIdx, idx int, 
	subScript []byte, hashType SigHashType, 
	crypto ICrypto, addr common.Address) ([]byte, error) {
	sig, err := RawTxInSignature(tx, msgIdx, idx, subScript, hashType, crypto, addr)
	if err != nil {
		return nil, err
	}

	return NewScriptBuilder().AddData(sig).Script()
}

// signMultiSig signs as many of the outputs in the provided multisig script as
// possible. It returns the generated script and a boolean if the script fulfills
// the contract (i.e. nrequired signatures are provided).  Since it is arguably
// legal to not be able to sign any of the outputs, no error is returned.
func signMultiSig(tx *modules.Transaction, msgIdx, idx int, 
	subScript []byte, hashType SigHashType,
	addresses []AddressOriginalData, nRequired int, crypto ICrypto) ([]byte, bool) {
	// We start with a single OP_FALSE to work around the (now standard)
	// but in the reference implementation that causes a spurious pop at
	// the end of OP_CHECKMULTISIG.
	builder := NewScriptBuilder().AddOp(OP_FALSE)
	signed := 0
	for _, addr := range addresses {

		sig, err := RawTxInSignature(tx, msgIdx, idx, subScript, hashType, crypto, addr.Address)
		if err != nil {
			continue
		}

		builder.AddData(sig)
		signed++
		if signed == nRequired {
			break
		}

	}

	script, _ := builder.Script()
	return script, signed == nRequired
}

func sign(tx *modules.Transaction, msgIdx, idx int,
	subScript []byte, hashType SigHashType, kdb ICrypto, sdb ScriptDB) ([]byte,
	ScriptClass, []AddressOriginalData, int, error) {

	class, addresses, nrequired, err := ExtractPkScriptAddrs(subScript)
	if err != nil {
		return nil, NonStandardTy, nil, 0, err
	}

	switch class {
	case PubKeyTy:
		// look up key for address
		//key, _, err := kdb.GetKey(addresses[0].Address)
		//if err != nil {
		//	return nil, class, nil, 0, err
		//}

		script, err := p2pkSignatureScript(tx, msgIdx, idx, subScript, hashType,
			kdb, addresses[0].Address)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil
	case PubKeyHashTy:
		// look up key for address
		pubkey, err := kdb.GetPubKey(addresses[0].Address)
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := SignatureScript(tx, msgIdx, idx, subScript, hashType,
			pubkey, kdb, addresses[0].Address)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil
	case ScriptHashTy:
		script, err := sdb.GetScript(addresses[0].Address)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil
	case ContractHashTy:
		script, err := sdb.GetScript(addresses[0].Address)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil
	case MultiSigTy:
		script, _ := signMultiSig(tx, msgIdx, idx, subScript, hashType,
			addresses, nrequired, kdb)
		return script, class, addresses, nrequired, nil
	case NullDataTy:
		return nil, class, nil, 0,
			errors.New("can't sign NULLDATA transactions")
	default:
		return nil, class, nil, 0,
			errors.New("can't sign unknown transactions")
	}
}

// mergeScripts merges sigScript and prevScript assuming they are both
// partial solutions for pkScript spending output idx of tx. class, addresses
// and nrequired are the result of extracting the addresses from pkscript.
// The return value is the best effort merging of the two scripts. Calling this
// function with addresses, class and nrequired that do not match pkScript is
// an error and results in undefined behavior.
func mergeScripts(tx *modules.Transaction, msgIdx, idx int,
	pkScript []byte, class ScriptClass, addresses []AddressOriginalData,
	nRequired int, sigScript, prevScript []byte, crypto ICrypto) []byte {

	// TODO(oga) the scripthash and multisig paths here are overly
	// inefficient in that they will recompute already known data.
	// some internal refactoring could probably make this avoid needless
	// extra calculations.
	switch class {
	case ScriptHashTy:
		// Remove the last push in the script and then recurse.
		// this could be a lot less inefficient.
		sigPops, err := parseScript(sigScript)
		if err != nil || len(sigPops) == 0 {
			return prevScript
		}
		prevPops, err := parseScript(prevScript)
		if err != nil || len(prevPops) == 0 {
			return sigScript
		}

		// assume that script in sigPops is the correct one, we just
		// made it.
		script := sigPops[len(sigPops)-1].data

		// We already know this information somewhere up the stack.
		class, addresses, nrequired, _ :=
			ExtractPkScriptAddrs(script)

		// regenerate scripts.
		sigScript, _ := unparseScript(sigPops)
		prevScript, _ := unparseScript(prevPops)

		// Merge
		mergedScript := mergeScripts(tx, msgIdx, idx, script,
			class, addresses, nrequired, sigScript, prevScript, crypto)

		// Reappend the script and return the result.
		builder := NewScriptBuilder()
		builder.script = mergedScript
		builder.AddData(script)
		finalScript, _ := builder.Script()
		return finalScript
	case MultiSigTy:
		return mergeMultiSig(tx, msgIdx, idx, addresses, nRequired, pkScript,
			sigScript, prevScript, crypto)

	// It doesn't actually make sense to merge anything other than multiig
	// and scripthash (because it could contain multisig). Everything else
	// has either zero signature, can't be spent, or has a single signature
	// which is either present or not. The other two cases are handled
	// above. In the conflict case here we just assume the longest is
	// correct (this matches behavior of the reference implementation).
	default:
		if len(sigScript) > len(prevScript) {
			return sigScript
		}
		return prevScript
	}
}

// mergeMultiSig combines the two signature scripts sigScript and prevScript
// that both provide signatures for pkScript in output idx of tx. addresses
// and nRequired should be the results from extracting the addresses from
// pkScript. Since this function is internal only we assume that the arguments
// have come from other functions internally and thus are all consistent with
// each other, behavior is undefined if this contract is broken.
func mergeMultiSig(tx *modules.Transaction, msgIdx, idx int, addresses []AddressOriginalData,
	nRequired int, pkScript, sigScript, prevScript []byte, crypto ICrypto) []byte {

	// This is an internal only function and we already parsed this script
	// as ok for multisig (this is how we got here), so if this fails then
	// all assumptions are broken and who knows which way is up?
	pkPops, _ := parseScript(pkScript)

	sigPops, err := parseScript(sigScript)
	if err != nil || len(sigPops) == 0 {
		return prevScript
	}

	prevPops, err := parseScript(prevScript)
	if err != nil || len(prevPops) == 0 {
		return sigScript
	}

	// Convenience function to avoid duplication.
	extractSigs := func(pops []parsedOpcode, sigs [][]byte) [][]byte {
		for _, pop := range pops {
			if len(pop.data) != 0 {
				sigs = append(sigs, pop.data)
			}
		}
		return sigs
	}

	possibleSigs := make([][]byte, 0, len(sigPops)+len(prevPops))
	possibleSigs = extractSigs(sigPops, possibleSigs)
	possibleSigs = extractSigs(prevPops, possibleSigs)

	// Now we need to match the signatures to pubkeys, the only real way to
	// do that is to try to verify them all and match it to the pubkey
	// that verifies it. we then can go through the addresses in order
	// to build our script. Anything that doesn't parse or doesn't verify we
	// throw away.
	addrToSig := make(map[string][]byte)
sigLoop:
	for _, sig := range possibleSigs {

		// can't have a valid signature that doesn't at least have a
		// hashtype, in practice it is even longer than this. but
		// that'll be checked next.
		if len(sig) < 1 {
			continue
		}
		tSig := sig[:len(sig)-1]
		hashType := SigHashType(sig[len(sig)-1])

		//pSig, err := btcec.ParseDERSignature(tSig, btcec.S256())
		//if err != nil {
		//	continue
		//}

		// We have to do this each round since hash types may vary
		// between signatures and so the hash will vary. We can,
		// however, assume no sigs etc are in the script since that
		// would make the transaction nonstandard and thus not
		// MultiSigTy, so we just need to hash the full thing.
		data := calcSignatureData(pkPops, hashType, tx, msgIdx, idx)

		for _, addr := range addresses {
			// All multisig addresses should be pubkey addreses
			// it is an error to call this internal function with
			// bad input.
			//pkaddr := addr.(*common.AddressPubKey)
			pubKey := addr.Original
			//pubKey := pkaddr.PubKey()

			// If it matches we put it in the map. We only
			// can take one signature per public key so if we
			// already have one, we can throw this away.
			if pass, _ := crypto.Verify(pubKey, tSig, data); pass {
				//if pSig.Verify(hash, pubKey) {
				aStr := addr.Address.String()
				if _, ok := addrToSig[aStr]; !ok {
					addrToSig[aStr] = sig
				}
				continue sigLoop
			}
		}
	}

	// Extra opcode to handle the extra arg consumed (due to previous bugs
	// in the reference implementation).
	builder := NewScriptBuilder().AddOp(OP_FALSE)
	doneSigs := 0
	// This assumes that addresses are in the same order as in the script.
	for _, addr := range addresses {
		sig, ok := addrToSig[addr.Address.String()]
		if !ok {
			continue
		}
		builder.AddData(sig)
		doneSigs++
		if doneSigs == nRequired {
			break
		}
	}

	// padding for missing ones.
	for i := doneSigs; i < nRequired; i++ {
		builder.AddOp(OP_0)
	}

	script, _ := builder.Script()
	return script
}

// KeyDB is an interface type provided to SignTxOutput, it encapsulates
// any user state required to get the private keys for an address.
//type KeyDB interface {
//	GetSignFunction(common.Address) SignHash
//	GetPubKey(common.Address) ([]byte, error)
//}
//
//// KeyClosure implements KeyDB with a closure.
//type KeyClosure func(common.Address, []byte) ([]byte, error)

// GetKey implements KeyDB by returning the result of calling the closure.
//func (kc KeyClosure) GetKey(address common.Address) (*btcec.PrivateKey,
//	bool, error) {
//	return kc(address)
//}

// ScriptDB is an interface type provided to SignTxOutput, it encapsulates any
// user state required to get the scripts for an pay-to-script-hash address.
type ScriptDB interface {
	GetScript(common.Address) ([]byte, error)
}

// ScriptClosure implements ScriptDB with a closure.
type ScriptClosure func(common.Address) ([]byte, error)

// GetScript implements ScriptDB by returning the result of calling the closure.
func (sc ScriptClosure) GetScript(address common.Address) ([]byte, error) {
	return sc(address)
}

// SignTxOutput signs output idx of the given tx to resolve the script given in
// pkScript with a signature type of hashType. Any keys required will be
// looked up by calling getKey() with the string of the given address.
// Any pay-to-script-hash signatures will be similarly looked up by calling
// getScript. If previousScript is provided then the results in previousScript
// will be merged in a type-dependant manner with the newly generated.
// signature script.
func SignTxOutput(tx *modules.Transaction, msgIdx, idx int,
	pkScript []byte, hashType SigHashType, crypto ICrypto, sdb ScriptDB,
	previousScript []byte) ([]byte, error) {

	sigScript, class, addresses, nrequired, err := sign(tx, msgIdx,
		idx, pkScript, hashType, crypto, sdb)
	if err != nil {
		return nil, err
	}

	if class == ScriptHashTy || class == ContractHashTy {
		// TODO keep the sub addressed and pass down to merge.
		realSigScript, _, _, _, err := sign(tx, msgIdx, idx,
			sigScript, hashType, crypto, sdb)
		if err != nil {
			return nil, err
		}

		// This is a bad thing. Append the p2sh script as the last
		// push in the script.
		builder := NewScriptBuilder()
		builder.script = realSigScript
		builder.AddData(sigScript)

		sigScript, _ = builder.Script()
		// TODO keep a copy of the script for merging.
	}

	// Merge scripts. with any previous data, if any.
	mergedScript := mergeScripts(tx, msgIdx, idx, pkScript, class,
		addresses, nrequired, sigScript, previousScript, crypto)
	return mergedScript, nil
}
