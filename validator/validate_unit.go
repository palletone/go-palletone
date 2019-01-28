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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/dag/modules"
)

/**
验证单元的签名，需要比对见证人列表
To validate unit's signature, and mediators' signature
*/
func (validate *Validate) validateUnitSignature(h *modules.Header, isGenesis bool) ValidationCode {

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
	sig := make([]byte, 65)
	copy(sig[32-len(h.Authors.R):32], h.Authors.R)
	copy(sig[64-len(h.Authors.S):64], h.Authors.S)
	copy(sig[64:], h.Authors.V)
	// recover pubkey

	hash := header.HashWithoutAuthor()
	pubKey, err := modules.RSVtoPublicKey(hash[:], h.Authors.R[:], h.Authors.S[:], h.Authors.V[:])
	if err != nil {
		log.Debug("Verify unit signature when recover pubkey", "error", err.Error())
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	//  pubKey to pubKey_bytes
	pubKey_bytes := crypto.CompressPubkey(pubKey)

	if !crypto.VerifySignature(pubKey_bytes, hash.Bytes(), sig[:64]) {
		log.Debug("Verify unit signature error.")
		return UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	}
	// if genesis unit just return
	if isGenesis == true {
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

/**
验证Unit
Validate unit
*/
// modified by Albert·Gou 新生产的unit暂时还没有群签名
//func (validate *Validate) ValidateUnit(unit *modules.Unit, isGenesis bool) byte {
func (validate *Validate) ValidateUnitExceptGroupSig(unit *modules.Unit, isGenesis bool) error {
	//  unit's size  should bigger than minimum.
	if unit.Size() < 125 {
		log.Debug("Validate size", "error", "size is invalid", "size", unit.Size())
		return NewValidateError(UNIT_STATE_INVALID_SIZE)
	}

	// step1. check header.New unit is no group signature yet
	//TODO must recover

	sigState := validate.validateHeaderExceptGroupSig(unit.UnitHeader, isGenesis)
	if sigState != modules.UNIT_STATE_VALIDATED &&
		sigState != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED && sigState != modules.UNIT_STATE_CHECK_HEADER_PASSED {
		log.Debug("Validate unit's header failed.", "error code", sigState)
		return NewValidateError(sigState)
	}

	// step2. check transactions in unit
	//_, isSuccess, err := validate.ValidateTransactions(&unit.Txs, isGenesis)
	isSuccess := true //TODO test for sync
	var err error
	if isSuccess != true {
		msg := fmt.Sprintf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
		log.Debug(msg)
		return NewValidateError(UNIT_STATE_HAS_INVALID_TRANSACTIONS)
	}
	return NewValidateError(sigState)
}

// modified by Albert·Gou 新生产的unit暂时还没有群签名
//func (validate *Validate) validateHeader(header *modules.Header, isGenesis bool) byte {
func (validate *Validate) validateHeaderExceptGroupSig(header *modules.Header, isGenesis bool) ValidationCode {
	// todo yangjie 应当错误返回前，打印验错误的具体消息
	if header == nil {
		return UNIT_STATE_INVALID_HEADER
	}

	if len(header.ParentsHash) == 0 {
		if !isGenesis {
			return UNIT_STATE_INVALID_HEADER
		}
	}
	//  check header's extra data
	if uint64(len(header.Extra)) > configure.MaximumExtraDataSize {
		msg := fmt.Sprintf("extra-data too long: %d > %d", len(header.Extra), configure.MaximumExtraDataSize)
		log.Debug(msg)
		return UNIT_STATE_INVALID_EXTRA_DATA
	}
	// check txroot
	if header.TxRoot == (common.Hash{}) {
		return UNIT_STATE_INVALID_HEADER
	}

	// check creation_time
	if header.Creationdate <= modules.UNIT_CREATION_DATE_INITIAL_UINT64 {
		return UNIT_STATE_INVALID_HEADER
	}

	// check header's number
	if header.Number == nil {
		return UNIT_STATE_INVALID_HEADER
	}
	//if len(header.AssetIDs) == 0 {
	//	return modules.UNIT_STATE_INVALID_HEADER
	//}

	//if isGenesis {
	//	if len(header.AssetIDs) != 1 {
	//		return modules.UNIT_STATE_INVALID_HEADER
	//	}
	//	//ptnAssetID, _ := modules.SetIdTypeByHex(dagconfig.DefaultConfig.PtnAssetHex)
	//	asset := modules.NewPTNAsset()
	//	ptnAssetID := asset.AssetId
	//	if header.AssetIDs[0] != ptnAssetID || !header.Number.IsMain || header.Number.Index != 0 {
	//		fmt.Println(6)
	//		fmt.Println(header.AssetIDs[0].String())
	//		fmt.Println(ptnAssetID.String())
	//		return modules.UNIT_STATE_INVALID_HEADER
	//	}
	//	// 	return modules.UNIT_STATE_CHECK_HEADER_PASSED
	//}
	//var isValidAssetId bool
	//for _, asset := range header.AssetIDs {
	//	if asset == header.Number.AssetID {
	//		isValidAssetId = true
	//		break
	//	}
	//}
	//if !isValidAssetId {
	//	fmt.Println(7)
	//	return modules.UNIT_STATE_INVALID_HEADER
	//}

	// check authors
	//TODO must recover
	//if header.Authors.Empty() {
	//	return modules.UNIT_STATE_INVALID_AUTHOR_SIGNATURE
	//}

	// comment by Albert·Gou 新生产的unit暂时还没有群签名
	//if len(header.GroupSign) < 64 {
	//	return modules.UNIT_STATE_INVALID_HEADER_WITNESS
	//}

	// TODO 同步过来的unit 没有Authors ，因此无法验证签名有效性。
	var thisUnitIsNotTransmitted bool

	if thisUnitIsNotTransmitted {
		sigState := validate.validateUnitSignature(header, isGenesis)
		return sigState
	}
	return TxValidationCode_VALID
}
func (validate *Validate) ValidateHeader(h *modules.Header) error {
	return nil
}
