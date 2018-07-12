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
	UpdateUtxo(&modules.Transaction{})
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
	balance := WalletBalance(common.Address{}, modules.Asset{})
	log.Println("Address total =", balance)
}
