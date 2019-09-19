/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package constants

// prefix info
//各个Entity的Prefix应该都是2小写字母，不可重复
var (
	HEADER_PREFIX               = []byte("uh") // prefix + hash
	HEADER_HEIGTH_PREFIX        = []byte("hh") // prefix + height:hash
	UNIT_HASH_NUMBER_PREFIX     = []byte("hn")
	BODY_PREFIX                 = []byte("ub")
	TRANSACTION_PREFIX          = []byte("tx")
	ADDR_TXID_PREFIX            = []byte("at") // to addr  transactions hash prefix
	ADDR_OUTPOINT_PREFIX        = []byte("ap") // addr outpoint
	OUTPOINT_ADDR_PREFIX        = []byte("pa") // outpoint addr
	CONTRACT_STATE_PREFIX       = []byte("cs")
	CONTRACT_TPL                = []byte("ct")
	CONTRACT_TPL_CODE           = []byte("cc")
	CONTRACT_DEPLOY             = []byte("cd")
	CONTRACT_DEPLOY_REQ         = []byte("ce")
	CONTRACT_STOP               = []byte("cp")
	CONTRACT_STOP_REQ           = []byte("cq")
	CONTRACT_INVOKE             = []byte("ci")
	CONTRACT_INVOKE_REQ         = []byte("ck")
	CONTRACT_SIGNATURE          = []byte("cn")
	CONTRACT_PREFIX             = []byte("co")
	CONTRACT_TPL_INSTANCE_MAP   = []byte("cm")
	CONTRACT_JURY_PREFIX        = []byte("cj")
	REQID_TXID_PREFIX           = []byte("rq")
	MEDIATOR_INFO_PREFIX        = []byte("mi")
	DEPOSIT_BALANCE_PREFIX      = []byte("db")
	DEPOSIT_JURY_BALANCE_PREFIX = []byte("djbp")
	//DEPOSIT_MEDIATOR_VOTE_PREFIX = []byte("dn")
	PLEDGE_DEPOSIT_PREFIX  = []byte("pd")
	PLEDGE_WITHDRAW_PREFIX = []byte("pw")

	GLOBAL_PROPERTY_HISTORY_PREFIX = []byte("gh")

	ACCOUNT_INFO_PREFIX        = []byte("ai")
	ACCOUNT_PTN_BALANCE_PREFIX = []byte("ab")
	TOKEN_TXID_PREFIX          = []byte("tt") //IndexDB中存储一个Token关联的TxId
	TOKEN_EX_PREFIX            = []byte("te") //IndexDB中存储一个Token关联的ProofOfExistence
	// lookup
	LOOKUP_PREFIX              = []byte("lu")
	UTXO_PREFIX                = []byte("uo")
	SPENT_UTXO_PREFIX          = []byte("us")
	UTXO_INDEX_PREFIX          = []byte("ui")
	TrieSyncKey                = []byte("TrieSync")
	LastUnitInfo               = []byte("stbu")
	GenesisUnitHash            = []byte("GenesisUnitHash")
	GLOBALPROPERTY_KEY         = []byte("gpGlobalProperty")
	DYNAMIC_GLOBALPROPERTY_KEY = []byte("dpDynamicGlobalProperty")
	MEDIATOR_SCHEDULE_KEY      = []byte("msMediatorSchedule")
	DATA_VERSION_KEY           = []byte("gptnversion")

	//filehash
	IDX_MAIN_DATA_TXID              = []byte("md") //Old value: mda
	IDX_REF_DATA_PREFIX             = []byte("re")
	RewardAddressPrefix             = "Addr:"
	JURY_PROPERTY_USER_CONTRACT_KEY = []byte("jpuck")
)

// symbols
var (
	CERT_SPLIT_CH = string("||")
	// certificate
	CERT_ISSUER_SYMBOL  = "certissuer_"
	CERT_SERVER_SYMBOL  = "certserver_"
	CERT_MEMBER_SYMBOL  = "certmember_"
	CERT_BYTES_SYMBOL   = "certbytes_"
	CERT_SUBJECT_SYMBOL = "certsubject_"
	CRL_BYTES_SYMBOL    = "crlbytes_"
)
