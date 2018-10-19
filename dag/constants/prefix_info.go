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
	UNIT_PREFIX                 = []byte("ut") // unit_prefix + mci + hash
	HEADER_PREFIX               = []byte("uh") // prefix + hash
	HeaderCanon_Prefix          = []byte("ch") // Canon Header Prefix
	UNIT_HASH_NUMBER_Prefix     = []byte("hn")
	UNIT_NUMBER_PREFIX          = []byte("nh") // number 和unit hash 的对应关系
	BODY_PREFIX                 = []byte("ub")
	TRANSACTION_PREFIX          = []byte("tx")
	Transaction_Index           = []byte("ti")
	TRANSACTIONS_PREFIX         = []byte("ts")
	AddrTransactionsHash_Prefix = []byte("at") // addr  transactions hash prefix
	AddrOutput_Prefix           = []byte("ao") // addr output tx's hash + msg index.
	AddrOutPoint_Prefix         = []byte("ap") // addr outpoint
	CONTRACT_STATE_PREFIX       = []byte("cs")
	CONTRACT_TPL                = []byte("ct")

	MESSAGES_PREFIX               = []byte("me")
	POLL_PREFIX                   = []byte("po")
	VOTE_PREFIX                   = []byte("vo")
	ATTESTATION_PREFIX            = []byte("at")
	ASSET_PREFIX                  = []byte("as")
	ASSET_ATTESTORS               = []byte("ae")
	MEDIATOR_INFO_PREFIX          = []byte("mi")
	GLOBALPROPERTY_PREFIX         = []byte("gp")
	DYNAMIC_GLOBALPROPERTY_PREFIX = []byte("dp")
	MEDIATOR_SCHEME_PREFIX        = []byte("ms")
	ACCOUNT_INFO_PREFIX           = []byte("ai")
	CONF_PREFIX                   = []byte("cf")
	// lookup
	LookupPrefix = []byte("l")

	// Head Fast Key
	HeadHeaderKey = []byte("LastHeader")
	HeadUnitKey   = []byte("LastUnit")
	HeadFastKey   = []byte("LastFast")
	TrieSyncKey   = []byte("TrieSync")

	// contract
	CONTRACT_PREFIX = []byte("cs")

	// other prefix
	EAENED_HEADERS_COMMISSION = "earned_headers_commossion"
	ALL_UNITS                 = "array_units"
	UTXOSNAPSHOT_PREFIX       = "us"

	// utxo && state storage
	CONTRACT_ATTRI    = []byte("contract") // like contract_[contract address]_[key]
	UTXO_PREFIX       = []byte("uo")
	UTXO_INDEX_PREFIX = []byte("ui")
	ASSET_INFO_PREFIX = []byte("ai")

	// token info
	TOKENTYPE  = []byte("tp") // tp[types]
	TOKENINFOS = []byte("tokeninfos")

	STATE_VOTE_LIST = []byte("MediatorVoteList")
)

// suffix
var (
	NumberSuffix = []byte("n")
)
