package common

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/asset"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestUpdateUtxo(t *testing.T) {
	UpdateUtxo(common.Hash{}, &modules.Message{}, 0, 0)
	dagconfig.DefaultConfig.DbPath = getTempDir(t)
}

func TestReadUtxos(t *testing.T) {
	dagconfig.DefaultConfig.DbPath = getTempDir(t)
	utxos, totalAmount := ReadUtxos(common.Address{}, modules.Asset{})
	log.Println(utxos, totalAmount)
}

func TestGetUxto(t *testing.T) {
	dagconfig.DefaultConfig.DbPath = getTempDir(t)
	log.Println(modules.Input{})
}

func getTempDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "leveldb-test")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestSaveAssetInfo(t *testing.T) {
	assetid := asset.NewAsset()
	asset := modules.Asset{
		AssertId: assetid,
		UniqueId: assetid,
		ChainId:  0,
	}
	assetInfo := modules.AssetInfo{
		Alias:        "Test",
		AssetID:      asset,
		InitialTotal: 1000000000,
		Decimal:      100000000,
	}
	assetInfo.OriginalHolder.SetString("Mytest")
}

func TestWalletBalance(t *testing.T) {
	addr := common.Address{}
	addr.SetString("P1CXn936dYuPKGyweKPZRycGNcwmTnqeDaA")
	balance := WalletBalance(addr, modules.Asset{})
	log.Println("Address total =", balance)
}

func TestGetAccountTokens(t *testing.T) {
	addr := common.Address{}
	addr.SetString("P14yBqQ4ogJ3RdEFoVTdd5ucsdu3PffZZRu")
	tokens, err := GetAccountTokens(addr)
	if err != nil {
		log.Println("Get account error:", err.Error())
	} else if len(tokens) == 0 {
		log.Println("Get none account")
	} else {
		for _, token := range tokens {
			log.Printf("Token (%s, %v) = %v\n",
				token.Alias, token.AssetID.AssertId, token.Balance)
			// test WalletBalance method
			log.Println(WalletBalance(addr, token.AssetID))
			// test ReadUtxos method
			utxos, amount := ReadUtxos(addr, token.AssetID)
			log.Printf("Addr(%s) balance=%v\n", addr.String(), amount)
			for outpoint, utxo := range utxos {
				log.Println(">>> UTXO txhash =", outpoint.TxHash.String())
				log.Println("    UTXO msg index =", outpoint.MessageIndex)
				log.Println("    UTXO out index =", outpoint.OutIndex)
				log.Println("    UTXO amount =", utxo.Amount)
			}
		}
	}

}
