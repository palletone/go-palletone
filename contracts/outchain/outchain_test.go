package outchain

import (
	"fmt"
	"testing"
)

func TestGetJuryBTCPrikeyTest(t *testing.T) {
	str, _ := GetJuryBTCPrikeyTest("sample_syscc")
	fmt.Println(str)
}
