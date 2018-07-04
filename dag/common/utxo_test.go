package common

import (
	"testing"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
	"log"
)

func TestReadUtxos(t *testing.T) {
	utxos, totalAmount := ReadUtxos(common.Address{}, modules.Asset{})
	log.Println(utxos, totalAmount)
}

func TestGetUxto(t *testing.T)  {
	log.Println(modules.Input{})
}

func TestUpdateUtxo(t *testing.T)  {
	UpdateUtxo(common.Address{}, modules.Transaction{})
}