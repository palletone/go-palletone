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

import "github.com/palletone/go-palletone/dag/errors"

type ValidationCode byte

const (
	TxValidationCode_VALID                        ValidationCode = 0
	TxValidationCode_INVALID_CONTRACT_TEMPLATE    ValidationCode = 1
	TxValidationCode_INVALID_FEE                  ValidationCode = 2
	TxValidationCode_BAD_COMMON_HEADER            ValidationCode = 3
	TxValidationCode_BAD_CREATOR_SIGNATURE        ValidationCode = 4
	TxValidationCode_INVALID_ENDORSER_TRANSACTION ValidationCode = 5
	TxValidationCode_INVALID_CONFIG_TRANSACTION   ValidationCode = 6
	TxValidationCode_UNSUPPORTED_TX_PAYLOAD       ValidationCode = 7
	TxValidationCode_BAD_PROPOSAL_TXID            ValidationCode = 8
	TxValidationCode_DUPLICATE_TXID               ValidationCode = 9
	TxValidationCode_ENDORSEMENT_POLICY_FAILURE   ValidationCode = 10
	TxValidationCode_MVCC_READ_CONFLICT           ValidationCode = 11
	TxValidationCode_PHANTOM_READ_CONFLICT        ValidationCode = 12
	TxValidationCode_UNKNOWN_TX_TYPE              ValidationCode = 13
	TxValidationCode_STATE_DATA_NOT_FOUND         ValidationCode = 14
	TxValidationCode_MARSHAL_TX_ERROR             ValidationCode = 15
	TxValidationCode_NIL_TXACTION                 ValidationCode = 16
	TxValidationCode_EXPIRED_CHAINCODE            ValidationCode = 17
	TxValidationCode_CHAINCODE_VERSION_CONFLICT   ValidationCode = 18
	TxValidationCode_BAD_HEADER_EXTENSION         ValidationCode = 19
	TxValidationCode_BAD_CHANNEL_HEADER           ValidationCode = 20
	TxValidationCode_BAD_RESPONSE_PAYLOAD         ValidationCode = 21
	TxValidationCode_BAD_RWSET                    ValidationCode = 22
	TxValidationCode_ILLEGAL_WRITESET             ValidationCode = 23
	TxValidationCode_INVALID_WRITESET             ValidationCode = 24
	TxValidationCode_INVALID_MSG                  ValidationCode = 25
	TxValidationCode_INVALID_PAYMMENTLOAD         ValidationCode = 26
	TxValidationCode_INVALID_PAYMMENT_INPUT       ValidationCode = 27
	TxValidationCode_INVALID_PAYMMENT_INPUT_COUNT ValidationCode = 28
	TxValidationCode_INVALID_COINBASE             ValidationCode = 29
	TxValidationCode_ADDRESS_IN_BLACKLIST             ValidationCode = 30
	TxValidationCode_INVALID_AMOUNT               ValidationCode = 31
	TxValidationCode_INVALID_ASSET                ValidationCode = 32
	TxValidationCode_INVALID_CONTRACT             ValidationCode = 33
	TxValidationCode_INVALID_DATAPAYLOAD          ValidationCode = 34
	TxValidationCode_INVALID_DOUBLE_SPEND         ValidationCode = 35
	TxValidationCode_INVALID_TOKEN_STATUS         ValidationCode = 36
	TxValidationCode_NOT_COMPARE_SIZE             ValidationCode = 37
	TxValidationCode_ORPHAN                       ValidationCode = 255

	TxValidationCode_INVALID_OTHER_REASON         ValidationCode = 251
	TxValidationCode_NOT_VALIDATED        ValidationCode = 250
	UNIT_STATE_AUTHOR_SIGNATURE_PASSED   ValidationCode = 101
	UNIT_STATE_INVALID_MEDIATOR_SCHEDULE ValidationCode = 102
	UNIT_STATE_INVALID_AUTHOR_SIGNATURE  ValidationCode = 103
	UNIT_STATE_INVALID_GROUP_SIGNATURE   ValidationCode = 104
	UNIT_STATE_HAS_INVALID_TRANSACTIONS  ValidationCode = 105
	UNIT_STATE_INVALID_AUTHOR              ValidationCode = 106
	UNIT_STATE_INVALID_EXTRA_DATA        ValidationCode = 107
	UNIT_STATE_INVALID_HEADER            ValidationCode = 108
	UNIT_STATE_INVALID_HEADER_NUMBER     ValidationCode = 109
	UNIT_STATE_INVALID_HEADER_TXROOT     ValidationCode = 110
	UNIT_STATE_INVALID_HEADER_TIME       ValidationCode = 111
	UNIT_STATE_ORPHAN                    ValidationCode = 254
)

var validationCode_name = map[byte]string{
	0:   "VALID",
	1:   "INVALID_CONTRACT_TEMPLATE",
	2:   "INVALID_FEE",
	3:   "BAD_COMMON_HEADER",
	4:   "BAD_CREATOR_SIGNATURE",
	5:   "INVALID_ENDORSER_TRANSACTION",
	6:   "INVALID_CONFIG_TRANSACTION",
	7:   "UNSUPPORTED_TX_PAYLOAD",
	8:   "BAD_PROPOSAL_TXID",
	9:   "DUPLICATE_TXID",
	10:  "ENDORSEMENT_POLICY_FAILURE",
	11:  "MVCC_READ_CONFLICT",
	12:  "PHANTOM_READ_CONFLICT",
	13:  "UNKNOWN_TX_TYPE",
	14:  "STATE_DATA_NOT_FOUND",
	15:  "MARSHAL_TX_ERROR",
	16:  "NIL_TXACTION",
	17:  "EXPIRED_CHAINCODE",
	18:  "CHAINCODE_VERSION_CONFLICT",
	19:  "BAD_HEADER_EXTENSION",
	20:  "BAD_CHANNEL_HEADER",
	21:  "BAD_RESPONSE_PAYLOAD",
	22:  "BAD_RWSET",
	23:  "ILLEGAL_WRITESET",
	24:  "INVALID_WRITESET",
	25:  "INVALID_MSG",
	26:  "INVALID_PAYMMENTLOAD",
	27:  "INVALID_PAYMMENT_INPUT",
	28:  "INVALID_PAYMMENT_INPUT_COUNT",
	29:  "INVALID_PAYMMENT_COINBASE",
	30:  "ADDRESS_IN_BLACKLIST",
	31:  "INVALID_AMOUNT",
	32:  "INVALID_ASSET",
	33:  "INVALID_CONTRACT",
	34:  "INVALID_DATAPAYLOAD",
	35:  "DOUBLE_SPEND",
	36:  "INVALID_TOKEN_STATUS",
	37:  "NOT_COMPARE_SIZE",
	101: "AUTHOR_SIGNATURE_PASSED",
	102: "UNIT_STATE_INVALID_MEDIATOR_SCHEDULE",
	103: "INVALID_AUTHOR_SIGNATURE",
	104: "INVALID_GROUP_SIGNATURE",
	105: "HAS_INVALID_TRANSACTIONS",
	106: "INVALID_AUTHOR",
	107: "INVALID_EXTRA_DATA",
	108: "INVALID_HEADER",
	109: "CHECK_HEADER_PASSED",
	110: "UNIT_STATE_INVALID_HEADER_TXROOT",
	111: "INVALID_HEADER_TIME",
	125: "OTHER_ERROR",

	251: "NOT_VALIDATED",

	250: "INVALID_OTHER_REASON",
	255: "ORPHAN TX",
	254: "ORPHAN UNIT",
}

func NewValidateError(code ValidationCode) error {
	if code == TxValidationCode_VALID {
		return nil
	}
	return errors.New(validationCode_name[byte(code)])
}
func IsOrphanError(err error) bool {
	return err.Error() == "ORPHAN TX" || err.Error() == "ORPHAN UNIT"
}
