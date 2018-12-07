package ptnjson

import (
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type TxPoolTxJson struct {
	TxHash     string       `json:"txhash"`
	UnitHash   string       `json:"unithash"`
	Payment    *PaymentJson `json:"payment"`
	TxMessages string       `json:"txmessages"`

	Froms        []*OutPointJson `json:"froms"`
	CreationDate time.Time       `json:"creation_date"`
	Priority     float64         `json:"priority"` // 打包的优先级
	Nonce        uint64          `json:"nonce"`
	Pending      bool            `json:"pending"`
	Confirmed    bool            `json:"confirmed"`
	Index        int             `json:"index"` // index 是该tx在优先级堆中的位置
	Extra        []byte          `json:"extra"`
}

func ConvertTxPoolTx2Json(tx *modules.TxPoolTransaction, hash common.Hash) *TxPoolTxJson {
	if tx == nil {
		return nil
	}
	if tx.Tx == nil {
		return nil
	}
	froms := make([]*OutPointJson, 0)
	pay := new(modules.PaymentPayload)
	if len(tx.Tx.TxMessages) > 0 {
		pay = tx.Tx.TxMessages[0].Payload.(*modules.PaymentPayload)

		for _, out := range tx.From {
			froms = append(froms, ConvertOutPoint2Json(out))
		}
	}

	payJson := ConvertPayment2Json(pay)
	return &TxPoolTxJson{
		TxHash:     tx.Tx.Hash().String(),
		UnitHash:   hash.String(),
		Payment:    &payJson,
		TxMessages: ConvertMegs2Json(tx.Tx.TxMessages),

		Froms:        froms,
		CreationDate: tx.CreationDate,
		Priority:     tx.Priority_lvl,
		Nonce:        tx.Nonce,
		Pending:      tx.Pending,
		Confirmed:    tx.Confirmed,
		Extra:        tx.Extra[:],
	}
}
