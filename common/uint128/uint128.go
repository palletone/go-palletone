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
 * @date 2018
 */

package uint128

import (
	"encoding/binary"
	"math/big"
	"math/bits"
)

// Zero is a zero-valued uint128.
var Zero Uint128

// MaxValue is a max value of uint128.
var MaxValue = New(^uint64(0), ^uint64(0))

// A Uint128 is an unsigned 128-bit number.
type Uint128 struct {
	lo, hi uint64
}

// IsZero returns true if u == 0.
func (u Uint128) IsZero() bool {
	return u == Zero
}

// Equals returns true if u == v.
//
// Uint128 values can be compared directly with ==, but use of the Equals method
// is preferred for consistency.
func (u Uint128) Equals(v Uint128) bool {
	return u == v
}

// Equals64 returns true if u == v.
func (u Uint128) Equals64(v uint64) bool {
	return u.lo == v && u.hi == 0
}

// Cmp compares u and v and returns:
//
//   -1 if u <  v
//    0 if u == v
//   +1 if u >  v
//
func (u Uint128) Cmp(v Uint128) int {
	if u == v {
		return 0
	} else if u.hi < v.hi || (u.hi == v.hi && u.lo < v.lo) {
		return -1
	} else {
		return 1
	}
}

// Cmp64 compares u and v and returns:
//
//   -1 if u <  v
//    0 if u == v
//   +1 if u >  v
//
func (u Uint128) Cmp64(v uint64) int {
	if u.hi == 0 && u.lo == v {
		return 0
	} else if u.hi == 0 && u.lo < v {
		return -1
	} else {
		return 1
	}
}

// And returns u&v.
func (u Uint128) And(v Uint128) Uint128 {
	return Uint128{u.lo & v.lo, u.hi & v.hi}
}

// And64 returns u&v.
func (u Uint128) And64(v uint64) Uint128 {
	return Uint128{u.lo & v, 0}
}

// Or returns u|v.
func (u Uint128) Or(v Uint128) Uint128 {
	return Uint128{u.lo | v.lo, u.hi | v.hi}
}

// Or64 returns u|v.
func (u Uint128) Or64(v uint64) Uint128 {
	return Uint128{u.lo | v, u.hi}
}

// Xor returns u^v.
func (u Uint128) Xor(v Uint128) Uint128 {
	return Uint128{u.lo ^ v.lo, u.hi ^ v.hi}
}

// Xor64 returns u^v.
func (u Uint128) Xor64(v uint64) Uint128 {
	return Uint128{u.lo ^ v, u.hi}
}

// Add returns u+v.
func (u Uint128) Add(v Uint128) Uint128 {
	lo, carry := bits.Add64(u.lo, v.lo, 0)
	hi, _ := bits.Add64(u.hi, v.hi, carry)
	return Uint128{lo, hi}
}

// Add64 returns u+v.
func (u Uint128) Add64(v uint64) Uint128 {
	lo, carry := bits.Add64(u.lo, v, 0)
	hi := u.hi + carry
	return Uint128{lo, hi}
}

// Sub returns u-v.
func (u Uint128) Sub(v Uint128) Uint128 {
	lo, borrow := bits.Sub64(u.lo, v.lo, 0)
	hi, _ := bits.Sub64(u.hi, v.hi, borrow)
	return Uint128{lo, hi}
}

// Sub64 returns u-v.
func (u Uint128) Sub64(v uint64) Uint128 {
	lo, borrow := bits.Sub64(u.lo, v, 0)
	hi := u.hi - borrow
	return Uint128{lo, hi}
}

// Mul returns u*v.
func (u Uint128) Mul(v Uint128) Uint128 {
	hi, lo := bits.Mul64(u.lo, v.lo)
	hi += u.hi*v.lo + u.lo*v.hi
	return Uint128{lo, hi}
}

// Mul64 returns u*v.
func (u Uint128) Mul64(v uint64) Uint128 {
	hi, lo := bits.Mul64(u.lo, v)
	hi += u.hi * v
	return Uint128{lo, hi}
}

// Div returns u/v.
func (u Uint128) Div(v Uint128) Uint128 {
	q, _ := u.QuoRem(v)
	return q
}

// Div64 returns u/v.
func (u Uint128) Div64(v uint64) Uint128 {
	q, _ := u.QuoRem64(v)
	return q
}

// QuoRem returns q = u/v and r = u%v.
func (u Uint128) QuoRem(v Uint128) (q, r Uint128) {
	if v.hi == 0 {
		var r64 uint64
		q, r64 = u.QuoRem64(v.lo)
		r = From64(r64)
	} else {
		// generate a "trial quotient," guaranteed to be within 1 of the actual
		// quotient, then adjust.
		n := uint(bits.LeadingZeros64(v.hi))
		v1 := v.Lsh(n)
		u1 := u.Rsh(1)
		tq, _ := bits.Div64(u1.hi, u1.lo, v1.hi)
		tq >>= 63 - n
		if tq != 0 {
			tq--
		}
		q = From64(tq)
		// calculate remainder using trial quotient, then adjust if remainder is
		// greater than divisor
		r = u.Sub(v.Mul64(tq))
		if r.Cmp(v) >= 0 {
			q = q.Add64(1)
			r = r.Sub(v)
		}
	}
	return
}

// QuoRem64 returns q = u/v and r = u%v.
func (u Uint128) QuoRem64(v uint64) (q Uint128, r uint64) {
	if u.hi < v {
		q.lo, r = bits.Div64(u.hi, u.lo, v)
	} else {
		q.hi, r = bits.Div64(0, u.hi, v)
		q.lo, r = bits.Div64(r, u.lo, v)
	}
	return
}

// Lsh returns u<<n.
func (u Uint128) Lsh(n uint) (s Uint128) {
	if n > 64 {
		s.lo = 0
		s.hi = u.lo << (n - 64)
	} else {
		s.lo = u.lo << n
		s.hi = u.hi<<n | u.lo>>(64-n)
	}
	return
}

// Rsh returns u>>n.
func (u Uint128) Rsh(n uint) (s Uint128) {
	if n > 64 {
		s.lo = u.hi >> (n - 64)
		s.hi = 0
	} else {
		s.lo = u.lo>>n | u.hi<<(64-n)
		s.hi = u.hi >> n
	}
	return
}

// String returns the base-10 representation of u as a string.
func (u Uint128) String() string {
	if u.IsZero() {
		return "0"
	}
	buf := []byte("0000000000000000000000000000000000000000") // log10(2^128) < 40
	for i := len(buf); ; i -= 19 {
		q, r := u.QuoRem64(1e19) // largest power of 10 that fits in a uint64
		var n int
		for ; r != 0; r /= 10 {
			n++
			buf[i-n] += byte(r % 10)
		}
		if q.IsZero() {
			return string(buf[i-n:])
		}
		u = q
	}
}

// PutBytes stores u in b in little-endian order. It panics if len(b) < 16.
func (u Uint128) PutBytes(b []byte) {
	binary.LittleEndian.PutUint64(b[:8], u.lo)
	binary.LittleEndian.PutUint64(b[8:], u.hi)
}

// Big returns u as a *big.Int.
func (u Uint128) Big() *big.Int {
	i := new(big.Int).SetUint64(u.hi)
	i = i.Lsh(i, 64)
	i = i.Xor(i, new(big.Int).SetUint64(u.lo))
	return i
}

// New returns the Uint128 value (lo,hi).
func New(lo, hi uint64) Uint128 {
	return Uint128{lo, hi}
}

// From64 converts v to a Uint128 value.
func From64(v uint64) Uint128 {
	return New(v, 0)
}

// FromBytes converts b to a Uint128 value.
func FromBytes(b []byte) Uint128 {
	return New(
		binary.LittleEndian.Uint64(b[:8]),
		binary.LittleEndian.Uint64(b[8:]),
	)
}

// FromBig converts i to a Uint128 value. It panics if i is negative or
// overflows 128 bits.
func FromBig(i *big.Int) (u Uint128) {
	if i.Sign() < 0 {
		panic("value cannot be negative")
	} else if i.BitLen() > 128 {
		panic("value overflows Uint128")
	}
	u.lo = i.Uint64()
	u.hi = new(big.Int).Rsh(i, 64).Uint64()
	return u
}

// PopCount64, count how many bits of a binary form corresponding to a uint64 type number are 1
func PopCount64(x uint64) uint8 {
	m := []uint64{
		0x5555555555555555,
		0x3333333333333333,
		0x0F0F0F0F0F0F0F0F,
		0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF,
		0x00000000FFFFFFFF,
	}

	var i, w uint8
	for i, w = 0, 1; i < 6; i, w = i+1, w+w {
		x = (x & m[i]) + ((x >> w) & m[i])
	}

	return uint8(x)
}

// PopCount, count how many bits of a binary form corresponding to a Uint128 type number are 1
func (u Uint128) PopCount() uint8 {
	return PopCount64(u.lo) + PopCount64(u.hi)
}
