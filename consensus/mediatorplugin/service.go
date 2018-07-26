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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package mediatorplugin

import (
	"fmt"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/ptn"
	"github.com/palletone/go-palletone/cmd/utils"
)

func (mp *MediatorPlugin) Protocols() []p2p.Protocol {
	return nil
}

func (mp *MediatorPlugin) APIs() []rpc.API {
	return nil
}

func (mp *MediatorPlugin) Start(server *p2p.Server) error {
	log.Info("mediator plugin startup begin")

	// 1. 判断是否满足生产验证单元的条件，主要判断本节点是否控制至少一个mediator账户
	if len(mp.mediators) == 0 {
		println("No mediators configured! Please add mediator and private keys to configuration.")
	} else {
		// 2. 开启循环生产计划
		log.Info(fmt.Sprintf("Launching verified unit production for %d mediators.", len(mp.mediators)))

		//if mp.ProductionEnabled {
		dag := mp.ptn.Dag()
		if dag.DynGlobalProp.LastVerifiedUnitNum == 0 {
			newChainBanner(dag)
		}
		//}

		go mp.ScheduleProductionLoop()
	}

	log.Info("mediator plugin startup end")

	return nil
}

func (mp *MediatorPlugin) Stop() error {
	return nil
}

// 匿名函数的好处之一：能在匿名函数内部直接使用本函数之外的变量;
// 函数使用外部变量的特性称之为闭包； 例如，以下匿名方法就直接使用cfg变量
func RegisterMediatorPluginService(stack *node.Node, cfg *Config) {
	log.Info("Register Mediator Plugin Service...")

	stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return Initialize(stack, cfg)
	})
}

func Initialize(node *node.Node, cfg *Config) (*MediatorPlugin, error) {
	log.Info("mediator plugin initialize begin")

	mss := cfg.Mediators
	msm := map[common.Address]string{}
	for address, passphrase := range mss {
		address := strings.TrimSpace(address)

		addr := common.StringToAddress(address)
		addrType, err := addr.Validate()
		if err != nil || addrType != common.PublicKeyHash {
			//			utils.Fatalf("Invalid mediator account address: %v", address)
			log.Info(fmt.Sprintf("Invalid mediator account address: %v", address))
		}

		log.Info(fmt.Sprintf("this node controll mediator account address: %v", address))

		msm[addr] = passphrase
	}

	var ptn *ptn.PalletOne
	if err := node.Service(&ptn); err != nil {
		utils.Fatalf("PalletOne service not running: %v", err)
	}

	mp := MediatorPlugin{
		ptn:              ptn,
		productionEnabled: cfg.EnableStaleProduction,
		mediators:         msm,
	}

	log.Info("mediator plugin initialize end")

	return &mp, nil
}
