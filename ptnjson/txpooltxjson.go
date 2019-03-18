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
	Pending      bool            `json:"pending"`
	Confirmed    bool            `json:"confirmed"`
	IsOrphan     bool            `json:"is_orphan"`
	Index        int             `json:"index"` // index 是该tx在优先级堆中的位置
	Extra        []byte          `json:"extra"`
}
type TxSerachEntryJson struct {
	UnitHash  string `json:"unit_hash"`
	AssetId   string `json:"asset_id"`
	UnitIndex uint64 `json:"unit_index"`
	TxIndex   uint64 `json:"tx_index"`
}

func ConvertTxPoolTx2Json(tx *modules.TxPoolTransaction, hash common.Hash) *TxPoolTxJson {
	if tx == nil {
		return nil
	}
	if tx.Tx == nil {
		return nil
	}
	var hex_hash string
	if hash != (common.Hash{}) {
		hex_hash = hash.String()
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
		UnitHash:   hex_hash,
		Payment:    &payJson,
		TxMessages: ConvertMegs2Json(tx.Tx.TxMessages),

		Froms:        froms,
		CreationDate: tx.CreationDate,
		Priority:     tx.GetPriorityfloat64(),
		Pending:      tx.Pending,
		Confirmed:    tx.Confirmed,
		IsOrphan:     tx.IsOrphan,
		Extra:        tx.Extra[:],
	}
}

func ConvertTxEntry2Json(entry *modules.TxLookupEntry) *TxSerachEntryJson {
	return &TxSerachEntryJson{
		UnitHash:  entry.UnitHash.String(),
		UnitIndex: entry.UnitIndex,
		TxIndex:   entry.Index,
	}
}
