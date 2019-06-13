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
 * @date 2018/11/05
 */

package ptnapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/contracts/syscontract/deposit"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/shopspring/decimal"
)

type PublicMediatorAPI struct {
	Backend
}

func NewPublicMediatorAPI(b Backend) *PublicMediatorAPI {
	return &PublicMediatorAPI{b}
}

func (a *PublicMediatorAPI) IsApproved(addStr string) (string, error) {
	// 构建参数
	cArgs := [][]byte{defaultMsg0, defaultMsg1, []byte(modules.IsApproved), []byte(addStr)}
	txid := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))

	// 调用系统合约
	rsp, err := a.ContractQuery(syscontract.DepositContractAddress.Bytes(), txid[:], cArgs, 0)
	if err != nil {
		return "", err
	}

	return string(rsp), nil
}

func (a *PublicMediatorAPI) GetDeposit(addStr string) (*deposit.MediatorDeposit, error) {
	// 构建参数
	cArgs := [][]byte{defaultMsg0, defaultMsg1, []byte(modules.GetMediatorDeposit), []byte(addStr)}
	txid := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().Unix())).Int31n(100000000))

	// 调用系统合约
	rsp, err := a.ContractQuery(syscontract.DepositContractAddress.Bytes(), txid[:], cArgs, 0)
	if err != nil {
		return nil, err
	}

	depositB := deposit.NewMediatorDeposit()
	err = json.Unmarshal(rsp, depositB)
	if err == nil {
		return depositB, nil
	}

	return nil, fmt.Errorf(string(rsp))
}

func (a *PublicMediatorAPI) IsInList(addStr string) (bool, error) {
	mediator, err := common.StringToAddress(addStr)
	if err != nil {
		return false, err
	}

	return a.Dag().IsMediator(mediator), nil
}

func (a *PublicMediatorAPI) ListAll() []string {
	addStrs := make([]string, 0)
	mas := a.Dag().GetMediators()

	for address, _ := range mas {
		addStrs = append(addStrs, address.Str())
	}

	return addStrs
}

func (a *PublicMediatorAPI) ListVoteResults() map[string]uint64 {
	mediatorVoteCount := make(map[string]uint64)

	for address, _ := range a.Dag().GetMediators() {
		mediatorVoteCount[address.String()] = 0
	}

	for med, stake := range a.Dag().MediatorVotedResults() {
		mediatorVoteCount[med] = stake
	}

	return mediatorVoteCount
}

func (a *PublicMediatorAPI) LookupMediatorInfo() []*modules.MediatorInfo {
	return a.Dag().LookupMediatorInfo()
}

func (a *PublicMediatorAPI) IsActive(addStr string) (bool, error) {
	mediator, err := common.StringToAddress(addStr)
	if err != nil {
		return false, err
	}

	return a.Dag().IsActiveMediator(mediator), nil
}

func (a *PublicMediatorAPI) ListActives() []string {
	addStrs := make([]string, 0)
	ms := a.Dag().GetActiveMediators()

	for _, medAdd := range ms {
		addStrs = append(addStrs, medAdd.Str())
	}

	return addStrs
}

func (a *PublicMediatorAPI) GetVoted(addStr string) ([]string, error) {
	addr, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, err
	}

	voted := a.Dag().GetAccountVotedMediators(addr)
	mediators := make([]string, 0, len(voted))

	for med, _ := range voted {
		mediators = append(mediators, med)
	}

	return mediators, nil
}

func (a *PublicMediatorAPI) GetNextUpdateTime() string {
	dgp := a.Dag().GetDynGlobalProp()
	time := time.Unix(int64(dgp.NextMaintenanceTime), 0)

	return time.Format("2006-01-02 15:04:05")
}

func (a *PublicMediatorAPI) GetInfo(addStr string) (*modules.MediatorInfo, error) {
	mediator, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, err
	}

	//if !a.Dag().IsMediator(mediator) {
	//	return nil, fmt.Errorf("%v is not mediator", mediator.Str())
	//}

	return a.Dag().GetMediatorInfo(mediator), nil
}

const DefaultResult = "Transaction executed locally, but may not be confirmed by the network yet!"

type PrivateMediatorAPI struct {
	Backend
}

func NewPrivateMediatorAPI(b Backend) *PrivateMediatorAPI {
	return &PrivateMediatorAPI{b}
}

// 交易执行结果
type TxExecuteResult struct {
	TxContent string      `json:"txContent"`
	TxHash    common.Hash `json:"txHash"`
	TxSize    string      `json:"txSize"`
	TxFee     string      `json:"txFee"`
	Tip       string      `json:"tip"`
	Warning   string      `json:"warning"`
}

// 创建 mediator 所需的参数, 至少包含普通账户地址
type MediatorCreateArgs struct {
	*modules.MediatorCreateOperation
}

// 相关参数检查
func (args *MediatorCreateArgs) setDefaults(addStr string) {
	if args.MediatorInfoBase == nil {
		args.MediatorInfoBase = core.NewMediatorInfoBase()
	}

	if args.AddStr == "" {
		args.AddStr = addStr
	}

	if args.InitPubKey == "" {
		args.InitPubKey = core.DefaultInitPubKey
	}

	if args.Node == "" {
		args.Node = core.DefaultNodeInfo
	}

	if args.MediatorApplyInfo == nil {
		args.MediatorApplyInfo = core.NewMediatorApplyInfo()
	}

	return
}

func (a *PrivateMediatorAPI) Apply(args MediatorCreateArgs) (*TxExecuteResult, error) {
	// 参数补全
	args.setDefaults("")

	// 参数验证
	err := args.Validate()
	if err != nil {
		return nil, err
	}

	addr := args.FeePayer()
	// 判断是否已经是mediator
	if a.Dag().IsMediator(addr) {
		return nil, fmt.Errorf("account %v is already a mediator", args.AddStr)
	}

	// 参数序列化
	argsB, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	cArgs := [][]byte{[]byte(modules.ApplyMediator), argsB}

	// 调用系统合约
	fee := a.Dag().CurrentFeeSchedule().MediatorCreateFee
	reqId, err := a.ContractInvokeReqTx(addr, addr, 0, fee, nil,
		syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Apply mediator %v with initPubKey : %v , node: %v , content: %v",
		args.AddStr, args.InitPubKey, args.Node, args.ApplyInfo)
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult
	res.Tip = "Your ReqId is: " + hex.EncodeToString(reqId[:]) +
		" , You can get the transaction hash with dag.getTxByReqId()"

	return res, nil
}

func (a *PrivateMediatorAPI) PayDeposit(from string, amount decimal.Decimal) (*TxExecuteResult, error) {
	// 参数检查
	fromAdd, err := common.StringToAddress(from)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", from)
	}

	if !amount.IsPositive() {
		return nil, fmt.Errorf("the amount of the deposit must be greater than 0")
	}

	// 调用系统合约
	cArgs := [][]byte{[]byte(modules.MediatorPayDeposit)}
	fee := a.Dag().CurrentFeeSchedule().TransferFee.BaseFee
	reqId, err := a.ContractInvokeReqTx(fromAdd, syscontract.DepositContractAddress, ptnjson.Ptn2Dao(amount),
		fee, nil, syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Account(%v) pay %vPTN to DepositContract(%v)",
		from, amount, syscontract.DepositContractAddress.Str())
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult
	res.Tip = "Your ReqId is: " + hex.EncodeToString(reqId[:]) +
		" , You can get the transaction hash with dag.getTxByReqId()"

	return res, nil
}

func (a *PrivateMediatorAPI) WithdrawDeposit(medAddStr string, amount decimal.Decimal) (*TxExecuteResult, error) {
	// 参数检查
	medAdd, err := common.StringToAddress(medAddStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", medAddStr)
	}

	// 判断是否是mediator
	if !a.Dag().IsMediator(medAdd) {
		return nil, fmt.Errorf("account %v is not a mediator", medAddStr)
	}

	if !amount.IsPositive() {
		return nil, fmt.Errorf("the amount of the deposit must be greater than 0")
	}

	// 调用系统合约
	amountStr := strconv.FormatUint(ptnjson.Ptn2Dao(amount), 10)
	cArgs := [][]byte{[]byte(modules.MediatorWithdrawDeposit), []byte(amountStr)}

	fee := a.Dag().CurrentFeeSchedule().TransferFee.BaseFee
	reqId, err := a.ContractInvokeReqTx(medAdd, medAdd, 0, fee,
		nil, syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Mediator(%v) apply to withdraw %vPTN to DepositContract(%v)",
		medAddStr, amount, syscontract.DepositContractAddress.Str())
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult
	res.Tip = "Your ReqId is: " + hex.EncodeToString(reqId[:]) +
		" , You can get the transaction hash with dag.getTxByReqId()"

	return res, nil
}

func (a *PrivateMediatorAPI) Quit(medAddStr string) (*TxExecuteResult, error) {
	// 参数检查
	medAdd, err := common.StringToAddress(medAddStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", medAddStr)
	}

	// 判断是否是mediator
	if !a.Dag().IsMediator(medAdd) {
		return nil, fmt.Errorf("account %v is not a mediator", medAddStr)
	}

	// 调用系统合约
	cArgs := [][]byte{[]byte(modules.MediatorApplyQuitList)}
	fee := a.Dag().CurrentFeeSchedule().TransferFee.BaseFee
	reqId, err := a.ContractInvokeReqTx(medAdd, medAdd, 0, fee,
		nil, syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("mediator(%v) apply to quit list", medAddStr)
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult
	res.Tip = "Your ReqId is: " + hex.EncodeToString(reqId[:]) +
		" , You can get the transaction hash with dag.getTxByReqId()"

	return res, nil
}

func (a *PrivateMediatorAPI) Vote(voterStr string, mediatorStrs []string) (*TxExecuteResult, error) {
	// 参数检查
	voter, err := common.StringToAddress(voterStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", voterStr)
	}

	// 判断本节点是否同步完成，数据是否最新
	//if !a.dag.IsSynced() {
	//	return nil, fmt.Errorf("the data of this node is not synced, and can't vote now")
	//}

	mp := make(map[string]bool)
	for _, mediatorStr := range mediatorStrs {
		mediator, err := common.StringToAddress(mediatorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid account address: %v", mediatorStr)
		}

		// 判断是否是mediator
		if !a.Dag().IsMediator(mediator) {
			return nil, fmt.Errorf("%v is not mediator", mediatorStr)
		}

		if mp[mediatorStr] {
			return nil, fmt.Errorf("this mediator(%v) has already been voted", mediatorStr)
		}

		mp[mediatorStr] = true
	}

	// 1. 创建交易
	tx, fee, err := a.Dag().GenVoteMediatorTx(voter, mp, a.TxPool())
	if err != nil {
		return nil, err
	}

	// 2. 签名和发送交易
	err = a.SignAndSendTransaction(voter, tx)
	if err != nil {
		return nil, err
	}

	// 5. 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Account %v vote mediator(s) %v", voterStr, mediatorStrs)
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size().TerminalString()
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult

	return res, nil
}
