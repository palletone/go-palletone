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
package algorithm

import (
	"math/big"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/util"
)

//subUsers return the selected amount of sub-users determined from the mathematics protocol.
//expectedNum 期望数量
//weight 设置固定数值,即返回概率值*weight,返回值落在0  --   expectedNum/Total*weight之间的数值
func Selected(expectedNum uint, weight, total uint64, vrf []byte) int {
	if total < 22 {
		return 1 //todo test, is danger
	}

	hh := util.RlpHash(vrf)
	h32 := hh[:common.HashLength]
	//Total := 100
	//weight=TokenPerUser; TotalTokenAmount = UserAmount * TokenPerUser
	binomial := NewBinomial(int64(weight), int64(expectedNum), int64(total))
	//binomial := NewApproxBinomial(int64(expectedNum), weight)
	//binomial := &distuv.Binomial{
	//	N: float64(weight),
	//	P: float64(expectedNum) / float64(TotalTokenAmount()),
	//}
	// hash / 2^hashlen ∉ [ ∑0,j B(k;w,p), ∑0,j+1 B(k;w,p))

	//	hashBig := new(big.Int).SetBytes(vrf)
	hashBig := new(big.Int).SetBytes(h32)
	maxHash := new(big.Int).Exp(big.NewInt(2), big.NewInt(common.HashLength*8), nil)
	hash := new(big.Rat).SetFrac(hashBig, maxHash)
	var lower, upper *big.Rat
	j := 0
	for uint64(j) <= weight {
		if upper != nil {
			lower = upper
		} else {
			lower = binomial.CDF(int64(j))
		}
		upper = binomial.CDF(int64(j + 1))
		//log.Infof("hash %v, lower %v , upper %v", hash.Sign(), lower.Sign(), upper.Sign())
		if hash.Cmp(lower) >= 0 && hash.Cmp(upper) < 0 {
			break
		}
		j++
	}
	//log.Infof("j %d", j)
	if uint64(j) > weight {
		j = 0
	}
	//j := parallelTrevels(runtime.NumCPU(), weight, hash, binomial)
	return j
}

//func parallelTrevels(core int, N uint64, hash *big.Rat, binomial Binomial) int {
//	var wg sync.WaitGroup
//	groups := N / uint64(core)
//	background, cancel := context.WithCancel(context.Background())
//	resChan := make(chan int)
//	notFound := make(chan struct{})
//	for i := 0; i < core; i++ {
//		go func(ctx context.Context, begin uint64) {
//			wg.Add(1)
//			defer wg.Done()
//			var (
//				end          uint64
//				upper, lower *big.Rat
//			)
//			if begin == uint64(core-2) {
//				end = N + 1
//			} else {
//				end = groups * (begin + 1)
//			}
//			for j := groups * begin; j < end; j++ {
//				select {
//				case <-ctx.Done():
//					return
//				default:
//				}
//				if upper != nil {
//					lower = upper
//				} else {
//					lower = binomial.CDF(int64(j))
//				}
//				upper = binomial.CDF(int64(j + 1))
//				//log.Infof("hash %v, lower %v , upper %v", hash.Sign(), lower.Sign(), upper.Sign())
//				if hash.Cmp(lower) >= 0 && hash.Cmp(upper) < 0 {
//					resChan <- int(j)
//					return
//				}
//				j++
//			}
//			return
//		}(background, uint64(i))
//	}
//
//	go func() {
//		wg.Wait()
//		close(notFound)
//	}()
//
//	select {
//	case j := <-resChan:
//		cancel()
//		return j
//	case <-notFound:
//		return 0
//	}
//}
