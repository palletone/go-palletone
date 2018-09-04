package contracts

import (
	unit "github.com/palletone/go-palletone/dag/modules"
	cc "github.com/palletone/go-palletone/contracts/manger"
	"time"
	"github.com/palletone/go-palletone/dag"
)


type Contract struct {

}

func (c *Contract) Start(dag *dag.Dag) {
	go cc.Init(dag)
}

func (c *Contract) Install(chainID string, ccName string, ccPath string, ccVersion string) (payload *unit.ContractTplPayload, err error) {
	return cc.Install(chainID, ccName, ccPath, ccVersion)
}

func(c *Contract) Deploy(chainID string, templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *unit.ContractDeployPayload, e error)  {
	return cc.Deploy(chainID, templateId, txid, args, timeout)
}

func(c *Contract)Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*unit.ContractInvokePayload, error) {
	return cc.Invoke(chainID, deployId, txid, args, timeout)
}

func(c *Contract)Stop(chainID string, deployId []byte, txid string, deleteImage bool) error {
	return cc.Stop(chainID, deployId, txid, deleteImage)
}


