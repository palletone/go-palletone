package constants

import (
	"log"
)

var (
	GENESIS_UNIT               string
	VERSION                    string
	ALT                        string
	COUNT_WITNESSES            int
	MAX_WITNESS_LIST_MUTATIONS int
	// anti-spam limits
	MAX_AUTHORS_PER_UNIT                   int = 16
	MAX_PARENTS_PER_UNIT                   int = 16
	MAX_MESSAGES_PER_UNIT                  int = 128
	MAX_SPEND_PROOFS_PER_MESSAGE           int = 128
	MAX_INPUTS_PER_PAYMENT_MESSAGE         int = 128
	MAX_OUTPUTS_PER_PAYMENT_MESSAGE        int = 128
	MAX_CHOICES_PER_POLL                   int = 128
	MAX_DENOMINATIONS_PER_ASSET_DEFINITION int = 64
	MAX_ATTESTORS_PER_ASSET                int = 64
	MAX_DATA_FEED_NAME_LENGTH              int = 64
	MAX_DATA_FEED_VALUE_LENGTH             int = 64
	MAX_AUTHENTIFIER_LENGTH                int = 4096
	MAX_CAP                                int = 9e15
	MAX_COMPLEXITY                         int = 100

	MAX_PROFILE_FIELD_LENGTH int = 50
	MAX_PROFILE_VALUE_LENGTH int = 100

	TEXTCOIN_CLAIM_FEE                int = 548
	TEXTCOIN_ASSET_CLAIM_FEE          int = 750
	TEXTCOIN_ASSET_CLAIM_HEADER_FEE   int = 391
	TEXTCOIN_ASSET_CLAIM_MESSAGE_FEE  int = 209
	TEXTCOIN_ASSET_CLAIM_BASE_MSG_FEE int = 158
)

func init() {
	VERSION = "1.0"
	ALT = "1"
	log.Println("start constant init...")
	if VERSION == "1.0" && ALT == "1" {
		GENESIS_UNIT = "TvqutGPz3T4Cs6oiChxFlclY92M2MvCvfXR5/FETato="

	} else {
		GENESIS_UNIT = "oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E="

	}
}
