package ethadaptor

import (
	"fmt"
	"testing"

	"github.com/palletone/adaptor"
)

func TestGetBalance(t *testing.T) {
	params := &adaptor.GetBalanceInput{Address: "0x7d7116a8706ae08baa7f4909e26728fa7a5f0365", Asset: "eth"}
	rpcParams := RPCParams{
		Rawurl: "https://ropsten.infura.io/", //"\\\\.\\pipe\\geth.ipc",
	}
	result, err := GetBalanceETH(params, &rpcParams, NETID_MAIN)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result.Balance.Amount.String())
	}
}
