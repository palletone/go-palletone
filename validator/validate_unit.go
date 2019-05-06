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
 *  * @date 2018-2019
 *
 */

package validator

import (
	"fmt"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

/**
验证unit的签名，需要比对见证人列表
To validate unit's signature, and mediators' signature
*/
func (validate *Validate) validateUnitSignature(h *modules.Header) ValidationCode {
	// copy unit's header
	header := modules.CopyHeader(h)
	// signature does not contain authors and witness fields
	//emptySigUnit.UnitHeader.Authors = nil
	header.GroupSign = make([]byte, 0)
	// recover signature
	//if h.Authors == nil {
	if h.Authors.Empty() {
		log.Debug("Verify unit signature ,header's authors is nil.")
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}

	hash := header.HashWithoutAuthor()
	//pubKey, err := modules.RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	//if err != nil {
	//	log.Debug("Verify unit signature when recover pubkey", "error", err.Error())
	//	return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	//}
	//  pubKey to pubKey_bytes
	//pubKey_bytes := crypto.CompressPubkey(pubKey)

	if !crypto.VerifySignature(h.Authors.PubKey, hash.Bytes(), h.Authors.Signature) {
		log.Debug("Verify unit signature error.")
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	// if genesis unit just return
	//if isGenesis == true {
	//	return TxValidationCode_VALID
	//}

	// get mediators
	//TODO Devin
	//data, _ := validate.statedb.GetCandidateMediatorAddrList() //.GetConfig([]byte("MediatorCandidates"))
	//var mList []core.MediatorInfo
	//if err := rlp.DecodeBytes(data, &mList); err != nil {
	//	log.Debug("Check unit signature when get mediators list", "error", err.Error())
	//	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	//}
	//bNum, _ := validate.statedb.GetActiveMediatorAddrList()
	//var mNum uint16
	//if err := rlp.DecodeBytes(bNum, &mNum); err != nil {
	//	log.Debug("Check unit signature", "error", err.Error())
	//	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	//}
	//if int(mNum) != len(mList) {
	//	log.Debug("Check unit signature", "error", "mediators info error, pls update network")
	//	return modules.UNIT_STATE_INVALID_GROUP_SIGNATURE
	//}
	// 这一步后续添加： 调用 mediator 模块校验见证人的接口

	//return modules.UNIT_STATE_VALIDATED
	return TxValidationCode_VALID
}

/**
验证Unit
Validate unit(除群签名以外), 新生产的unit暂时还没有群签名
*/
//func (validate *Validate) ValidateUnit(unit *modules.Unit, isGenesis bool) byte {
func (validate *Validate) ValidateUnitExceptGroupSig(unit *modules.Unit) error {
	start := time.Now()
	defer func() {
		log.Debugf("ValidateUnitExceptGroupSig unit[%s],cost:%s", unit.Hash().String(), time.Since(start))
	}()

	// step1. check header ---New unit is no group signature yet
	unitHeaderValidateResult := validate.validateHeaderExceptGroupSig(unit.UnitHeader)
	if unitHeaderValidateResult != TxValidationCode_VALID &&
		unitHeaderValidateResult != UNIT_STATE_AUTHOR_SIGNATURE_PASSED &&
		unitHeaderValidateResult != UNIT_STATE_ORPHAN {
		log.Debug("Validate unit's header failed.", "error code", unitHeaderValidateResult)
		return NewValidateError(unitHeaderValidateResult)
	}

	//validate tx root
	root := core.DeriveSha(unit.Txs)
	if root != unit.UnitHeader.TxRoot {
		log.Debugf("Validate unit's header failed, root:[%#x],  unit.UnitHeader.TxRoot:[%#x], txs:[%#x]", root, unit.UnitHeader.TxRoot, unit.Txs.GetTxIds())
		return NewValidateError(UNIT_STATE_INVALID_HEADER_TXROOT)
	}
	// step2. check transactions in unit
	code := validate.validateTransactions(unit.Txs, unit.Timestamp())
	if code != TxValidationCode_VALID {
		msg := fmt.Sprintf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), code)
		log.Debug(msg)
		return NewValidateError(code)
	}
	//maybe orphan unit
	if unitHeaderValidateResult != TxValidationCode_VALID {
		return NewValidateError(unitHeaderValidateResult)
	}
	return nil
}

func (validate *Validate) validateHeaderExceptGroupSig(header *modules.Header) ValidationCode {
	if header == nil {
		log.Info("header is nil.")
		return UNIT_STATE_INVALID_HEADER
	}

	if len(header.ParentsHash) == 0 {
		log.Info("the header's parentHash is null.")
		return UNIT_STATE_INVALID_HEADER
	}

	//  check header's extra data
	if uint64(len(header.Extra)) > configure.MaximumExtraDataSize {
		msg := fmt.Sprintf("extra-data too long: %d > %d", len(header.Extra), configure.MaximumExtraDataSize)
		log.Info(msg)
		return UNIT_STATE_INVALID_EXTRA_DATA
	}

	// check txroot
	if header.TxRoot == (common.Hash{}) {
		log.Info("the header's txroot is null.")
		return UNIT_STATE_INVALID_HEADER_TXROOT
	}

	// check creation_time
	if header.Time <= modules.UNIT_CREATION_DATE_INITIAL_UINT64 {
		return UNIT_STATE_INVALID_HEADER_TIME
	}

	// check header's number
	if header.Number == nil {
		return UNIT_STATE_INVALID_HEADER_NUMBER
	}
	var thisUnitIsNotTransmitted bool
	if thisUnitIsNotTransmitted {
		sigState := validate.validateUnitSignature(header)
		return sigState
	}
	//Is orphan?
	parent := header.ParentsHash[0]
	if validate.dagquery != nil {
		parentHeader, err := validate.dagquery.GetHeaderByHash(parent)
		if err != nil {
			return UNIT_STATE_ORPHAN
		}
		if parentHeader.Number.Index+1 != header.Number.Index {
			return UNIT_STATE_INVALID_HEADER_NUMBER
		}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) ValidateHeader(h *modules.Header) error {
	return nil
}
