package comm

import (
	"github.com/palletone/go-palletone/dag"
)

var useDag *dag.Dag

func SetCcDagHand(dag *dag.Dag) error{
	useDag = dag
	return nil
}

func GetCcDagHand() (*dag.Dag, error){
	return useDag, nil
}
