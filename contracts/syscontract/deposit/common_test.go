///*
//	This file is part of go-palletone.
//	go-palletone is free software: you can redistribute it and/or modify
//	it under the terms of the GNU General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//	go-palletone is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.
//	You should have received a copy of the GNU General Public License
//	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
//*/
//
package deposit

import (
	"fmt"
	"github.com/palletone/go-palletone/common/hexutil"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	// 通过当前时间格式化
	now := time.Now().UTC()
	fmt.Println(now)
	l1 := now.Format("2006-01-02 15")
	fmt.Println(l1)
	l2 := now.Format("2006-01-02 15:04")
	fmt.Println(l2)
	l3 := now.Format("2006-01-02 15:04:05")
	fmt.Println(l3)
	//
	t1, _ := time.Parse("2006-01-02 15", l1)
	fmt.Println(t1)
	t2, _ := time.Parse("2006-01-02 15:04", l2)
	fmt.Println(t2)
	t3, _ := time.Parse("2006-01-02 15:04:05", l3)
	fmt.Println(t3)
	fmt.Println(t1.String())
}

func TestArray(t *testing.T) {
	var arr []string
	for i, v := range arr {
		fmt.Println(i, v)
	}
}

func TestLaa(t *testing.T) {
	encode := "0x03fb01988b65751b86d10fc7bfc34b127febb0602ca64edd42003274640bf5148c"
	fmt.Println(len(encode))
	byte, _ := hexutil.Decode(encode)
	fmt.Println(len(byte))
	encode2 := hexutil.Encode(byte)
	if encode == encode2 {
		t.Log("success")
		return
	}
	t.Error("faild")
}
