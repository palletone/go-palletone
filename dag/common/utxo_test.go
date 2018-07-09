package common

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"io/ioutil"
	"log"
	"testing"
)

func TestReadUtxos(t *testing.T) {

	dagconfig.DefaultConfig.DbPath = getTempDir(t)
	utxos, totalAmount := ReadUtxos(common.Address{}, modules.Asset{})
	log.Println(utxos, totalAmount)
}

func TestGetUxto(t *testing.T) {
	dagconfig.DefaultConfig.DbPath = getTempDir(t)
	log.Println(modules.Input{})
}

func TestUpdateUtxo(t *testing.T) {
	dagconfig.DefaultConfig.DbPath = getTempDir(t)
	UpdateUtxo(common.Address{}, modules.Transaction{})
}

func getTempDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "leveldb-test")
	if err != nil {
		t.Fatal(err)
	}
	return d
}
