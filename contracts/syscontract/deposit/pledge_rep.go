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
	log.Infof("pledgeRewardAllocation--haveCount = %d", haveCount)
	tmp := pledgeList.Members[haveCount:]
	for i := range tmp {
		//  最多循环 threshold 次
		if i+1 > threshold {
			log.Infof("break in i = %d, i + 1 = %d, t = %d", i, i+1, threshold)
			break
		}
		log.Infof("i = %d members[%d] was allocating...", i, i)
		reward := uint64(rewardPerDao * float64(tmp[i].Amount))
		newAmount := tmp[i].Amount + reward
		havePledgeList.Members = append(havePledgeList.Members, &modules.AddressRewardAmount{
			Address: tmp[i].Address,
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

//  增发分红奖励
func payoutDepositDailyReward(stub shim.ChaincodeStubInterface, depositDailyReward uint64) error {
	if depositDailyReward > 0 {
		//增发到合约
		log.Debugf("Create coinbase %d to pledge contract", depositDailyReward)
		err := stub.SupplyToken(dagconfig.DagConfig.GetGasToken().Bytes(),
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, depositDailyReward, syscontract.DepositContractAddress.String())
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

//  增加当前分红地址的新质押
func addDepositRecords(stub shim.ChaincodeStubInterface, pledgeList *modules.PledgeList) error {
	log.Info("enter addDepositRecords")
	for _, m := range pledgeList.Members {
		aab, err := stub.GetState(string(constants.PLEDGE_DEPOSIT_PREFIX) + m.Address)
		if err != nil {
			return err
		}
		if aab != nil {
			log.Infof("address = %s, amount = %d", m.Address, m.Amount)
			aa := modules.AddressAmount{}
			err = json.Unmarshal(aab, &aa)
			if err != nil {
				return err
			}
			pledgeList.Add(aa.Address, aa.Amount, 0)
			err = delPledgeDepositRecord(stub, aa.Address)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//  处理当前分红地址的提取
func withdrawRecords(stub shim.ChaincodeStubInterface, pledgeList *modules.PledgeList) error {
	log.Info("enter withdrawRecords")
	//处理提币请求
	for _, m := range pledgeList.Members {
		aab, err := stub.GetState(string(constants.PLEDGE_WITHDRAW_PREFIX) + m.Address)
		if err != nil {
			return err
		}
		if aab != nil {
			log.Infof("address = %s, amount = %d", m.Address, m.Amount)
			aa := modules.AddressAmount{}
			err = json.Unmarshal(aab, &aa)
			if err != nil {
				return err
			}
			gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
			withdrawAmt, err := pledgeList.Reduce(aa.Address, aa.Amount)
			if err != nil {
				log.Warnf("address[%s] withdraw pledge %d error:", aa.Address, aa.Amount)
			}
			if withdrawAmt > 0 {
				err := stub.PayOutToken(aa.Address, modules.NewAmountAsset(withdrawAmt, gasToken), 0)
				if err != nil {
					return err
				}
				err = delPledgeWithdrawRecord(stub, aa.Address) //清空提取请求列表
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//  继续按批次分红
//func handleRewardAllocationContinue(stub shim.ChaincodeStubInterface, pledgeList *modules.PledgeList, rewardPerDao float64, haveCount int, threshold int, today string, depositDailyReward uint64) error {
//	//  分红当前阈值之内的地址
//	allM := pledgeRewardAllocation(pledgeList, rewardPerDao, haveCount, threshold)
//	allM.Date = today
//	//  立即添加新质押
//	err := addDepositRecords(stub, allM)
//	if err != nil {
//		return err
//	}
//	//  立即添加提取
//	err = withdrawRecords(stub, allM)
//	if err != nil {
//		return err
//	}
//	b, err := json.Marshal(allM)
//	if err != nil {
//		return err
//	}
//	count := haveCount / threshold
//	if count == 0 {
//		err = stub.PutState(constants.PledgeList+allM.Date, b)
//		if err != nil {
//			return err
//		}
//	} else {
//		err = stub.PutState(constants.PledgeList+allM.Date+strconv.Itoa(count), b)
//		if err != nil {
//			return err
//		}
//	}
//	//  如果已经分配的 haveCount 大于等于 len(allM.Members) 个数，则证明按批次已经分配完成
//	if haveCount+threshold >= len(pledgeList.Members) {
//		log.Info("over...")
//		err := handleRewardAllocationOver(stub, today, depositDailyReward)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//	err = stub.PutState("haveAllocatedCount", []byte(strconv.Itoa(haveCount+threshold)))
//	if err != nil {
//		return err
//	}
//	return nil
//}

//  完成按批次分红
func handleRewardAllocationOver(stub shim.ChaincodeStubInterface, today string, depositDailyReward uint64) error {
	//  置为0
	err := stub.DelState("haveAllocatedCount")
	if err != nil {
		return err
	}
	err = stub.PutState("allocate", []byte("allocate"))
	if err != nil {
		return err
	}
	//err = saveLastPledgeListDate(stub, today)
	//if err != nil {
	//	return err
	//}
	err = stub.PutState(constants.AddNewAddress, []byte(today))
	if err != nil {
		return err
	}
	err = payoutDepositDailyReward(stub, depositDailyReward)
	if err != nil {
		return err
	}
	return nil
}

//  需要按批次分红
func need(stub shim.ChaincodeStubInterface, haveCount int, pledgeList *modules.PledgeList, today string, rewardPerDao float64, depositDailyReward uint64, threshold int) error {
	haveAllocatedCount, _ := stub.GetState("haveAllocatedCount")
	if haveAllocatedCount != nil {
		haveCount, _ = strconv.Atoi(string(haveAllocatedCount))
	}
	log.Infof("rewardPerDao = %f,haveCount = %d", rewardPerDao, haveCount)
	//  分红当前阈值之内的地址
	allM := pledgeRewardAllocation(pledgeList, rewardPerDao, haveCount, threshold)
	allM.Date = today
	//  立即添加新质押
	err := addDepositRecords(stub, allM)
	if err != nil {
		return err
	}
	//  立即添加提取
	err = withdrawRecords(stub, allM)
	if err != nil {
		return err
	}
	b, err := json.Marshal(allM)
	if err != nil {
		return err
	}
	count := haveCount / threshold
	log.Infof("continue count = %d", haveCount/threshold)
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
	haveCount += threshold
	log.Infof("haveCount = %d", haveCount)
	//  如果已经分配的 haveCount 大于等于 len(allM.Members) 个数，则证明按批次已经分配完成
	if haveCount >= len(pledgeList.Members) {
		log.Info("over...")
		err := handleRewardAllocationOver(stub, today, depositDailyReward)
		if err != nil {
			return err
		}
		return nil
	}
	err = stub.PutState("haveAllocatedCount", []byte(strconv.Itoa(haveCount)))
	if err != nil {
		return err
	}

	return nil
}

//  不需要分批
func unNeed(stub shim.ChaincodeStubInterface, haveCount int, allM *modules.PledgeList, today string, rewardPerDao float64, depositDailyReward uint64, threshold int) error {
	allM = pledgeRewardAllocation(allM, rewardPerDao, haveCount, threshold)
	allM.Date = today
	//  立即添加新质押
	err := addDepositRecords(stub, allM)
	if err != nil {
		return err
	}
	//  立即添加提取
	err = withdrawRecords(stub, allM)
	if err != nil {
		return err
	}
	//err = saveLastPledgeList(stub, allM)
	//if err != nil {
	//	return err
	//}

	b, err := json.Marshal(allM)
	if err != nil {
		return err
	}
	err = stub.PutState(constants.PledgeList+allM.Date, b)
	if err != nil {
		return err
	}
	err = stub.PutState(constants.AddNewAddress, []byte(allM.Date))
	if err != nil {
		return err
	}
	err = stub.PutState("allocate", []byte("allocate"))
	if err != nil {
		return err
	}
	err = payoutDepositDailyReward(stub, depositDailyReward)
	if err != nil {
		return err
	}
	return nil
}

//质押分红处理
func handleRewardAllocation(stub shim.ChaincodeStubInterface, depositDailyReward uint64, pledgeAllocateThreshold int) error {
	//  判断当天是否处理过
	today := getToday(stub)
	lastDate, err := getLastPledgeListDate(stub)
	if err != nil {
		return err
	}
	//  判断是否是基金会触发
	if isFoundationInvoke(stub) {
		t, _ := strconv.Atoi(lastDate)
		t += 1
		today = strconv.Itoa(t)
	} else {
		if lastDate == today {
			return fmt.Errorf("%s pledge reward has been allocated before", today)
		}

	}

	finish, err := stub.GetState("allocate")
	if err != nil {
		return err
	}

	//  第一次
	if lastDate == "" {
		finish = []byte("allocate")
	}
	allM, err := getLastPledgeList(stub)
	if err != nil {
		return err
	}
	//  计算分红
	if allM != nil && finish == nil {
		log.Infof("allM is not nil, today = %s, lastDate = %s", today, lastDate)
		//  当前的分红奖励与当前的分红数量的比例
		rewardPerDao := float64(depositDailyReward) / float64(allM.TotalAmount)
		threshold := pledgeAllocateThreshold
		//  获取已分红个数
		haveCount := 0
		//  判断是否超过默认个数
		if len(allM.Members) > threshold {
			log.Infof("handle need func, today = %s", today)
			err := need(stub, haveCount, allM, today, rewardPerDao, depositDailyReward, threshold)
			if err != nil {
				return err
			}
		} else {
			log.Infof("handle unNeed func, today = %s, lastDate = %s", today, lastDate)
			err := unNeed(stub, haveCount, allM, today, rewardPerDao, depositDailyReward, threshold)
			if err != nil {
				return err
			}
		}
	} else {
		newdate,err := stub.GetState(constants.AddNewAddress)
		if err != nil {
			return err
		}
		if newdate != nil {
			today = string(newdate)
		}
		//  添加新地址
		log.Infof("today = %s, lastDate = %s", today, newdate)
		return addNewAddrPledgeRecords(stub, today)
	}
	return nil
}

//func isAllocated(stub shim.ChaincodeStubInterface) bool {
//	date, err := getLastPledgeListDate(stub)
//	if err != nil {
//		return true
//	}
//	if date == "" {
//		return true
//	}
//	today := getToday(stub)
//	if date == today {
//		return true
//	}
//	return false
//}

//  增加新地址的质押
func addNewAddrPledgeRecords(stub shim.ChaincodeStubInterface, date string) error {
	// 增加新的质押
	depositList, err := getAllPledgeDepositRecords(stub)
	if err != nil {
		log.Info("getAllPledgeDepositRecords error: ", err.Error())
		return err
	}
	if len(depositList) != 0 {
		gp, err := stub.GetSystemConfig()
		if err != nil {
			return err
		}
		h := 0
		t := gp.ChainParameters.PledgeRecordsThreshold
		log.Infof("depositList lens = %d", len(depositList))
		haveAllocatedCount, _ := stub.GetState("haveAllocatedCount")
		if haveAllocatedCount != nil {
			h, _ = strconv.Atoi(string(haveAllocatedCount))
		}
		pledgeList := &modules.PledgeList{}
		for i, m := range depositList {
			if i+1 > t {
				log.Infof("break in i = %d, i + 1 = %d, t = %d", i, i+1, t)
				break
			}
			pledgeList.Date = date
			pledgeList.Add(m.Address, m.Amount, 0)
			err = delPledgeDepositRecord(stub, m.Address)
			if err != nil {
				return err
			}
		}
		b, err := json.Marshal(pledgeList)
		if err != nil {
			return err
		}
		err = stub.PutState(constants.PledgeList+date+"addNew"+strconv.Itoa(h/t), b)
		if err != nil {
			return err
		}
		if len(depositList) <= t {
			//  置为0
			err = stub.DelState("haveAllocatedCount")
			if err != nil {
				return err
			}
			err = stub.DelState("allocate")
			if err != nil {
				return err
			}
			err = stub.PutState(constants.AddNewAddress, []byte(date))
			if err != nil {
				return err
			}
			//  TODO 当第一次调用该函数
			err = saveLastPledgeListDate(stub, date)
			if err != nil {
				return err
			}
			return nil
		}
		h += t
		err = stub.PutState("haveAllocatedCount", []byte(strconv.Itoa(h)))
		if err != nil {
			return err
		}
		return nil
	}
	err = stub.DelState("allocate")
	if err != nil {
		return err
	}
	err = saveLastPledgeListDate(stub, date)
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

func getTotalPledgeStatus(stub shim.ChaincodeStubInterface) (*modules.PledgeStatus, error) {

	d, err := getAllPledgeDepositRecords(stub)
	if err != nil {
		return nil, err
	}
	totalDeposit := uint64(0)
	for _, dep := range d {
		totalDeposit += dep.Amount
	}
	//w, err := getPledgeWithdrawRecord(stub, addr)
	//if err != nil {
	//	return nil, err
	//}
	list, err := getLastPledgeList(stub)
	if err != nil {
		return nil, err
	}
	status := &modules.PledgeStatus{}
	if d != nil {
		status.NewDepositAmount = totalDeposit
	}
	//if w != nil {
	//	status.WithdrawApplyAmount = w.Amount
	//}
	if list != nil {
		status.PledgeAmount = list.TotalAmount
	}

	return status, nil
}
