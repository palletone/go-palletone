package modules

import (
	"time"
)

// 个人账户
// key: useraccount.hash(accountid+timestamp)
type Account struct {
	AccountId     string    `json:"account_id"` //帐户id
	Alias         string    `json:"alias"`      // 帐户名
	Signer        string    `json:"signer"`     //私钥对
	Utxos         []string  `json:"utxos"`      // 所有未花费的utxo的索引:key
	Status        int       `json:"status"`
	CreationDate  time.Time `json:"creation_date"`
	DestroyedDate time.Time `json:"destroyed_date"`
}

// 合约账户
// key: contractaccount.hash(contractid+timestamp)
type ContractAccount struct {
	ContractId    string    `json:"contract_id"`
	Scope         string    `json:"scope"` // account
	Code          string    `json:"code"`  // the account's operation
	Table         string    `json:"table"` // where is being Stored
	Key           string    `json:"key"`   // index in db
	Status        int       `json:"status"`
	CreationDate  time.Time `json:"creation_date"`
	DestroyedDate time.Time `json:"destroyed_date"`
}
