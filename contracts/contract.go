package contracts

import (
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	cc "github.com/palletone/go-palletone/contracts/manger"
	"github.com/palletone/go-palletone/dag"
	unit "github.com/palletone/go-palletone/dag/modules"
	"time"
)

type Contract struct {
	//cfg *contractcfg.Config
	name string
	dag  dag.IDag
}

func Initialize(cfg *contractcfg.Config) (*Contract, error) {
	var contractCfg contractcfg.Config
	if cfg == nil {
		contractCfg = contractcfg.DefaultConfig
	} else {
		contractCfg = *cfg
	}

	cc := &Contract{name: "name"}
	contractcfg.SetConfig(&contractCfg)

	log.Debug("contract initialize ok")
	return cc, nil
}

func (c *Contract) Start(idag dag.IDag) {
	//c.dag = dag
	if c.dag == nil {
		c.dag = idag
	}
	go cc.Init(idag)
}

func (c *Contract) Install(chainID string, ccName string, ccPath string, ccVersion string) (payload *unit.ContractTplPayload, err error) {
	return cc.Install(c.dag, chainID, ccName, ccPath, ccVersion)
}

func (c *Contract) Deploy(chainID string, templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *unit.ContractDeployPayload, e error) {
	return cc.Deploy(c.dag, chainID, templateId, txid, args, timeout)
}

func (c *Contract) Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*unit.ContractInvokePayload, error) {
	return cc.Invoke(c.dag, chainID, deployId, txid, args, timeout)
}

func (c *Contract) Stop(chainID string, deployId []byte, txid string, deleteImage bool) error {
	return cc.Stop(chainID, deployId, txid, deleteImage)
}
