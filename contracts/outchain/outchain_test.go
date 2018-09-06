package outchain

import (
	"fmt"
	"testing"
)

func TestGetJuryBTCPrikeyTest(t *testing.T) {
	str, err := GetJuryBTCPrikeyTest("sample_syscc")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(str)
	}
}
