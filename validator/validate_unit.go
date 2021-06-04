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

	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
)

const ENABLE_TX_FEE_CHECK_TIME = 1570870800             //2019-10-12 17:00:00
const ENABLE_CONTRACT_SIGN_CHECK_TIME = 1588348800      //2019-12-1
const ENABLE_CONTRACT_DEVELOPER_CHECK_TIME = 1577808000 //2020-1-1
const ENABLE_CONTRACT_RWSET_CHECK_TIME = 2588262400     //2020-5-1
const ENABLE_TX_FULL_CHECK_TIME = 1588348800

//1588348800  2020-5-2

/**
验证unit的签名，需要比对见证人列表
To validate unit's signature, and mediators' signature
*/
func validateUnitSignature(header *modules.Header) ValidationCode {
	// copy unit's header
	//header := modules.CopyHeader(h)
	// signature does not contain authors and witness fields
	//emptySigUnit.UnitHeader.Authors = nil
	//header.GroupSign = make([]byte, 0)
	// recover signature
	//if h.Authors == nil {
	author := header.GetAuthors()
	if author.Empty() {
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

	if pass, _ := crypto.MyCryptoLib.Verify(author.PubKey, author.Signature, hash.Bytes()); !pass {
		log.Debug("Verify unit signature error.")
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	// if genesis unit just return
	if len(header.ParentHash()) == 0 {
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
	authorAddr := h.Author()
	if !mediators[authorAddr] {
		mediatorAddrs := ""
		for m := range mediators {
			mediatorAddrs += m.String() + ","
		}
		log.Warnf("Active mediator list is:%s, current unit[%s %d] author is %s",
			mediatorAddrs, h.Hash().String(), h.NumberU64(), authorAddr.String())
		return UNIT_STATE_INVALID_AUTHOR
	}
	return TxValidationCode_VALID
}

// 验证该单元是否是满足调度顺序的mediator生产的
//func (validate *Validate) validateMediatorSchedule(header *modules.Header) ValidationCode {
//	if validate.propquery == nil {
//		log.Warn("Validator don't have propquery, cannot validate mediator schedule")
//		return TxValidationCode_VALID
//	}
//
//	gasToken := dagconfig.DagConfig.GetGasToken()
//	ts, _ := validate.propquery.GetNewestUnitTimestamp(gasToken)
//	if !(header.Timestamp() > ts) {
//		errStr := "invalidated unit's timestamp"
//		log.Warnf("%s,db newest unit timestamp=%d,current unit[%s] timestamp=%d", errStr, ts,
//			header.Hash().String(), header.Timestamp())
//		return UNIT_STATE_INVALID_HEADER_TIME
//	}
//
//	slotNum := validate.propquery.GetSlotAtTime(time.Unix(header.Timestamp(), 0))
//	if slotNum == 0 {
//		log.Warnf("invalidated unit's slot(%v), slotNum must be greater than 0", slotNum)
//		return UNIT_STATE_INVALID_MEDIATOR_SCHEDULE
//	}
//
//	scheduledMediator := validate.propquery.GetScheduledMediator(slotNum)
//	if !scheduledMediator.Equal(header.Author()) {
//		errStr := fmt.Sprintf("mediator(%v) produced unit at wrong time, scheduled slot number:%d, "+
//			"scheduled mediator is %v", header.Author().Str(), slotNum, scheduledMediator.String())
//		log.Warn(errStr)
//		return UNIT_STATE_INVALID_MEDIATOR_SCHEDULE
//	}
//
//	return TxValidationCode_VALID
//}

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

	if len(header.ParentHash()) == 0 {
		log.Info("the header's parentHash is null.")
		return UNIT_STATE_INVALID_HEADER
	}

	//  check header's extra data
	if uint64(len(header.Extra())) > configure.MaximumExtraDataSize {
		msg := fmt.Sprintf("extra-data too long: %d > %d", len(header.Extra()), configure.MaximumExtraDataSize)
		log.Info(msg)
		return UNIT_STATE_INVALID_EXTRA_DATA
	}

	// check creation_time
	if header.Timestamp() <= modules.UNIT_CREATION_DATE_INITIAL_UINT64 {
		return UNIT_STATE_INVALID_HEADER_TIME
	}

	// check header's number
	if header.GetNumber() == nil {
		return UNIT_STATE_INVALID_HEADER_NUMBER
	}
	var thisUnitIsNotTransmitted bool
	if thisUnitIsNotTransmitted {
		sigState := validateUnitSignature(header)
		return sigState
	}
	//validate tx root
	root := core.DeriveSha(unit.Txs)
	if root != unit.UnitHeader.TxRoot() {
		log.Warnf("Validate unit's header failed, root:[%#x],  unit.UnitHeader.TxRoot:[%#x], txs:[%#x]", root, unit.UnitHeader.TxRoot(), unit.Txs.GetTxIds())
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
	//start := time.Now()
	////为每个Unit创建一个Tempdb
	//tempdb,_:=ptndb.NewTempdb(validate.db)
	//origindb:=validate.db
	//validate.db=tempdb
	//defer func() {
	//	validate.db=origindb
	//	log.Debugf("ValidateUnitExceptGroupSig unit[%s],cost:%s", unitHash.String(), time.Since(start))
	//}()

	// 1568197800 2019-09-11 18:30:00 testNet分叉修复后，统一的leveldb
	// 2019-07-11 12:56:46 849c2cb5c7b3fbd37b2ac5f318716f90613259f2 将洗牌算法的种子由时间戳改成hash
	// 并在 1.0.1 版本升级后，在主网和测试网中使用新的调度策略
	//1570870800 20191012 17:00:00 之前的mediator schedule可能验证通不过
	enableMediatorSchedule := unit.UnitHeader.Timestamp() > 1570870800
	// step1. check header ---New unit is no group signature yet
	unitHeaderValidateResult := validate.validateHeaderExceptGroupSig(
		unit.UnitHeader, enableMediatorSchedule)
	if unitHeaderValidateResult != TxValidationCode_VALID &&
		unitHeaderValidateResult != UNIT_STATE_AUTHOR_SIGNATURE_PASSED &&
		unitHeaderValidateResult != UNIT_STATE_ORPHAN {
		log.Debug("Validate unit's header failed.", "error code", unitHeaderValidateResult)
		return unitHeaderValidateResult
	}

	//validate tx root
	root := core.DeriveSha(unit.Txs)
	if root != unit.UnitHeader.TxRoot() {
		log.Warnf("Validate unit's header failed, root:[%#x],  unit.UnitHeader.TxRoot:[%#x], txs:[%#x]", root, unit.UnitHeader.TxRoot(), unit.Txs.GetTxIds())
		return UNIT_STATE_INVALID_HEADER_TXROOT
	}

	// step2. check transactions in unit
	medAdd := unit.Author()
	med := validate.statequery.GetMediator(medAdd)
	if med == nil {
		log.Warnf("validate.statequery.RetrieveMediator %v err", medAdd.Str())
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	validate.enableTxFeeCheck = unit.Timestamp() > ENABLE_TX_FEE_CHECK_TIME                 // 1.0.3升级，支持交易费检查
	validate.enableContractSignCheck = unit.Timestamp() > ENABLE_CONTRACT_SIGN_CHECK_TIME   // 1.0.4升级，支持交易费检查
	validate.enableDeveloperCheck = unit.Timestamp() > ENABLE_CONTRACT_DEVELOPER_CHECK_TIME // 1.0.5升级，支持合约模板部署时的开发者角色检查
	validate.enableContractRwSetCheck = unit.Timestamp() > ENABLE_CONTRACT_RWSET_CHECK_TIME
	validate.enableTxFullCheck = unit.Timestamp() > ENABLE_TX_FULL_CHECK_TIME
	//if validate.enableTxFeeCheck{
	//	log.Infof("Enable tx fee check since %d",unit.Timestamp())
	//}
	rwM, err := rwset.NewRwSetMgr(unit.NumberString())
	if err != nil {
		log.Errorf("NewRwSetMgr error:%s", err.Error())
		return TxValidationCode_INVALID_OTHER_REASON
	}
	code := validate.validateTransactions(rwM, unit.Txs, unit.Timestamp(), med.GetRewardAdd())
	rwM.Close()
	if code != TxValidationCode_VALID {
		msg := fmt.Sprintf("Validate unit(%s) transactions failed code: %v", unit.DisplayId(), code)
		log.Error(msg)
		return code
	}
	//maybe orphan unit
	if unitHeaderValidateResult != TxValidationCode_VALID {
		return unitHeaderValidateResult
	}
	validate.cache.AddUnitValidateResult(unitHash, TxValidationCode_VALID)
	return TxValidationCode_VALID
}

func (validate *Validate) validateHeaderExceptGroupSig(header *modules.Header, enableMediatorSchedule bool) ValidationCode {
	if header == nil {
		log.Info("header is nil.")
		return UNIT_STATE_INVALID_HEADER
	}

	if len(header.ParentHash()) == 0 {
		log.Info("the header's parentHash is null.")
		return UNIT_STATE_INVALID_HEADER
	}

	//  check header's extra data
	if uint64(len(header.Extra())) > configure.MaximumExtraDataSize {
		msg := fmt.Sprintf("extra-data too long: %d > %d", len(header.Extra()), configure.MaximumExtraDataSize)
		log.Info(msg)
		return UNIT_STATE_INVALID_EXTRA_DATA
	}

	// Only check txroot when has unit body
	//if header.TxRoot == (common.Hash{}) {
	//	log.Info("the header's txroot is null.")
	//	return UNIT_STATE_INVALID_HEADER_TXROOT
	//}

	// check creation_time
	if header.Timestamp() <= modules.UNIT_CREATION_DATE_INITIAL_UINT64 {
		return UNIT_STATE_INVALID_HEADER_TIME
	}

	// check header's number
	if header.GetNumber() == nil {
		return UNIT_STATE_INVALID_HEADER_NUMBER
	}
	var thisUnitIsNotTransmitted bool
	if thisUnitIsNotTransmitted {
		sigState := validateUnitSignature(header)
		return sigState
	}
	//Check author
	if !validate.light { //轻节点无法验证Mediator
		validateAuthorCode := validate.validateUnitAuthor(header)
		if validateAuthorCode != TxValidationCode_VALID {
			return validateAuthorCode
		}
	}
	//Is orphan?
	parent := header.ParentHash()[0]
	if validate.dagquery != nil {
		parentHeader, err := validate.dagquery.GetHeaderByHash(parent)
		if err != nil {
			return UNIT_STATE_ORPHAN
		}
		if parentHeader.GetNumber().Index+1 != header.GetNumber().Index {
			return UNIT_STATE_INVALID_HEADER_NUMBER
		}

		//if enableMediatorSchedule && !validate.light {
		//	vcode := validate.validateMediatorSchedule(header)
		//	if vcode != TxValidationCode_VALID {
		//		return vcode
		//	}
		//}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) ValidateHeader(h *modules.Header) ValidationCode {
	hash := h.Hash()
	has, code := validate.cache.HasHeaderValidateResult(hash)
	if has {
		return code
	}
	unitHeaderValidateResult := validate.validateHeaderExceptGroupSig(h, false)
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
