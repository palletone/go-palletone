package outchain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/palletone/adaptor"
)

func TestGetJuryKeyInfo(t *testing.T) {
	input := adaptor.NewPrivateKeyInput{}
	inputJSON, err := json.Marshal(&input)
	fmt.Println(string(inputJSON))

	key, err := GetJuryKeyInfo("sample_syscc", "erc20", inputJSON, GetERC20Adaptor())
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("prikey: %x\n", key)
	}
}
