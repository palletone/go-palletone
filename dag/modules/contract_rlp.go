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
package modules

//
//import (
//	"github.com/palletone/go-palletone/common"
//	"math/big"
//	"time"
//)
//
//type ContractTemp struct {
//	//Contract Id 即Address，[20]byte，
//	// 根据用户创建合约实例申请的RequestId截取其后20字节生成
//	Id       []byte
//	Name     string
//	ConType  string // 合约类型： 系统合约 用户合约
//	LangCode string // 代码类别
//
//	Sign              []*Authentifier // 单一签名和多方签名
//	Code              []byte          // 合约代码。
//	CodeHash          common.Hash
//	CodeAddress       common.Address
//	Input             []byte
//	JuryPubKeys       [][]byte //该合约对于的陪审员公钥列表
//	NeedApprovalCount uint8    //需要多少个陪审员同意才算共识达成
//	CallerAddress     common.Address
//	caller            common.Address
//	self              common.Address // 合約地址
//	jumpdests         map[common.Hash][]byte
//
//	value *big.Int
//
//	Args []byte
//
//	status uint32 // 合约状态
//	// creator
//	creation uint32 // creation  date
//}
//
//func ContractToTemp(contract *Contract) *ContractTemp {
//	c := ContractTemp{}
//	c.Id = contract.Id
//	c.Name = contract.Name
//	c.ConType = contract.ConType
//	c.LangCode = contract.LangCode
//	c.Sign = contract.Sign
//	c.Code = contract.Code
//	c.CodeHash = contract.CodeHash
//	c.CodeAddress = contract.CodeAddress
//	c.Input = contract.Input
//	c.JuryPubKeys = contract.JuryPubKeys
//	c.NeedApprovalCount = contract.NeedApprovalCount
//	c.CallerAddress = contract.CallerAddress
//	c.caller = contract.caller
//	c.self = contract.self
//	c.jumpdests = contract.jumpdests
//	c.value = contract.value
//	c.Args = contract.Args
//	c.status = uint32(contract.status)
//	c.creation = uint32(contract.creation.Unix())
//	return &c
//}
//func ContractTempToContract(contractTemp *ContractTemp) *Contract {
//	c := Contract{}
//	c.Id = contractTemp.Id
//	c.Name = contractTemp.Name
//	c.ConType = contractTemp.ConType
//	c.LangCode = contractTemp.LangCode
//	c.Sign = contractTemp.Sign
//	c.Code = contractTemp.Code
//	c.CodeHash = contractTemp.CodeHash
//	c.CodeAddress = contractTemp.CodeAddress
//	c.Input = contractTemp.Input
//	c.JuryPubKeys = contractTemp.JuryPubKeys
//	c.NeedApprovalCount = contractTemp.NeedApprovalCount
//	c.CallerAddress = contractTemp.CallerAddress
//	c.caller = contractTemp.caller
//	c.self = contractTemp.self
//	c.jumpdests = contractTemp.jumpdests
//	c.value = contractTemp.value
//	c.Args = contractTemp.Args
//	c.status = int(contractTemp.status)
//	timestamp := int64(contractTemp.creation)
//	c.creation = time.Unix(timestamp, 0)
//	return &c
//}
