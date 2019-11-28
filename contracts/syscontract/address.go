/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package syscontract

import "github.com/palletone/go-palletone/common"

var (
	//1保证金合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM
	DepositContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000011C")

	//2创币合约PRC20
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
	CreateTokenContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000021C")

	//3投票合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRLGbeyd
	VoteTokenContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000031C")

	//4系统参数维护合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRS71ZEM
	SysConfigContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000041C")

	//5产块奖励合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRUp5qmM
	CoinbaseContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000051C")
	//6 黑名单合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF
	BlacklistContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000061C")

	//7创币合约PRC721
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRijspoq
	CreateToken721ContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000071C")

	//8数字身份管理合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk
	DigitalIdentityContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000081C")

	//9分区管理合约
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DRxVdGDZ
	PartitionContractAddress = common.HexToAddress("0x00000000000000000000000000000000000000091C")

	//15测试调试用
	//PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf
	TestContractAddress = common.HexToAddress("0x000000000000000000000000000000000000000F1C")
	//PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX
	TestRunContractAddress = common.HexToAddress("0x00000000000000000000000000000000000095271C")
)
