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

package adaptor

//ISmartContract 智能合约的相关操作接口
type ISmartContract interface {
	IUtility
	//创建一个安装合约的交易，未签名
	CreateContractInstallTx(input *CreateContractInstallTxInput) (*CreateContractInstallTxOutput, error)
	//查询合约安装的结果的交易
	GetContractInstallTx(input *GetContractInstallTxInput) (*GetContractInstallTxOutput, error)
	//初始化合约实例
	CreateContractInitialTx(input *CreateContractInitialTxInput) (*CreateContractInitialTxOutput, error)
	//查询初始化合约实例的交易
	GetContractInitialTx(input *GetContractInitialTxInput) (*GetContractInitialTxOutput, error)
	//调用合约方法
	CreateContractInvokeTx(input *CreateContractInvokeTxInput) (*CreateContractInvokeTxOutput, error)
	//查询调用合约方法的交易
	GetContractInvokeTx(input *GetContractInvokeTxInput) (*GetContractInvokeTxOutput, error)
	//调用合约的查询方法
	QueryContract(input *QueryContractInput) (*QueryContractOutput, error)
	//销毁合约
	// CreateContractDestroyTx(input *CreateContractDestroyTx) (tx []byte, err error)
	// GetContractDestroyTxByTxId(txId []byte) (*ContractDestroyTx, error)
}
type CreateContractInstallTxInput struct {
	Address  string       `json:"address"`
	Fee      *AmountAsset `json:"fee"`
	Contract []byte       `json:"contract"`
	Extra    []byte       `json:"extra"`
}
type CreateContractInstallTxOutput struct {
	RawTransaction []byte `json:"raw_transaction"`
	Extra          []byte `json:"extra"`
}
type CreateContractInitialTxInput struct {
	Address  string       `json:"address"`
	Fee      *AmountAsset `json:"fee"`
	Contract []byte       `json:"contract"`
	Args     [][]byte     `json:"args"`
	Extra    []byte       `json:"extra"`
}
type CreateContractInitialTxOutput struct {
	RawTransaction []byte `json:"raw_transaction"`
	Extra          []byte `json:"extra"`
}
type CreateContractInvokeTxInput struct {
	Address         string       `json:"address"`
	Fee             *AmountAsset `json:"fee"`
	ContractAddress string       `json:"contract_address"`
	Function        string       `json:"function"`
	Args            [][]byte     `json:"args"`
	Extra           []byte       `json:"extra"`
}
type CreateContractInvokeTxOutput struct {
	RawTransaction []byte `json:"raw_transaction"`
	Extra          []byte `json:"extra"`
}
type QueryContractInput struct {
	Address         string       `json:"address"`
	Fee             *AmountAsset `json:"fee"`
	ContractAddress string       `json:"contract_address"`
	Function        string       `json:"function"`
	Args            [][]byte     `json:"args"`
	Extra           []byte       `json:"extra"`
}
type QueryContractOutput struct {
	QueryResult []byte `json:"query_result"`
	Extra       []byte `json:"extra"`
}

type GetContractInstallTxInput struct {
	TxID  []byte `json:"tx_id"`
	Extra []byte `json:"extra"`
}
type GetContractInstallTxOutput struct {
	TxBasicInfo
	TemplateID []byte `json:"template_id"`
}
type GetContractInitialTxInput struct {
	TxID  []byte `json:"tx_id"`
	Extra []byte `json:"extra"`
}
type GetContractInitialTxOutput struct {
	TxBasicInfo
	ContractAddress string `json:"contract_address"`
	Extra           []byte `json:"extra"`
}
type GetContractInvokeTxInput struct {
	TxID  []byte `json:"tx_id"`
	Extra []byte `json:"extra"`
}
type GetContractInvokeTxOutput struct {
	TxBasicInfo
	UpdateStateSuccess bool   `json:"update_state_success"` //读写集一致，成功更新StateDB
	InvokeResult       []byte `json:"invoke_result"`
	Extra              []byte `json:"extra"`
}
