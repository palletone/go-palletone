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
	"math"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	//bigZero = big.NewInt(0)
	bigOne  = big.NewInt(1)
	ratZero = big.NewRat(0, 1)
	ratOne  = big.NewRat(1, 1)
)

type Binomial interface {
	CDF(int64) *big.Rat
	Prob(int64) *big.Rat
}

// ApproxBinomial is an approximate distribution function of Binomial distribution.
// It can only be used when τ >>> W (when p = τ / W ).
type ApproxBinomial struct {
	expected  *big.Int
	N         uint64
	probCache sync.Map //map[int64]*big.Rat
}

func NewApproxBinomial(expected int64, n uint64) *ApproxBinomial {
	return &ApproxBinomial{
		expected: big.NewInt(expected),
		N:        n,
	}
}

func (ab *ApproxBinomial) CDF(k int64) *big.Rat {
	if k < 0 {
		return ratZero
	}
	if uint64(k) >= ab.N {
		return ratOne
	}
	j := int64(0)
	res := new(big.Rat).SetInt64(0)
	for j <= k {
		res.Add(res, ab.Prob(j))
		j++
	}
	return res
}

func (ab *ApproxBinomial) Prob(k int64) *big.Rat {
	if prob, exist := ab.probCache.Load(k); exist {
		return prob.(*big.Rat)
	}
	numer := new(big.Int).Exp(ab.expected, big.NewInt(k), nil)
	denom := new(big.Int).MulRange(1, k)
	e := new(big.Rat).SetFloat64(math.Pow(math.E, -float64(ab.expected.Int64())))
	lval := new(big.Rat).SetFrac(numer, denom)

	prob := lval.Mul(lval, e)
	ab.probCache.Store(k, prob)
	return prob
}

// Binomial implements the binomial distribution function.
// It's too slow!!!
type RegularBinomial struct {
	N *big.Int

	pn *big.Int // numerator of P
	pd *big.Int // denominator of P

	pnExpCache    sync.Map // map[power(int64)]*big.Int
	tpnExpCache   sync.Map // the same with above
	probCache     sync.Map // map[int64]*big.Rat
	cdfCache      sync.Map // map[int64]*big.Rat
	resDenomCache atomic.Value
}

func NewBinomial(n, pn, pd int64) *RegularBinomial {
	return &RegularBinomial{
		N:  big.NewInt(n),
		pn: big.NewInt(pn),
		pd: big.NewInt(pd),
	}
}

// exp computes x**y with a optimize way.
func (b *RegularBinomial) Exp(x *big.Int, y float64) *big.Int {
	if y == 0 {
		return bigOne
	}

	if y == 1 {
		return x
	}

	var (
		cache *sync.Map
		res   *big.Int
	)
	// get the corresponding cache map
	if x.Cmp(b.pn) == 0 {
		cache = &b.pnExpCache
	} else if x.Cmp(new(big.Int).Sub(b.pd, b.pn)) == 0 {
		cache = &b.tpnExpCache
	}
	if cache != nil {
		if exp, exist := cache.Load(y); exist {
			return exp.(*big.Int)
		}
	}

	if y == 2 {
		res = new(big.Int).Mul(x, x)
	} else {
		ly := math.Floor(math.Sqrt(y))
		ry := math.Floor(y - ly)
		res = new(big.Int).Mul(b.Exp(x, ly), b.Exp(x, ry))
	}
	if cache != nil {
		cache.Store(y, res)
	}
	return res
}

//二项式累计分布函数
func (b *RegularBinomial) CDF(j int64) *big.Rat {
	if j < 0 {
		return new(big.Rat).SetInt64(0)
	}
	if j >= b.N.Int64() {
		return new(big.Rat).SetInt64(1)
	}

	if cdf, exist := b.cdfCache.Load(j); exist {
		return cdf.(*big.Rat)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	k := j - 1
	var res = new(big.Rat).SetInt64(0)
	for k >= 0 {
		if cdf, exist := b.cdfCache.Load(k); exist {
			res = new(big.Rat).Set(cdf.(*big.Rat))
			break
		}
		k--
	}
	resChan := make(chan *big.Rat, j+1)
	k++
	begin := k
	for k <= j {
		go func(i int64) {
			prob := b.Prob(i)
			resChan <- prob
		}(k)
		k++
	}
	k = begin

	for k <= j {
		select {
		case rat := <-resChan:
			res.Add(res, rat)
			k++
		}
	}
	close(resChan)
	b.cdfCache.Store(j, res)
	return res
}

func (b *RegularBinomial) Prob(k int64) *big.Rat {
	if prob, exist := b.probCache.Load(k); exist {
		return prob.(*big.Rat)
	}

	// calculate p^k * p^(n-k)
	N := b.N.Int64()
	lnum := b.Exp(b.pn, float64(k))
	rnum := b.Exp(new(big.Int).Sub(b.pd, b.pn), float64(N-k))
	resNum := new(big.Int).Mul(lnum, rnum)

	mulRat := new(big.Rat).SetFrac(resNum, b.resDenom())

	// calculate C(n,k)
	bino := new(big.Int).Binomial(N, k)

	// prob = C(n,k) * p^k * p^(n-k)
	prob := new(big.Rat).Mul(new(big.Rat).SetInt(bino), mulRat)

	b.probCache.Store(k, prob)
	return prob
}

func (b *RegularBinomial) Probability() *big.Rat {
	return new(big.Rat).SetFrac(b.pn, b.pd)
}

func (b *RegularBinomial) resDenom() *big.Int {
	if res := b.resDenomCache.Load(); res != nil {
		return res.(*big.Int)
	}
	resDenom := b.Exp(b.pd, float64(b.N.Int64()))
	b.resDenomCache.Store(resDenom)
	return resDenom
}
