package storage

import (
	"log"

	palletdb "github.com/palletone/go-palletone/common/ptndb"
)

var (
	PACKET_PREFIX                    = []byte("p") // packet_prefix  + mci + hash
	UNIT_PREFIX                      = []byte("u") // unit_prefix + mci + hash
	HEADERPREFIX                     = []byte("h") // prefix + hash
	TRANSACTIONPREFIX                = []byte("t")
	TRANSACTIONSPREFIX               = []byte("ts")
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

	EAENED_HEADERS_COMMISSION = "earned_headers_commossion"
	ALL_UNITS                 = "array_units"

	// state storage
	CONTRACT_ATTRI = []byte("contract_") // like contract_[contract address]_[key]

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
	log.Println("renew dbconn start...")
	if dbconn, err := palletdb.NewLDBDatabase(path, 0, 0); err != nil {
		log.Println("renew dbconn error:", err)
		return nil
	} else {
		return dbconn
	}
}
