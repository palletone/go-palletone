package peer

type ContractInvokePayload struct {
	ContractId string                 `json:"contract_id"` // contract id
	Function   []byte                 `json:"function"`    // serialized value of invoked function with call parameters
	ReadSet    map[string]interface{} `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   map[string]interface{} `json:"write_set"`   // the set data of write, and value could be any type
}

type ContractDeployPayload struct {
	ContractId string                 `json:"contract_id"` // contract id
	Function   []byte                 `json:"function"`    // serialized value of invoked function with call parameters
	ReadSet    map[string]interface{} `json:"read_set"`    // the set data of read, and value could be any type
	WriteSet   map[string]interface{} `json:"write_set"`   // the set data of write, and value could be any type
}
