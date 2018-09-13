package comm

import (
	"github.com/palletone/go-palletone/dag"
)

var useDag dag.IDag

func SetCcDagHand(dag dag.IDag) error {
	useDag = dag
	return nil
}

func GetCcDagHand() (dag.IDag, error) {
	return useDag, nil
}
