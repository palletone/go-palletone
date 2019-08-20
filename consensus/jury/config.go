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
package jury

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/palletone/go-palletone/common"
)

const (
	DefaultElectionNum      = 2          //todo
	DefaultContractSigNum   = 2          //todo
	MaxLengthTplName        = 64         //合约模板名字长度
	MaxLengthTplPath        = 512        //合约模板文件路径长度
	MaxLengthTplVersion     = 12         //合约模板版本号长度
	MaxNumberTplEleAddrHash = 5          //合约模板指定节点地址hash数量
	MaxLengthTplId          = 128        //合约模板Id长度
	MaxNumberArgs           = 32         //合约请求参数数量
	MaxLengthArgs           = 1024       //合约请求输入参数长度
	MaxLengthExtData        = 16         //合约请求扩展数据长度
	MaxLengthAbi            = 1024 * 500 //合约Abi数据长度
	MaxLengthLanguage       = 32         //合约模板语言类型长度
	MaxLengthDescription    = 1024       //合约描述数据长度
)

var (
	AccountInfoFlag = cli.BoolFlag{
		Name:  "accountInfo",
		Usage: "The About information account address,password,public and private key and on ",
	}
)

type AccountConf struct {
	Address,
	Password string
}
type JuryAccount struct {
	Address  common.Address
	Password string
}
type Config struct {
	//ContractSigNum int   //user contract jury sig number  //todo  no used
	//ElectionNum    int   //vrf election jury number       //todo  no used
	Accounts []*AccountConf // the set of the mediator info
}

func (aConf *AccountConf) configToAccount() *JuryAccount {
	addr, _ := common.StringToAddress(aConf.Address)
	if addr != (common.Address{}) {
		medAcc := &JuryAccount{
			addr,
			aConf.Password,
		}
		return medAcc
	}
	return nil
}

var DefaultConfig = Config{
	Accounts: []*AccountConf{
		&AccountConf{},
	},
}

func MakeConfig() Config {
	cfg := DefaultConfig
	cfg.Accounts = nil
	return cfg
}

func SetJuryConfig(ctx *cli.Context, cfg *Config) {
	switch {
	case ctx.GlobalIsSet(AccountInfoFlag.Name):
	}
}
