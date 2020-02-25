package comm

import (
	"github.com/palletone/go-palletone/dag/dboperation"
)

var useDag dboperation.IContractDag

func SetCcDagHand(dag dboperation.IContractDag) error {
	useDag = dag
	return nil
}

func GetCcDagHand() (dboperation.IContractDag, error) {
	return useDag, nil
}
