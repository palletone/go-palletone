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
import(
	"fmt"
	"context"
        "testing"
	"github.com/palletone/go-palletone"
	"github.com/palletone/go-palletone/common/rpc"
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
func TestSimpleContractCcstop(t *testing.T) {
    client, _:= rpc.Dial("http://127.0.0.1:8485")
	defer client.Close()
	addr := "P1PwFUG7ydvC1KhGsbyQzXCR8TEgdvx9Hut"
	result, err := client.Contract_Ccstop(context.Background(), addr)
	if err != nil {
		t.Error("TestSimpleContractCcstop No Pass")
	}
	fmt.Println("TestSimpleContractCcstop",result)
	t.Error("TestSimpleContractCcstop Pass")
}

func TestSimpleContractCcquery(t *testing.T) {
    input := []string{"a", "b"}
    client, _:= rpc.Dial("http://127.0.0.1:8485")
	defer client.Close()
	addr := "P1PwFUG7ydvC1KhGsbyQzXCR8TEgdvx9Hut"
	result, err := client.Contract_CcQuery(context.Background(), addr,input)
	if err != nil {
		t.Error("TestSimpleContractCcquery No Pass")
	}
	fmt.Println("TestSimpleContractCcquery",result)
	t.Error("TestSimpleContractCcquery Pass")
}
