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

package storage

import (
	"log"

	palletdb "github.com/palletone/go-palletone/common/ptndb"
)

var (
	UNIT_PREFIX                      = []byte("ut") // unit_prefix + mci + hash
	HEADER_PREFIX                    = []byte("uh") // prefix + hash
	UNIT_HASH_NUMBER_Prefix          = []byte("hn")
	UNIT_NUMBER_PREFIX               = []byte("nh") // number 和unit hash 的对应关系
	BODY_PREFIX                      = []byte("ub")
	TRANSACTION_PREFIX               = []byte("tx")
	Transaction_Index                = []byte("ti")
	TRANSACTIONS_PREFIX              = []byte("ts")
	CONTRACT_STATE_PREFIX            = []byte("cs")
	CONTRACT_TPL                     = []byte("ct")
	ALL_UNITS_PREFIX                 = []byte("au")
	UNITAUTHORS_PREFIX               = []byte("ua")
	HASH_TREE_BALLS_PREFIX           = []byte("ht")
	UNIT_WITNESS_PREFIX              = []byte("uw")
	WITNESS_LIST_HASHES_PREFIX       = []byte("wl")
	DEFINITIONS_PREFIX               = []byte("de")
	ADDRESS_PREFIX                   = []byte("ad")
	ADDRESS_DEFINITION_CHANGE_PREFIX = []byte("ac")
	AUTHENTIFIERS_PREFIX             = []byte("au")
	MESSAGES_PREFIX                  = []byte("me")
	POLL_PREFIX                      = []byte("po")
	VOTE_PREFIX                      = []byte("vo")
	ATTESTATION_PREFIX               = []byte("at")
	ASSET_PREFIX                     = []byte("as")
	ASSET_ATTESTORS                  = []byte("ae")
	NumberSuffix                     = []byte("n")
	// lookup
	LookupPrefix = []byte("l")

	// Head Fast Key
	HeadHeaderKey = []byte("LastHeader")
	HeadUnitKey   = []byte("LastUnit")
	HeadFastKey   = []byte("LastFast")
	TrieSyncKey   = []byte("TrieSync")

	// contract
	CONTRACT_PTEFIX = []byte("cs")

	// other prefix
	EAENED_HEADERS_COMMISSION = "earned_headers_commossion"
	ALL_UNITS                 = "array_units"
)

func Init(path string) *palletdb.LDBDatabase {
	var err error
	if Dbconn == nil {
		if path == "" {
			path = DBPath
		}
		Dbconn, err = palletdb.NewLDBDatabase(path, 0, 0)
		if err != nil {
			log.Println("new dbconn error:", err)
		}
		log.Println("db_path:", Dbconn.Path())
	}
	return Dbconn
}
func ReNewDbConn(path string) *palletdb.LDBDatabase {
	if dbconn, err := palletdb.NewLDBDatabase(path, 0, 0); err != nil {
		log.Println("renew dbconn error:", path, err)
		return nil
	} else {
		return dbconn
	}
}
