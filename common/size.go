// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

// StorageSize is a wrapper around a float value that supports user friendly
// formatting.
type StorageSize float64

// String implements the stringer interface.
func (s StorageSize) String() string {
	if s > 1024000 {
		return fmt.Sprintf("%.2f MB", s/1024000)
	} else if s > 1024 {
		return fmt.Sprintf("%.2f KB", s/1024)
	} else {
		return fmt.Sprintf("%.2f B", s)
	}
}

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (s StorageSize) TerminalString() string {
	if s > 1024000 {
		return fmt.Sprintf("%.2fmB", s/1024000)
	} else if s > 1024 {
		return fmt.Sprintf("%.2fkB", s/1024)
	} else {
		return fmt.Sprintf("%.2fB", s)
	}
}

func (s StorageSize) Float64() float64 {
	return float64(s)
}

func (s StorageSize) Bytes() []byte {
	bits := math.Float64bits(float64(s))
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}

// str to uint64
func Str2Int64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return -1
	}
	return num
}

func Str2Uint64(str string) uint64 {
	num := Str2Int64(str)
	if num < 0 {
		return 0
	}
	return uint64(num)
}
func Uint642Str(num uint64) string {
	return strconv.FormatInt(int64(num), 10)
}
