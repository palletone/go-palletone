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

package util

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGetSourceString(t *testing.T) {
	log.Println("bool", GetSourceString(true))
	log.Println("int64", GetSourceString(int64(123)))
	log.Println("float", GetSourceString(123.182318278))

	log.Println("[]", GetSourceString([]string{"1", "2", "3"}))
	var mapobj map[string]string
	mapobj = make(map[string]string)
	mapobj["lan"] = "22"
	mapobj["jay"] = "23"
	mapobj["ma"] = "24"
	log.Println("map", GetSourceString(mapobj))
	type obj struct {
		Name string
		Age  int
	}
	o := obj{
		Name: "jay",
		Age:  20,
	}

	arr := GetSourceString(o)
	log.Println("ok object", arr, f2(s))
}

func f1(s string) int {
	return bytes.Count([]byte(s), nil) - 1
}

func f2(s string) int {
	return strings.Count(s, "") - 1
}

func f3(s string) int {
	return len([]rune(s))
}

func f4(s string) int {
	return utf8.RuneCountInString(s)
}

var s = "Hello, 世界"

func Benchmark1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f1(s)
	}
}

func Benchmark2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f2(s)
	}
}

func Benchmark3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f3(s)
	}
}

func Benchmark4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f4(s)
	}
}
