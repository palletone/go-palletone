/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package contracts

import (
	"errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	cc "github.com/palletone/go-palletone/contracts/manger"

	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/dag"
	md "github.com/palletone/go-palletone/dag/modules"
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

type ContractInf interface {
	Close() error
	Install(chainID string, ccName string, ccPath string, ccVersion string) (payload *md.ContractTplPayload, err error)
	Deploy(chainID string, templateId []byte, txId string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *md.ContractDeployPayload, e error)
	//Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*md.ContractInvokePayload, error)
	Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*md.ContractInvokeResult, error)
	//Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*modules.ContractInvokeResult, error)
	Stop(chainID string, deployId []byte, txid string, deleteImage bool) (*md.ContractStopPayload, error)
}

//var (
//	//保证金合约地址（0x1号合约）
//	DepositeContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000011C")
//)

// Initialize 初始化合约管理模块以及加载系统合约，
// 由上层应用指定dag以及初始合约配置信息
// Initialize the contract management module and load the system contract,
// Specify dag and initial contract configuration information by the upper application
func Initialize(idag dag.IDag, jury core.IAdapterJury, cfg *contractcfg.Config) (*Contract, error) {
	atomic.LoadInt32(&initFlag)
	if initFlag > 0 {
		//todo  tmp delete
		//return nil, errors.New("contract already init")
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
	if err := cc.Init(idag, jury); err != nil {
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
func (c *Contract) Install(chainID string, ccName string, ccPath string, ccVersion string) (payload *md.ContractTplPayload, err error) {
	log.Info("===========================enter contract.go Install==============================")
	defer log.Info("===========================exit contract.go Install==============================")
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		log.Error("initFlag == 0")
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
func (c *Contract) Deploy(chainID string, templateId []byte, txId string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *md.ContractDeployPayload, e error) {
	log.Info("===========================enter contract.go Deploy==============================")
	defer log.Info("===========================exit contract.go Deploy==============================")
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		log.Error("initFlag == 0")
		return nil, nil, errors.New("Contract not initialized")
	}
	return cc.Deploy(c.dag, chainID, templateId, txId, args, timeout)
}

// Invoke 合约invoke调用，根据指定合约调用参数执行已经部署的合约，函数返回合约调用单元。
// The contract invoke call, execute the deployed contract according to the specified contract call parameters,
// and the function returns the contract call unit.
func (c *Contract) Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*md.ContractInvokeResult, error) {
	log.Info("===========================enter contract.go Invoke==============================")
	defer log.Info("===========================exit contract.go Invoke==============================")
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		log.Error("initFlag == 0")
		return nil, errors.New("contract not initialized")
	}
	//depId := common.NewAddress(deployId[:20], common.ContractHash)
	return cc.Invoke(c.dag, chainID, deployId, txid, args, timeout)
}

// Stop 停止指定合约。根据需求可以对镜像文件进行删除操作
//Stop the specified contract. The image file can be deleted according to requirements.
func (c *Contract) Stop(chainID string, deployId []byte, txid string, deleteImage bool) (*md.ContractStopPayload, error) {
	log.Info("===========================enter contract.go Stop==============================")
	defer log.Info("===========================exit contract.go Stop==============================")
	atomic.LoadInt32(&initFlag)
	if initFlag == 0 {
		log.Error("initFlag == 0")
		return nil, errors.New("contract not initialized")
	}
	return cc.Stop(deployId, chainID, deployId, txid, deleteImage)
}
