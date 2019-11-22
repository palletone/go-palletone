// Copyright 2016 The go-ethereum Authors
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

package ptnclient

import (
	"github.com/palletone/go-palletone"
)

// Verify that Client implements the palletone interfaces.
var (
	_ = palletone.ChainReader(&Client{})
	_ = palletone.TransactionReader(&Client{})
	_ = palletone.ChainStateReader(&Client{})
	_ = palletone.ChainSyncReader(&Client{})
	_ = palletone.ContractCaller(&Client{})
	_ = palletone.GasEstimator(&Client{})
	_ = palletone.GasPricer(&Client{})
	//_ = palletone.LogFilterer(&Client{})
	_ = palletone.PendingStateReader(&Client{})
	// _ = palletone.PendingStateEventer(&Client{})
	_ = palletone.PendingContractCaller(&Client{})
)
/*func TestSimpleContractCcstop(t *testing.T) {
        client, _:= rpc.Dial("http://123.126.106.82:38555")
	defer client.Close()
        from :="P1LxMi9Lu1aaf6GXg63iJESruk6eVxjDhE2"
        to   :="P17K7gWSoSDhJW6zhGdcyJJqNNXiGuLTtmV"
	addr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"
        amount ,_ := decimal.NewFromString("10")
        fee,_ := decimal.NewFromString("1")
	result, err := client.Contract_Ccstop(context.Background(), from,to,amount,fee,addr)
	if err != nil {
                fmt.Println(err)
		//t.Error("TestSimpleContractCcstop No Pass")
                t.Log("Pass")
	}
	fmt.Println("TestSimpleContractCcstop",result)
        t.Log("Pass")
}

func TestSimpleContractCcquery(t *testing.T) {
    input := []string{"getTokenInfo", "btc"}
    client, _:= rpc.Dial("http://123.126.106.82:8545")
	defer client.Close()
	addr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43"
	result, err := client.Contract_CcQuery(context.Background(),addr,input,10)
	if err != nil {
		t.Error("TestSimpleContractCcquery No Pass")
	}
        fmt.Println(result)
        t.Log("Pass")
}

func TestSimpleContractCcinvoke(t *testing.T) {
    input := []string{"getTokenInfo", "btc"}
    client, _:= rpc.Dial("http://123.126.106.82:8545")
	defer client.Close()
	addr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43"
	result, err := client.Contract_Ccinvoke(context.Background(),addr,input)
	if err != nil {
		t.Error("TestSimpleContractCcinvoke No Pass")
	}
    fmt.Println(result)
    t.Log("Pass")
}

func TestSimpleContractCcinstall(t *testing.T) {
    addr := []string{"", ""}
    client, _:= rpc.Dial("http://123.126.106.82:8545")
	defer client.Close()
	from :="P1BbTByTVxG4GTRKUF3EdWdfjib2NzJvtSe"
	to   :="P1MU8eCfXBX9meTAy5UZ5wuCh6E9zm5TG3e"
	amount,_:=decimal.NewFromString("10")
	fee ,_:=decimal.NewFromString("0.5")
	tplName := "testPtnContract"
	path := "chaincode/testPtnContractTemplate"
	version := "ptn110"
	ccdescription := ""
	ccabi := ""
	cclanguage :="go"
	result, err := client.Contract_Ccinstall(context.Background(),from, to, amount,fee,tplName, path, version, ccdescription, ccabi, cclanguage,addr)
	if err != nil {
		//t.Error("TestSimpleContractCcinstall No Pass")
                t.Log("Pass")
	}
    fmt.Println(result)
    t.Log("Pass")
}
*/
