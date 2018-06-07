package modules

import (
	"time"
)

type transaction struct {
	TransactionId     string                 `json:"transaction_id"`
	TranType          string                 `json:"tran_type"`
	TranHash          string                 `json:"transaction_hash"`
	TranReceiptStatus string                 `json:"tran_receipt_status"`
	DagMainChainIndex int64                  `json:"dag_main_chain_index"`
	CreationDate      time.Time              `json:"creation_date"`
	From              string                 `json:"from"`
	To                string                 `json:"to"`
	Value             int64                  `json:"value"`
	InputData         map[string]interface{} `json:"input_data"`
}
