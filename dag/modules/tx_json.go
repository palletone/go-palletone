package modules

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
)

var _ = (*txdataMarshaling)(nil)

func (t txdata) MarshalJSON() ([]byte, error) {
	type txdata struct {
		From      *common.Address `json:"from"    gencodec:"required"`
		Price     *hexutil.Big    `json:"gasPrice" gencodec:"required"`
		GasLimit  hexutil.Uint64  `json:"gas"      gencodec:"required"`
		Recipient *common.Address `json:"to"       rlp:"nil"`
		Amount    *hexutil.Big    `json:"value"    gencodec:"required"`
		Payload   hexutil.Bytes   `json:"input"    gencodec:"required"`
		V         *hexutil.Big    `json:"v" gencodec:"required"`
		R         *hexutil.Big    `json:"r" gencodec:"required"`
		S         *hexutil.Big    `json:"s" gencodec:"required"`
		Hash      *common.Hash    `json:"hash" rlp:"-"`
	}
	var enc txdata
	enc.From = t.From
	enc.Price = (*hexutil.Big)(t.Price)
	enc.Recipient = t.Recipient
	enc.Amount = (*hexutil.Big)(t.Amount)
	enc.Payload = t.Payload
	enc.V = (*hexutil.Big)(t.V)
	enc.R = (*hexutil.Big)(t.R)
	enc.S = (*hexutil.Big)(t.S)
	enc.Hash = t.Hash
	return json.Marshal(&enc)
}

func (t *txdata) UnmarshalJSON(input []byte) error {
	type txdata struct {
		From      *common.Address `json:"from"    gencodec:"required"`
		Price     *hexutil.Big    `json:"gasPrice" gencodec:"required"`
		GasLimit  *hexutil.Uint64 `json:"gas"      gencodec:"required"`
		Recipient *common.Address `json:"to"       rlp:"nil"`
		Amount    *hexutil.Big    `json:"value"    gencodec:"required"`
		Payload   *hexutil.Bytes  `json:"input"    gencodec:"required"`
		V         *hexutil.Big    `json:"v" gencodec:"required"`
		R         *hexutil.Big    `json:"r" gencodec:"required"`
		S         *hexutil.Big    `json:"s" gencodec:"required"`
		Hash      *common.Hash    `json:"hash" rlp:"-"`
	}
	var dec txdata
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	// if dec.From == nil {
	// 	return errors.New("missing required field 'from' for txdata")
	// }
	if dec.From != nil {
		t.From = dec.From
	}
	if dec.Price == nil {
		return errors.New("missing required field 'gasPrice' for txdata")
	}
	t.Price = (*big.Int)(dec.Price)
	if dec.GasLimit == nil {
		return errors.New("missing required field 'gas' for txdata")
	}

	if dec.Recipient != nil {
		t.Recipient = dec.Recipient
	}
	if dec.Amount == nil {
		return errors.New("missing required field 'value' for txdata")
	}
	t.Amount = (*big.Int)(dec.Amount)
	if dec.Payload == nil {
		return errors.New("missing required field 'input' for txdata")
	}
	t.Payload = *dec.Payload
	if dec.V == nil {
		return errors.New("missing required field 'v' for txdata")
	}
	t.V = (*big.Int)(dec.V)
	if dec.R == nil {
		return errors.New("missing required field 'r' for txdata")
	}
	t.R = (*big.Int)(dec.R)
	if dec.S == nil {
		return errors.New("missing required field 's' for txdata")
	}
	t.S = (*big.Int)(dec.S)
	if dec.Hash != nil {
		t.Hash = dec.Hash
	}
	return nil
}
