package common

import (
	"testing"
	"log"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestNewGenesisUnit(t *testing.T) {
	//ParentUnits []common.Hash `json:"parent_units"`
	//AssetIDs    []IDType36    `json:"assets"`
	//Authors     *Author      `json:"authors"` // the unit creation authors
	//Witness     []Author      `json:"witness"`
	//GasLimit    uint64        `json:"gasLimit"`
	//GasUsed     uint64        `json:"gasUsed"`
	//Root        common.Hash   `json:"root"`
	//Number      ChainIndex    `json:"index"`
	//Extra       []byte        `json:"extra"`
	gUnit, err := NewGenesisUnit(&core.Genesis{}, modules.Transactions{})
	if err!=nil {
		log.Println(err)
	} else {
		log.Println("Genesis unit struct:")
		log.Println("--- Genesis unit header --- ")
		log.Println("parent units:", gUnit.UnitHeader.ParentUnits)
		log.Println("asset ids:", gUnit.UnitHeader.AssetIDs)
		log.Println("witness:", gUnit.UnitHeader.Witness)
		log.Println("gaslimit:", gUnit.UnitHeader.GasLimit)
		log.Println("gasUsed:", gUnit.UnitHeader.GasUsed)
		log.Println("Root:", gUnit.UnitHeader.Root)
		log.Println("Number:", gUnit.UnitHeader.Number)
	}
}
