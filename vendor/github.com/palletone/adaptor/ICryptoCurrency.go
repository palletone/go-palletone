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

//ICryptoCurrency 加密货币相关的API
type ICryptoCurrency interface {
	IUtility
	//获取某地址下持有某资产的数量,返回数量为该资产的最小单位
	GetBalance(input *GetBalanceInput) (*GetBalanceOutput, error)
	//获取某资产的小数点位数
	GetAssetDecimal(asset *GetAssetDecimalInput) (*GetAssetDecimalOutput, error)
	//创建一个转账交易，但是未签名
	CreateTransferTokenTx(input *CreateTransferTokenTxInput) (*CreateTransferTokenTxOutput, error)
	//获取某个地址对某种Token的交易历史,支持分页和升序降序排列
	GetAddrTxHistory(input *GetAddrTxHistoryInput) (*GetAddrTxHistoryOutput, error)
	//根据交易ID获得对应的转账交易
	GetTransferTx(input *GetTransferTxInput) (*GetTransferTxOutput, error)
	//创建一个多签地址，该地址必须要满足signCount个签名才能解锁
	CreateMultiSigAddress(input *CreateMultiSigAddressInput) (*CreateMultiSigAddressOutput, error)
}

//查询余额时的输入参数
type GetBalanceInput struct {
	Address string `json:"address"`
	Asset   string `json:"asset"`
	Extra   []byte `json:"extra"`
}

//查询余额的返回值
type GetBalanceOutput struct {
	Balance AmountAsset `json:"balance"`
	Extra   []byte      `json:"extra"`
}

//获得某种资产小数位数时的输入
type GetAssetDecimalInput struct {
	Asset string `json:"asset"` //资产标识
	Extra []byte `json:"extra"`
}

//获得的资产的小数位数
type GetAssetDecimalOutput struct {
	Decimal uint   `json:"decimal"`
	Extra   []byte `json:"extra"`
}
type CreateTransferTokenTxInput struct {
	FromAddress string       `json:"from_address"`
	ToAddress   string       `json:"to_address"`
	Amount      *AmountAsset `json:"amount"`
	Fee         *AmountAsset `json:"fee"`
	Extra       []byte       `json:"extra"`
}
type CreateTransferTokenTxOutput struct {
	Transaction []byte `json:"transaction"`
	Extra       []byte `json:"extra"`
}
type GetAddrTxHistoryInput struct {
	FromAddress       string `json:"from_address"`         //转账的付款方地址
	ToAddress         string `json:"to_address"`           //转账的收款方地址
	Asset             string `json:"asset"`                //资产标识
	PageSize          uint32 `json:"page_size"`            //分页大小，0表示不分页
	PageIndex         uint32 `json:"page_index"`           //分页后的第几页数据
	AddressLogicAndOr bool   `json:"address_logic_and_or"` //付款地址,收款地址是And=1关系还是Or=0关系
	Asc               bool   `json:"asc"`                  //按时间顺序从老到新
	Extra             []byte `json:"extra"`
}
type GetAddrTxHistoryOutput struct {
	Txs   []*SimpleTransferTokenTx `json:"transactions"` //返回的交易列表
	Count uint32                   `json:"count"`        //忽略分页，有多少条记录
	Extra []byte                   `json:"extra"`
}
type CreateMultiSigAddressInput struct {
	Keys      [][]byte `json:"keys"`
	SignCount int      `json:"sign_count"`
	Extra     []byte   `json:"extra"`
}
type CreateMultiSigAddressOutput struct {
	Address string `json:"address"`
	Extra   []byte `json:"extra"`
}
type GetTransferTxInput struct {
	TxID  []byte `json:"tx_id"`
	Extra []byte `json:"extra"`
}
type GetTransferTxOutput struct {
	Tx    SimpleTransferTokenTx `json:"transaction"`
	Extra []byte                `json:"extra"`
}
