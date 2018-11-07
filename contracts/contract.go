package contracts

import (
	"errors"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	cc "github.com/palletone/go-palletone/contracts/manger"
	"github.com/palletone/go-palletone/dag"
	unit "github.com/palletone/go-palletone/dag/modules"
	"sync/atomic"
	"time"
)

var initFlag int32

type Contract struct {
	cfg  *contractcfg.Config
	name string
	dag  dag.IDag
	//status int32 //   1:init   2:start
}

var (
	//保证金合约地址（0x1号合约）
	DepositeContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000011C")
)

// Initialize 初始化合约管理模块以及加载系统合约，
// 由上层应用指定dag以及初始合约配置信息
// Initialize the contract management module and load the system contract,
// Specify dag and initial contract configuration information by the upper application
func Initialize(idag dag.IDag, cfg *contractcfg.Config) (*Contract, error) {
	atomic.LoadInt32(&initFlag)
	if initFlag > 0 {
		return nil, errors.New("contract already init")
	}

	var contractCfg contractcfg.Config
	if cfg == nil {
		contractCfg = contractcfg.DefaultConfig
	} else {
		contractCfg = *cfg
	}
	contract := &Contract{
		name: "palletone",
		dag:  idag,
		cfg:  cfg,
	}
	contractcfg.SetConfig(&contractCfg)
	if err := cc.Init(idag); err != nil {
		return nil, err
	}

	atomic.StoreInt32(&initFlag, 1)
	log.Debug("contract initialize ok")
	return contract, nil
}

func (c *Contract) Close() error {
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		return errors.New("contract already deInit")
	}
	cc.Deinit()
	atomic.StoreInt32(&initFlag, 0)
	return nil
}

// Install 合约安装，将指定的合约路径文件打包，并与合约名称、版本一起构成合约模板单元
// chainID 链码ID，用于多链
// Contract installation, packaging the specified contract path file,
// and forming a contract template unit together with the contract name and version
// Chain code ID for multiple chains
func (c *Contract) Install(chainID string, ccName string, ccPath string, ccVersion string) (payload *unit.ContractTplPayload, err error) {
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		return nil, errors.New("Contract not initialized")
	}
	return cc.Install(c.dag, chainID, ccName, ccPath, ccVersion)
}

// Deploy 将指定的合约模板部署到本地，生成对应Docker镜像及启动带有初始化合约参数的容器，用于合约的执行。
// txid由上层应用指定，合约部署超时时间根据具体服务器配置指定，默认40秒。接口返回合约部署ID（每次部署其返回ID不同）以及部署单元
// Deploy the specified contract template locally,
// generate the corresponding Docker image and launch a container with initialization contract parameters for contract execution.
// The txid is specified by the upper application.
// The contract deployment timeout is specified according to the configuration of server.The default is 40 seconds.
// The interface returns the contract deployment ID (there is a different return ID for each deployment)
// and the deployment unit
func (c *Contract) Deploy(chainID string, templateId []byte, txid string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *unit.ContractDeployPayload, e error) {
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		return nil, nil, errors.New("Contract not initialized")
	}
	return cc.Deploy(c.dag, chainID, templateId, txid, args, timeout)
}

// Invoke 合约invoke调用，根据指定合约调用参数执行已经部署的合约，函数返回合约调用单元。
// The contract invoke call, execute the deployed contract according to the specified contract call parameters,
// and the function returns the contract call unit.
func (c *Contract) Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*unit.ContractInvokePayload, error) {
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		return nil, errors.New("Contract not initialized")
	}
	return cc.Invoke(deployId, c.dag, chainID, deployId, txid, args, timeout)
}

// Stop 停止指定合约。根据需求可以对镜像文件进行删除操作
//Stop the specified contract. The image file can be deleted according to requirements.
func (c *Contract) Stop(chainID string, deployId []byte, txid string, deleteImage bool) error {
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		return errors.New("Contract not initialized")
	}
	return cc.Stop(deployId, chainID, deployId, txid, deleteImage)
}
