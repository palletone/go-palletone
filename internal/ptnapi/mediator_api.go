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
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/dagconfig"
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

func getDeposit(addStr string, a Backend) (*modules.MediatorDepositJson, error) {
	// 构建参数
	cArgs := [][]byte{defaultMsg0, defaultMsg1, []byte(modules.GetMediatorDeposit), []byte(addStr)}
	txid := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().Unix())).Int31n(100000000))

	// 调用系统合约
	rsp, err := a.ContractQuery(syscontract.DepositContractAddress.Bytes(), txid[:], cArgs, 0)
	if err != nil {
		return nil, err
	}

	depositB := &modules.MediatorDepositJson{}
	err = json.Unmarshal(rsp, depositB)
	if err == nil {
		return depositB, nil
	}

	return nil, fmt.Errorf(string(rsp))
}

func (a *PublicMediatorAPI) GetDeposit(addStr string) (*modules.MediatorDepositJson, error) {
	// 参数检查
	_, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", addStr)
	}

	return getDeposit(addStr, a.Backend)
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

	for address := range mas {
		addStrs = append(addStrs, address.Str())
	}

	return addStrs
}

func (a *PublicMediatorAPI) ListVoteResults() map[string]uint64 {
	res, _ := a.Dag().MediatorVotedResults()
	return res
}

func (a *PublicMediatorAPI) ListVotingFor(addStr string) (map[string]uint64, error) {
	mediator, err := common.StringToAddress(addStr)
	if err != nil {
		return nil, err
	}

	if !a.Dag().IsMediator(mediator) {
		return nil, fmt.Errorf("%v is not mediator", addStr)
	}

	res, _ := a.Dag().GetVotingForMediator(addStr)
	return res, nil
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

	for med := range voted {
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

	if !a.Dag().IsMediator(mediator) {
		return nil, fmt.Errorf("%v is not mediator", addStr)
	}

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
	TxContent string      `json:"txContent"` // 交易内容
	TxHash    common.Hash `json:"txHash"`    // 交易hash
	TxSize    string      `json:"txSize"`    // 交易大小
	TxFee     string      `json:"txFee"`     // 交易费用
	Tip       string      `json:"tip"`       // 提示
	Warning   string      `json:"warning"`   // 警告
}

func (a *PrivateMediatorAPI) Apply(args modules.MediatorCreateArgs) (*TxExecuteResult, error) {
	// 参数补全
	if args.MediatorApplyInfo == nil {
		args.MediatorApplyInfo = core.NewMediatorApplyInfo()
	}

	// 参数验证
	if args.MediatorInfoBase == nil {
		return nil, fmt.Errorf("invalid args, is null")
	}

	addr, err := args.Validate()
	if err != nil {
		return nil, err
	}

	// 判断本节点是否同步完成，数据是否最新
	if !a.Dag().IsSynced() {
		return nil, fmt.Errorf("this node is not synced, and can't apply mediator now")
	}

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
	fee := a.Dag().GetChainParameters().MediatorCreateFee
	reqId, err := a.ContractInvokeReqTx(addr, addr, 0, fee, nil,
		syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("account(%v) apply mediator with rewardAdd: %v, initPubKey: %v, node: %v, "+
		"name: %v, url: %v, logo: %v, location: %v, applyInfo: %v", args.AddStr, args.RewardAdd, args.InitPubKey,
		args.Node, args.Name, args.Url, args.Logo, args.Location, args.Description)
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

	// 判断本节点是否同步完成，数据是否最新
	if !a.Dag().IsSynced() {
		return nil, fmt.Errorf("this node is not synced, and can't pay deposit now")
	}

	// 判断是否已经申请过，即是否创建保证金对象
	md, err := getDeposit(from, a.Backend)
	if err != nil {
		return nil, fmt.Errorf("account %v does not apply for mediator", from)
	}

	// 判断保证金是否已交齐
	cp := a.Dag().GetChainParameters()
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	dam := gasToken.DisplayAmount(cp.DepositAmountForMediator)
	if dam.Equal(md.Balance) {
		return nil, fmt.Errorf("the deposit of account %v is enough %v", from, dam.String())
	}

	// 判断是否已经是mediator
	// TODO 不满足追缴逻辑
	//if a.Dag().IsMediator(fromAdd) {
	//	return nil, fmt.Errorf("account %v is already a mediator", from)
	//}

	// 调用系统合约
	cArgs := [][]byte{[]byte(modules.MediatorPayDeposit)}
	reqId, err := a.ContractInvokeReqTx(fromAdd, syscontract.DepositContractAddress, ptnjson.Ptn2Dao(amount),
		cp.TransferPtnBaseFee, nil, syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Account(%v) pay %vPTN to DepositContract(%v)",
		from, amount, syscontract.DepositContractAddress.Str())
	res.TxFee = fmt.Sprintf("%vdao", cp.TransferPtnBaseFee)
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

	// 判断本节点是否同步完成，数据是否最新
	if !a.Dag().IsSynced() {
		return nil, fmt.Errorf("this node is not synced, and can't quit now")
	}

	// 判断是否是mediator
	if !a.Dag().IsMediator(medAdd) {
		return nil, fmt.Errorf("account %v is not a mediator", medAddStr)
	}

	// 调用系统合约
	cArgs := [][]byte{[]byte(modules.MediatorApplyQuit)}
	fee := a.Dag().GetChainParameters().TransferPtnBaseFee
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
	if !a.Dag().IsSynced() {
		return nil, fmt.Errorf("this node is not synced, and can't vote now")
	}

	maxMediatorCount := int(a.Dag().GetChainParameters().MaximumMediatorCount)
	mediatorCount := len(mediatorStrs)
	if mediatorCount > maxMediatorCount {
		return nil, fmt.Errorf("the total number(%v) of mediators voted exceeds the maximum limit: %v",
			mediatorCount, maxMediatorCount)
	}

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

	// 创建交易
	tx, fee, err := a.Dag().GenVoteMediatorTx(voter, mp, a.TxPool())
	if err != nil {
		return nil, err
	}

	// 签名和发送交易
	err = a.SignAndSendTransaction(voter, tx)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("Account %v vote mediator(s) %v", voterStr, mediatorStrs)
	res.TxHash = tx.Hash()
	res.TxSize = tx.Size().TerminalString()
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult

	return res, nil
}

func (a *PrivateMediatorAPI) Update(args modules.MediatorUpdateArgs) (*TxExecuteResult, error) {
	// 参数验证
	addr, err := args.Validate()
	if err != nil {
		return nil, err
	}

	// 判断本节点是否同步完成，数据是否最新
	if !a.Dag().IsSynced() {
		return nil, fmt.Errorf("this node is not synced, and can't update mediator now")
	}

	// 判断是否已经是mediator
	if !a.Dag().IsMediator(addr) {
		return nil, fmt.Errorf("account %v is not a mediator", args.AddStr)
	}

	// 参数序列化
	argsB, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	cArgs := [][]byte{[]byte(modules.UpdateMediatorInfo), argsB}

	// 调用系统合约
	fee := a.Dag().GetChainParameters().TransferPtnBaseFee
	reqId, err := a.ContractInvokeReqTx(addr, addr, 0, fee, nil,
		syscontract.DepositContractAddress, cArgs, 0)
	if err != nil {
		return nil, err
	}

	// 返回执行结果
	logoStr := ""
	if args.Logo != nil {
		logoStr = *args.Logo
	}
	nameStr := ""
	if args.Name != nil {
		nameStr = *args.Name
	}
	locStr := ""
	if args.Location != nil {
		locStr = *args.Location
	}
	urlStr := ""
	if args.Url != nil {
		urlStr = *args.Url
	}
	descStr := ""
	if args.Description != nil {
		descStr = *args.Description
	}
	nodeStr := ""
	if args.Node != nil {
		nodeStr = *args.Node
	}
	initPubKeyStr := ""
	if args.InitPubKey != nil {
		initPubKeyStr = *args.InitPubKey
	}
	rewardAddStr := ""
	if args.RewardAdd != nil {
		rewardAddStr = *args.RewardAdd
	}

	res := &TxExecuteResult{}
	res.TxContent = fmt.Sprintf("mediator(%v) update info with name: %v, url: %v logo: %v, location: %v, "+
		"applyInfo: %v, node: %v, initPubKey: %v, rewardAdd: %v", args.AddStr, nameStr, urlStr, logoStr, locStr,
		descStr, nodeStr, initPubKeyStr, rewardAddStr)
	res.TxFee = fmt.Sprintf("%vdao", fee)
	res.Warning = DefaultResult
	res.Tip = "Your ReqId is: " + hex.EncodeToString(reqId[:]) +
		" , You can get the transaction hash with dag.getTxByReqId()"

	return res, nil
}
