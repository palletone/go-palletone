// Unspent Transaction Output module.
package modules

// key: utxo.hash(utxo+timestamp)
type Utxo struct {
	AccountId string `json:"account_id"` // 所属人id
	Amount    int    `json:"amount"`     // 数量
	AssertId  string `json:"assert_id"`  // 资产类别
	Alias     string `json:"alias"`      // 资产别名
	Program   string `json:"program"`    // 要执行的代码段
	Key       string `json:"key"`        // 索引值
}

type Input struct {
	Unit               string `json:"unit"`
	MessageIndex       int    `json:"message_index"`
	InputIndex         int    `json:"input_index"`
	Asset              string `json:"asset"`
	Denomination       int    `json:"denomination"` // default 1
	IsUnique           int    `json:"is_unique"`    //default 1
	TypeEnum           string `json:"type_unum"`    //'transfer','headers_commission','witnessing','issue'
	SrcUnit            string `json:"src_unit"`
	SrcMessageIndex    int    `json:"src_message_index"`
	SrcOutputIndex     int    `json:"src_output_index"`
	FromMainChainIndex int64  `json:"from_main_chain_index"`
	ToMainChainIndex   int64  `json:"to_main_chain_index"`
	SerialNumber       int64  `json:"serial_number"`
	Amount             int64  `json:"amount"`
	Address            string `json:"address"`
}

type Output struct {
	OutputId     int    `json:"output_id"`
	Unit         string `json:"unit"`
	MessageIndex int    `json:"message_index"`
	OutputIndex  int    `json:"output_index"`
	Asset        string `json:"asset"`
	Denomination int    `json:"denomination"` // default 1
	Amount       int64  `json:"amount"`
	Address      string `json:"address"`
	Blinding     string `json:"blinding"`
	OutputHash   string `json:"output_hash"`
	IsSerial     int    `json:"is_serial"` // default 0 if not stable yet
	IsSpent      int    `json:"is_spend"`  //  default 0
}

type SpendProof struct {
	Unit string `json:"unit"`
}
