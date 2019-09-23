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

	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

/**
验证unit的签名，需要比对见证人列表
To validate unit's signature, and mediators' signature
*/
func validateUnitSignature(h *modules.Header) ValidationCode {
	// copy unit's header
	//header := modules.CopyHeader(h)
	// signature does not contain authors and witness fields
	//emptySigUnit.UnitHeader.Authors = nil
	//header.GroupSign = make([]byte, 0)
	// recover signature
	//if h.Authors == nil {
	if h.Authors.Empty() {
		log.Debug("Verify unit signature ,header's authors is nil.")
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}

	hash := h.HashWithoutAuthor()
	//pubKey, err := modules.RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	//if err != nil {
	//	log.Debug("Verify unit signature when recover pubkey", "error", err.Error())
	//	return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	//}
	//  pubKey to pubKey_bytes
	//pubKey_bytes := crypto.CompressPubkey(pubKey)

	if pass, _ := crypto.MyCryptoLib.Verify(h.Authors.PubKey, h.Authors.Signature, hash.Bytes()); !pass {
		log.Debug("Verify unit signature error.")
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	// if genesis unit just return
	if len(h.ParentsHash) == 0 {
		return TxValidationCode_VALID
	}

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

//验证Author必须是一个活跃的Mediator，防止其他节点冒充产块
func (validate *Validate) validateUnitAuthor(h *modules.Header) ValidationCode {
	if validate.statequery == nil {
		log.Warnf("Don't set validate.statequery, cannot validate unit author is a mediator.")
		return TxValidationCode_VALID
	}

	mediators := validate.statequery.GetMediators()
	authorAddr := h.Authors.Address()
	if !mediators[authorAddr] {
		mediatorAddrs := ""
		for m := range mediators {
			mediatorAddrs += m.String() + ","
		}
		log.Warnf("Active mediator list is:%s, current unit[%s %d] author is %s",
			mediatorAddrs, h.Hash(), h.NumberU64(), authorAddr.String())
		return UNIT_STATE_INVALID_AUTHOR
	}
	return TxValidationCode_VALID
}

// 验证该单元是否是满足调度顺序的mediator生产的
func (validate *Validate) validateMediatorSchedule(header *modules.Header) ValidationCode {
	if validate.propquery == nil {
		log.Warn("Validator don't have propquery, cannot validate mediator schedule")
		return TxValidationCode_VALID
	}

	gasToken := dagconfig.DagConfig.GetGasToken()
	ts, _ := validate.propquery.GetNewestUnitTimestamp(gasToken)
	if !(header.Time > ts) {
		errStr := "invalidated unit's timestamp"
		log.Warnf("%s,db newest unit timestamp=%d,current unit[%s] timestamp=%d", errStr, ts,
			header.Hash().String(), header.Time)
		return UNIT_STATE_INVALID_HEADER_TIME
	}

	slotNum := validate.propquery.GetSlotAtTime(time.Unix(header.Time, 0))
	if slotNum <= 0 {
		log.Info("invalidated unit's slot")
		return UNIT_STATE_INVALID_MEDIATOR_SCHEDULE
	}

	scheduledMediator := validate.propquery.GetScheduledMediator(slotNum)
	if !scheduledMediator.Equal(header.Author()) {
		errStr := fmt.Sprintf("mediator(%v) produced unit at wrong time", header.Author().Str())
		log.Warn(errStr)
		return UNIT_STATE_INVALID_MEDIATOR_SCHEDULE
	}

	return TxValidationCode_VALID
}

//不基于数据库，进行Unit最基本的验证
func ValidateUnitBasic(unit *modules.Unit) error {
	return NewValidateError(validateUnitBasic(unit))
}

//不基于数据库，进行Unit最基本的验证
func validateUnitBasic(unit *modules.Unit) ValidationCode {
	header := unit.UnitHeader
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
		sigState := validateUnitSignature(header)
		return sigState
	}
	//validate tx root
	root := core.DeriveSha(unit.Txs)
	if root != unit.UnitHeader.TxRoot {
		log.Debugf("Validate unit's header failed, root:[%#x],  unit.UnitHeader.TxRoot:[%#x], txs:[%#x]", root, unit.UnitHeader.TxRoot, unit.Txs.GetTxIds())
		return UNIT_STATE_INVALID_HEADER_TXROOT
	}

	return TxValidationCode_VALID
}

/**
验证Unit
Validate unit(除群签名以外), 新生产的unit暂时还没有群签名
*/
//func (validate *Validate) ValidateUnit(unit *modules.Unit, isGenesis bool) byte {
func (validate *Validate) ValidateUnitExceptGroupSig(unit *modules.Unit) ValidationCode {
	unitHash := unit.Hash()
	if has, code := validate.cache.HasUnitValidateResult(unitHash); has {
		return code
	}
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
		return unitHeaderValidateResult
	}

	//validate tx root
	root := core.DeriveSha(unit.Txs)
	if root != unit.UnitHeader.TxRoot {
		log.Debugf("Validate unit's header failed, root:[%#x],  unit.UnitHeader.TxRoot:[%#x], txs:[%#x]", root, unit.UnitHeader.TxRoot, unit.Txs.GetTxIds())
		return UNIT_STATE_INVALID_HEADER_TXROOT
	}

	// step2. check transactions in unit
	medAdd := unit.Author()
	med := validate.statequery.GetMediator(medAdd)
	if med == nil {
		log.Debugf("validate.statequery.RetrieveMediator %v err", medAdd.Str())
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}

	code := validate.validateTransactions(unit.Txs, unit.Timestamp(), med.GetRewardAdd())
	if code != TxValidationCode_VALID {
		msg := fmt.Sprintf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), code)
		log.Debug(msg)
		return code
	}
	//maybe orphan unit
	if unitHeaderValidateResult != TxValidationCode_VALID {
		return unitHeaderValidateResult
	}
	validate.cache.AddUnitValidateResult(unitHash, TxValidationCode_VALID)
	return TxValidationCode_VALID
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

	// Only check txroot when has unit body
	//if header.TxRoot == (common.Hash{}) {
	//	log.Info("the header's txroot is null.")
	//	return UNIT_STATE_INVALID_HEADER_TXROOT
	//}

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
		sigState := validateUnitSignature(header)
		return sigState
	}
	//Check author
	validateAuthorCode := validate.validateUnitAuthor(header)
	if validateAuthorCode != TxValidationCode_VALID {
		return validateAuthorCode
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
		if header.Time > 1564675200 { //2019.8.2主网升级，有些之前的mediator schedule可能验证不过。
			vcode := validate.validateMediatorSchedule(header)
			if vcode != TxValidationCode_VALID {
				return vcode
			}
		}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) ValidateHeader(h *modules.Header) ValidationCode {
	hash := h.Hash()
	has, code := validate.cache.HasHeaderValidateResult(hash)
	if has {
		return code
	}
	unitHeaderValidateResult := validate.validateHeaderExceptGroupSig(h)
	if unitHeaderValidateResult != TxValidationCode_VALID &&
		unitHeaderValidateResult != UNIT_STATE_AUTHOR_SIGNATURE_PASSED &&
		unitHeaderValidateResult != UNIT_STATE_ORPHAN {
		log.Debug("Validate unit's header failed.", "error code", unitHeaderValidateResult)
		return unitHeaderValidateResult
	}
	if unitHeaderValidateResult == TxValidationCode_VALID {
		validate.cache.AddHeaderValidateResult(hash, unitHeaderValidateResult)
	}
	return TxValidationCode_VALID
}
