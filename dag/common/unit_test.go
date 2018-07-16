package common

import (
	"log"
	"testing"

	"fmt"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestNewGenesisUnit(t *testing.T) {
	gUnit, _ := NewGenesisUnit(modules.Transactions{})

	log.Println("Genesis unit struct:")
	log.Println("--- Genesis unit header --- ")
	log.Println("parent units:", gUnit.UnitHeader.ParentUnits)
	log.Println("asset ids:", gUnit.UnitHeader.AssetIDs)
	log.Println("witness:", gUnit.UnitHeader.Witness)
	log.Println("Root:", gUnit.UnitHeader.Root)
	log.Println("Number:", gUnit.UnitHeader.Number)

}

func TestGenGenesisConfigPayload(t *testing.T) {
	var genesisConf core.Genesis
	genesisConf.SystemConfig.DepositRate = 0.02
	genesisConf.SystemConfig.MediatorInterval = 10

	payload, err := GenGenesisConfigPayload(&genesisConf)

	if err != nil {
		log.Println(err)
	}

	for k, v := range payload.ConfigSet {
		log.Println(k, v)
	}
}

func TestSaveUnit(t *testing.T) {
	if err := SaveUnit(modules.Unit{}); err != nil {
		log.Println(err)
	}
}

type TestByte string

func TestRlpDecode(t *testing.T) {
	var t1, t2, t3 TestByte
	t1 = "111"
	t2 = "222"
	t3 = "333"

	bytes := []TestByte{t1, t2, t3}
	encodeBytes, _ := rlp.EncodeToBytes(bytes)
	var data []TestByte
	rlp.DecodeBytes(encodeBytes, &data)
	fmt.Printf("%q", data)
}
