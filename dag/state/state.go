package state

import (
	"time"
	"github.com/palletone/go-palletone/common"
)

// contract template
type ContractTplObj struct {
	TplID    	string			`json:"address"`			// contract template address
	CodeHash 	  	string            		`json:"codeHash"`			// contract code hash
	Code          	[]byte            		`json:"code"`				// contract bytecode
	CreationDate  	time.Time 				`json:"creation_date"`		// contract template create time
}

// instance of contract
type ContractObj struct {
	TplID    		string					`json:"tpl_id"`			// contract template address
	Address 		common.Address			`json:"address"`			// the contract instance address
	Params        	map[string]interface{}  `json:"params"`				// contract params status
	Status		  	string					`json:"status"`				// the contract status, like 'good', 'bad', 'close' etc.
	CreationDate  	time.Time 				`json:"creation_date"`		// contract template create time
	DestroyedDate 	time.Time 				`json:"destroyed_date"`
}


type ContractAccount struct {
	Address 	common.Address				`json:"address"`			// user accounts address
	Contracts 	map[string]ContractObj		`json:"contracts"`			// user contracts
}
