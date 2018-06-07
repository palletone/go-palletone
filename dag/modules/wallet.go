package modules

import (
	"time"
)

// key: wallet.hash(wallet+create_date)
type Wallet struct {
	Wallet             string                 `json:"wallet"`
	Address            string                 `json:"address"`
	Amount             int                    `json:"amount"`
	DefinitionTemplate map[string]interface{} `json:"definition_template"`
	CreationDate       time.Time              `json:"creation_date"`
	FullApprovalDate   time.Time              `json:"full_approval_date"`
	ReadyDate          time.Time              `json:"ready_date"`
}
