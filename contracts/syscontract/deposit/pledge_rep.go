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

//记录了所有用户的质押充币、提币、分红等过程
//最新状态集
//Advance：形成流水日志，
package deposit

import (
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"strconv"

	//pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	//"github.com/palletone/go-palletone/dag/constants"
	//"github.com/palletone/go-palletone/dag/modules"
	//"github.com/shopspring/decimal"
	"fmt"
	"github.com/palletone/go-palletone/dag/modules"
)

//质押充币
func pledgeDepositRep(stub shim.ChaincodeStubInterface, addr common.Address, amount uint64) error {
	addrStr := addr.String()
	node, err := getPledgeDepositRecord(stub, addrStr)
	if err != nil {
		return err
	}
	if node == nil {
		node = &modules.AddressAmount{}
	}
	node.Amount += amount
	node.Address = addrStr
	return savePledgeDepositRecord(stub, node)
}

//质押提币，以当天最后一次为准
func pledgeWithdrawRep(stub shim.ChaincodeStubInterface, addr common.Address, amount uint64) error {
	err := savePledgeWithdrawRecord(stub, modules.NewAddressAmount(addr.String(), amount))
	return err
}

//撤销当天的提币请求
//func pledgeWithdrawCancelRep(stub shim.ChaincodeStubInterface, addr common.Address) error {
//	err := delPledgeWithdrawRecord(stub, addr.String())
//	return err
//}

//质押分红,按持仓比例分固定金额
func pledgeRewardAllocation(pledgeList *modules.PledgeList, rewardPerDao float64, haveCount int, threshold int) *modules.PledgeList {
	havePledgeList := &modules.PledgeList{TotalAmount: 0, Members: []*modules.AddressRewardAmount{}}
	log.Infof("pledgeRewardAllocation--haveCount = %s", haveCount)
	pledgeList.Members = pledgeList.Members[haveCount:]
	for i := range pledgeList.Members {
		//  最多循环 threshold 次
		if i+1 > threshold {
			log.Infof("break in i = %d, i + 1 = %d, t = %d", i, i+1, threshold)
			break
		}
		log.Infof("i = %d members[%d] was allocating...", i, i)
		reward := uint64(rewardPerDao * float64(pledgeList.Members[i].Amount))
		newAmount := pledgeList.Members[i].Amount + reward
		havePledgeList.Members = append(havePledgeList.Members, &modules.AddressRewardAmount{
			Address: pledgeList.Members[i].Address,
			Reward:  reward,
			Amount:  newAmount})
		havePledgeList.TotalAmount += newAmount
	}
	return havePledgeList
}

//质押分红,按持仓比例分固定金额
//func pledgeRewardAllocation(pledgeList *modules.PledgeList, rewardAmount uint64) *modules.PledgeList {
//	newPledgeList := &modules.PledgeList{TotalAmount: 0, Members: []*modules.AddressRewardAmount{}}
//	rewardPerDao := float64(rewardAmount) / float64(pledgeList.TotalAmount)
//	for _, pledge := range pledgeList.Members {
//		reward := uint64(rewardPerDao * float64(pledge.Amount))
//		newAmount := pledge.Amount + reward
//		newPledgeList.Members = append(newPledgeList.Members, &modules.AddressRewardAmount{
//			Address: pledge.Address,
//			Reward:  reward,
//			Amount:  newAmount})
//		newPledgeList.TotalAmount += newAmount
//	}
//	return newPledgeList
//}

//质押分红处理
func handleRewardAllocation(stub shim.ChaincodeStubInterface, depositDailyReward uint64) error {
	//  判断当天是否处理过
	today := getToday(stub)
	lastDate, err := getLastPledgeListDate(stub)
	if err != nil {
		return err
	}
	//判断是否是基金会触发
	if isFoundationInvoke(stub) {
		t, _ := strconv.Atoi(lastDate)
		t += 1
		today = strconv.Itoa(t)
	} else {
		if lastDate == today {
			return fmt.Errorf("%s pledge reward has been allocated before", today)
		}
	}
	allM, err := getLastPledgeList(stub)
	if err != nil {
		return err
	}
	//计算分红
	if allM != nil {
		log.Info("enter xiaozhi")
		//  当前的分红奖励与当前的分红数量的比例
		rewardPerDao := float64(depositDailyReward) / float64(allM.TotalAmount)
		//len(allM.Members) = 3
		//  当前分红默认处理个数
		//threshold := 1
		//threshold := 2
		threshold := 3
		//threshold := 4
		//  获取已分红个数
		haveCount := 0
		//  判断是否超过默认个数
		if len(allM.Members) > threshold {
			log.Info("xuyaofenpi====>enter xiaozhi 1")
			haveAllcocatedCount, _ := stub.GetState("haveAllcocatedCount")
			if haveAllcocatedCount != nil {
				haveCount, _ = strconv.Atoi(string(haveAllcocatedCount))
			}
			log.Infof("rewardPerDao = %s,haveCount = %d", rewardPerDao, haveCount)
			//  如果已经分配的 haveCount 大于等于 len(allM.Members) 个数，则证明按批次已经分配完成
			if haveCount >= len(allM.Members) {
				log.Info("over...")
				//  置为0
				err = stub.DelState("haveAllcocatedCount")
				if err != nil {
					return err
				}
				err = saveLastPledgeListDate(stub,today)
				if err != nil {
					return err
				}
				log.Info("44")
				if depositDailyReward > 0 {
					log.Infof("xiaozhi===>>>")
					//增发到合约
					log.Debugf("Create coinbase %d to pledge contract", depositDailyReward)
					err = stub.SupplyToken(dagconfig.DagConfig.GetGasToken().Bytes(),
						[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, depositDailyReward, syscontract.DepositContractAddress.String())
					if err != nil {
						return err
					}
				}
				return nil
			} else {
				log.Infof("continue...%d", haveCount/threshold)
				//  分红当前阈值之内的地址
				allM = pledgeRewardAllocation(allM, rewardPerDao, haveCount, threshold)
				allM.Date = today
				b, err := json.Marshal(allM)
				if err != nil {
					return err
				}
				count := haveCount / threshold
				if count == 0 {
					err = stub.PutState(constants.PledgeList+allM.Date, b)
					if err != nil {
						return err
					}
				} else {
					err = stub.PutState(constants.PledgeList+allM.Date+strconv.Itoa(count), b)
					if err != nil {
						return err
					}
				}
				err = stub.PutState("haveAllcocatedCount", []byte(strconv.Itoa(haveCount+threshold)))
				if err != nil {
					return err
				}
				return nil
			}
			log.Info("not passed")
		} else {
			log.Info("buxuyaofenpi===>enter xiaozhi 2")
			allM = pledgeRewardAllocation(allM, rewardPerDao, haveCount, threshold)
			allM.Date = today
			b, err := json.Marshal(allM)
			if err != nil {
				return err
			}
			err = stub.PutState(constants.PledgeList+allM.Date, b)
			if err != nil {
				return err
			}
			err = saveLastPledgeListDate(stub,today)
			if err != nil {
				return err
			}
			if depositDailyReward > 0 {
				log.Infof("xiaozhi===>>>")
				//增发到合约
				log.Debugf("Create coinbase %d to pledge contract", depositDailyReward)
				err = stub.SupplyToken(dagconfig.DagConfig.GetGasToken().Bytes(),
					[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, depositDailyReward, syscontract.DepositContractAddress.String())
				if err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		allM = &modules.PledgeList{}
	}
	log.Info("tiaojiaxinzhiya===")
	allM.Date = today
	// 增加新的质押
	depositList, err := getAllPledgeDepositRecords(stub)
	if err != nil {
		return err
	}

	for _, awardNode := range depositList {
		allM.Add(awardNode.Address, awardNode.Amount, 0) //新增加的质押当天不会有分红
		err = delPledgeDepositRecord(stub, awardNode.Address)
		if err != nil {
			return err
		}
	}
	b, err := json.Marshal(allM)
	if err != nil {
		return err
	}
	err = stub.PutState(constants.PledgeList+allM.Date, b)
	if err != nil {
		return err
	}
	err = saveLastPledgeListDate(stub,today)
	if err != nil {
		return err
	}
	return nil

	////计算分红
	//if allM != nil {
	//
	//	allM = pledgeRewardAllocation(allM, depositDailyReward)
	//} else {
	//	allM = &modules.PledgeList{}
	//}
	//allM.Date = today
	//// 增加新的质押
	//depositList, err := getAllPledgeDepositRecords(stub)
	//if err != nil {
	//	return err
	//}
	//
	//for _, awardNode := range depositList {
	//
	//	allM.Add(awardNode.Address, awardNode.Amount, 0) //新增加的质押当天不会有分红
	//	err = delPledgeDepositRecord(stub, awardNode.Address)
	//	if err != nil {
	//		return err
	//	}
	//}
	////处理提币请求
	//withdrawList, err := getAllPledgeWithdrawRecords(stub)
	//if err != nil {
	//	return err
	//}
	//gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	//for _, withdraw := range withdrawList {
	//	withdrawAmt, err := allM.Reduce(withdraw.Address, withdraw.Amount)
	//	if err != nil {
	//		log.Warnf("address[%s] withdraw pledge %d error:", withdraw.Address, withdraw.Amount)
	//	}
	//	if withdrawAmt > 0 {
	//		err := stub.PayOutToken(withdraw.Address, modules.NewAmountAsset(withdrawAmt, gasToken), 0)
	//		if err != nil {
	//			return err
	//		}
	//		err = delPledgeWithdrawRecord(stub, withdraw.Address) //清空提取请求列表
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}
	err = saveLastPledgeList(stub, allM)
	if err != nil {
		return err
	}
	return nil
}

//查询一个账户的质押状态
func getPledgeStatus(stub shim.ChaincodeStubInterface, addr string) (*modules.PledgeStatus, error) {
	//Check addr format
	_, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	d, err := getPledgeDepositRecord(stub, addr)
	if err != nil {
		return nil, err
	}
	w, err := getPledgeWithdrawRecord(stub, addr)
	if err != nil {
		return nil, err
	}
	list, err := getLastPledgeList(stub)
	if err != nil {
		return nil, err
	}
	status := &modules.PledgeStatus{}
	if d != nil {
		status.NewDepositAmount = d.Amount
	}
	if w != nil {
		status.WithdrawApplyAmount = w.Amount
	}
	if list != nil {
		status.PledgeAmount = list.GetAmount(addr)
	}

	return status, nil
}

//func getTotalPledgeStatus(stub shim.ChaincodeStubInterface) (*modules.PledgeStatus, error) {
//
//	d, err := getAllPledgeDepositRecords(stub)
//	if err != nil {
//		return nil, err
//	}
//	totalDeposit := uint64(0)
//	for _, dep := range d {
//		totalDeposit += dep.Amount
//	}
//	//w, err := getPledgeWithdrawRecord(stub, addr)
//	//if err != nil {
//	//	return nil, err
//	//}
//	list, err := getLastPledgeList(stub)
//	if err != nil {
//		return nil, err
//	}
//	status := &modules.PledgeStatus{}
//	if d != nil {
//		status.NewDepositAmount = totalDeposit
//	}
//	//if w != nil {
//	//	status.WithdrawApplyAmount = w.Amount
//	//}
//	if list != nil {
//		status.PledgeAmount = list.TotalAmount
//	}
//
//	return status, nil
//}
