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
var (
	UNIT_PREFIX             = []byte("ut")  // unit_prefix + mci + hash
	HEADER_PREFIX           = []byte("uh")  // prefix + hash
	HEADER_HEIGTH_PREFIX    = []byte("uht") // prefix + height:hash
	HeaderCanon_Prefix      = []byte("ch")  // Canon Header Prefix
	UNIT_HASH_NUMBER_Prefix = []byte("hn")
	//UNIT_NUMBER_PREFIX          = []byte("nh") // number 和unit hash 的对应关系
	BODY_PREFIX                 = []byte("ub")
	TRANSACTION_PREFIX          = []byte("tx")
	Transaction_Index           = []byte("ti")
	TRANSACTIONS_PREFIX         = []byte("ts")
	AddrTransactionsHash_Prefix = []byte("at")  // to addr  transactions hash prefix
	AddrTx_From_Prefix          = []byte("fat") // from addr transactions hash prefix
	AddrOutput_Prefix           = []byte("ao")  // addr output tx's hash + msg index.
	AddrOutPoint_Prefix         = []byte("ap")  // addr outpoint
	OutPointAddr_Prefix         = []byte("pa")  // outpoint addr
	CONTRACT_STATE_PREFIX       = []byte("cs")
	CONTRACT_TPL                = []byte("ct")
	CONTRACT_TPL_CODE           = []byte("code")
	CONTRACT_DEPLOY             = []byte("cdy")
	CONTRACT_DEPLOY_REQ         = []byte("cdr")
	CONTRACT_STOP               = []byte("csp")
	CONTRACT_STOP_REQ           = []byte("csr")
	CONTRACT_INVOKE             = []byte("civ")
	CONTRACT_INVOKE_REQ         = []byte("ciq")
	CONTRACT_SIGNATURE          = []byte("csn")

	MESSAGES_PREFIX        = []byte("me")
	POLL_PREFIX            = []byte("po")
	CREATE_VOTE_PREFIX     = []byte("vo")
	ATTESTATION_PREFIX     = []byte("at")
	ASSET_PREFIX           = []byte("as")
	ASSET_ATTESTORS        = []byte("ae")
	MEDIATOR_INFO_PREFIX   = []byte("mi")
	DEPOSIT_BALANCE_PREFIX = []byte("db")

	GLOBALPROPERTY_HISTORY_PREFIX = []byte("gh")

	ACCOUNT_INFO_PREFIX        = []byte("ai")
	ACCOUNT_PTN_BALANCE_PREFIX = []byte("ab")
	TokenTxHash_Prefix         = []byte("tt")
	// lookup
	LookupPrefix = []byte("l")

	LastUnitInfo = []byte("stbu")
	// LastStableUnitHash   = []byte("stbu")
	// LastUnstableUnitHash = []byte("ustbu")
	HeadUnitHash               = []byte("HeadUnitHash")
	HeadHeaderKey              = []byte("LastHeader")
	HeadFastKey                = []byte("LastFast")
	TrieSyncKey                = []byte("TrieSync")
	GenesisUnitHash            = []byte("GenesisUnitHash")
	GLOBALPROPERTY_KEY         = []byte("gpGlobalProperty")
	DYNAMIC_GLOBALPROPERTY_KEY = []byte("dpDynamicGlobalProperty")
	MEDIATOR_SCHEDULE_KEY      = []byte("msMediatorSchedule")
	// contract
	CONTRACT_PREFIX           = []byte("cp")
	CONTRACT_TPL_INSTANCE_MAP = []byte("cmap")
	CONTRACT_JURY_PREFIX      = []byte("jury")

	// other prefix
	EAENED_HEADERS_COMMISSION = "earned_headers_commossion"
	ALL_UNITS                 = "array_units"
	UTXOSNAPSHOT_PREFIX       = "us"

	// utxo && state storage
	UTXO_PREFIX       = []byte("uo")
	UTXO_INDEX_PREFIX = []byte("ui")
	//ASSET_INFO_PREFIX = []byte("pi") // ACCOUNT_INFO_PREFIX is also "ai"  asset=property

	// token info
	TOKENTYPE  = []byte("tp") // tp[types]
	TOKENINFOS = []byte("tokeninfos")
	// state current chain index
	CURRENTCHAININDEX_PREFIX = "ccix"

	STATE_VOTER_LIST = []byte("vl")

	// ReqId && TxHash maping
	ReqIdPrefix      = []byte("req")
	TxHash2ReqPrefix = []byte("tx2req")

	//filehash
	IDX_FileHash_Txid   = []byte("mda")
	RewardAddressPrefix = "Addr:"
)

// suffix
var (
	NumberSuffix = []byte("n")
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
