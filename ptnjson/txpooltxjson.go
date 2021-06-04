package ptnjson

import (
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/txspool"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

type TxPoolTxJson struct {
	TxHash     string       `json:"tx_hash"`
	Version    uint32       `json:"version"`
	Nonce      uint64       `json:"nonce"`
	UnitHash   string       `json:"unit_hash"`
	UnitIndex  uint64       `json:"unit_index"`
	Timestamp  uint64       `json:"timestamp"`
	TxIndex    uint64       `json:"tx_index"`
	Payment    *PaymentJson `json:"payment"`
	TxMessages string       `json:"tx_messages"`

	Froms        []*OutPointJson `json:"froms"`
	CreationDate time.Time       `json:"creation_date"`
	Priority     float64         `json:"priority"` // 打包的优先级
	Pending      bool            `json:"pending"`
	Confirmed    bool            `json:"confirmed"`
	IsOrphan     bool            `json:"is_orphan"`
	NotExsit     bool            `json:"not_exist"`
	Index        int             `json:"index"` // index 是该tx在优先级堆中的位置
	Extra        []byte          `json:"extra"`
	TxHex        string          `json:"tx_hex"`
}
type TxPoolPendingJson struct {
	TxHash       string    `json:"tx_hash"`
	Fee          uint64    `json:"fee"`
	Asset        string    `json:"asset"`
	CreationDate time.Time `json:"creation_date"`
	Amount       uint64    `json:"amount"`
}
type TxSerachEntryJson struct {
	UnitHash  string `json:"unit_hash"`
	AssetId   string `json:"asset_id"`
	UnitIndex uint64 `json:"unit_index"`
	TxIndex   uint64 `json:"tx_index"`
}

func ConvertTxPoolTx2PendingJson(tx *txspool.TxPoolTransaction) *TxPoolPendingJson {
	if tx == nil {
		return nil
	}
	if tx.Tx == nil {
		return nil
	}
	// pay := new(modules.PaymentPayload)
	var amount uint64
	if len(tx.Tx.Messages()) > 0 {
		pay := tx.Tx.TxMessages()[0].Payload.(*modules.PaymentPayload)

		for _, out := range pay.Outputs {
			amount += out.Value
		}
	}
	txfee := tx.GetTxFee().Uint64()
	var asset_str string
	if tx.TxFee != nil {
		asset_str = tx.TxFee[0].Asset.String()
	}
	return &TxPoolPendingJson{
		TxHash:       tx.Tx.Hash().String(),
		CreationDate: tx.CreationDate,
		Fee:          txfee,
		Asset:        asset_str,
		Amount:       amount,
	}
}

func ConvertTxPoolTx2Json(tx *txspool.TxPoolTransaction, hash common.Hash) *TxPoolTxJson {
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
	msgs := tx.Tx.TxMessages()
	if len(msgs) > 0 {
		pay = msgs[0].Payload.(*modules.PaymentPayload)

		for _, out := range tx.From {
			froms = append(froms, ConvertOutPoint2Json(out))
		}
	}
	txBytes, _ := rlp.EncodeToBytes(tx.Tx)
	payJson := ConvertPayment2Json(pay)
	return &TxPoolTxJson{
		TxHash:     tx.Tx.Hash().String(),
		Version:    tx.Tx.Version(),
		Nonce:      tx.Tx.Nonce(),
		UnitHash:   hex_hash,
		Payment:    payJson,
		TxMessages: ConvertMegs2Json(msgs),

		Froms:        froms,
		CreationDate: tx.CreationDate,
		Priority:     tx.GetPriorityfloat64(),
		Pending:      tx.Pending,
		Confirmed:    tx.Confirmed,
		IsOrphan:     tx.IsOrphan,
		Extra:        tx.Extra[:],
		TxHex:        hex.EncodeToString(txBytes),
	}
}

func ConvertTxEntry2Json(entry *modules.TxLookupEntry) *TxSerachEntryJson {
	return &TxSerachEntryJson{
		UnitHash:  entry.UnitHash.String(),
		UnitIndex: entry.UnitIndex,
		TxIndex:   entry.Index,
	}
}

func ConvertTxWithInfo2Json(tx *modules.TransactionWithUnitInfo) *TxPoolTxJson {
	pay := new(modules.PaymentPayload)
	msgs := tx.TxMessages()
	if len(msgs) > 0 {
		pay = msgs[0].Payload.(*modules.PaymentPayload)
	}

	payJson := ConvertPayment2Json(pay)
	return &TxPoolTxJson{
		TxHash:     tx.Hash().String(),
		UnitHash:   tx.UnitHash.String(),
		UnitIndex:  tx.UnitIndex,
		Timestamp:  tx.Timestamp,
		TxIndex:    tx.TxIndex,
		Payment:    payJson,
		TxMessages: ConvertMegs2Json(msgs),

		Pending:   true,
		Confirmed: true,
		IsOrphan:  false,
		NotExsit:  false,
	}
}
