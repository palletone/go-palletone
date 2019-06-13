package jury

import (
	"fmt"
	"testing"

	"github.com/palletone/go-palletone/common"
)

func TestSortSigs(t *testing.T) {
	//pubkeyStrs := []string{"03d79b178131ab8685b7551f0f6dad920a6eadc8a4addcb1ce59632ed128a8e770",
	//	"02ee43ae1e6b86f0a01d97f9d483e03cd6b4e5e0772d310159de183193fefc6892",
	//}
	//sigStrs := []string{"a2f880d5520e128fc00f2265da2c5830d6df7e6c9d62a1ee9b3e47695c630d113b9349cede0b12ded199a8fdb49791499e27fa6e700c4028f8459e4fe938bb0d00",
	//	"1f9d8bb450ccd6b3b7e0d3348a21845ac324d926f1ffd71e7c54615787d462e74d6f9611eade3f78a371b3561dc4af303e39ff4da3c1f67d7f0b61da5049237201",
	//}

	pubkeyStrs := []string{"02ee43ae1e6b86f0a01d97f9d483e03cd6b4e5e0772d310159de183193fefc6892",
		"03d79b178131ab8685b7551f0f6dad920a6eadc8a4addcb1ce59632ed128a8e770",
	}
	sigStrs := []string{"1f9d8bb450ccd6b3b7e0d3348a21845ac324d926f1ffd71e7c54615787d462e74d6f9611eade3f78a371b3561dc4af303e39ff4da3c1f67d7f0b61da5049237201",
		"a2f880d5520e128fc00f2265da2c5830d6df7e6c9d62a1ee9b3e47695c630d113b9349cede0b12ded199a8fdb49791499e27fa6e700c4028f8459e4fe938bb0d00",
	}

	redeemStr := "522103d79b178131ab8685b7551f0f6dad920a6eadc8a4addcb1ce59632ed128a8e7702102ee43ae1e6b86f0a01d97f9d483e03cd6b4e5e0772d310159de183193fefc689252ae"

	pubkys := [][]byte{}
	for i := range pubkeyStrs {
		pubkys = append(pubkys, common.Hex2Bytes(pubkeyStrs[i]))
	}
	sigs := [][]byte{}
	for i := range sigStrs {
		sigs = append(sigs, common.Hex2Bytes(sigStrs[i]))
	}
	redeem := common.Hex2Bytes(redeemStr)

	sigOrder := SortSigs(pubkys, sigs, redeem)

	//sigOrder is sorted by redeem's Pubkey
	for i := range sigOrder {
		fmt.Printf("%d %x\n", i, sigOrder[i])
	}
}

func TestDeleOneMax(t *testing.T) {
	sigStrs := []string{"1f9d8bb450ccd6b3b7e0d3348a21845ac324d926f1ffd71e7c54615787d462e74d6f9611eade3f78a371b3561dc4af303e39ff4da3c1f67d7f0b61da5049237201",
		"a2f880d5520e128fc00f2265da2c5830d6df7e6c9d62a1ee9b3e47695c630d113b9349cede0b12ded199a8fdb49791499e27fa6e700c4028f8459e4fe938bb0d00",
		"1f9d8bb450ccd6b3b7e0d3348a21845ac324d926f1ffd71e7c54615787d462e74d6f9611eade3f78a371b3561dc4af303e39ff4da3c1f67d7f0b61da5049237201",
		"a2f880d5520e128fc00f2265da2c5830d6df7e6c9d62a1ee9b3e47695c630d113b9349cede0b12ded199a8fdb49791499e27fa6e700c4028f8459e4fe938bb0d00",
	}
	sigs := [][]byte{}
	for i := range sigStrs {
		sigs = append(sigs, common.Hex2Bytes(sigStrs[i]))
	}

	delNum := 2
	for delNum > 0 {
		sigs = DeleOneMax(sigs)
		delNum--
	}
	for i := range sigs {
		fmt.Printf("%d %x\n", i, sigs[i])
	}
}
