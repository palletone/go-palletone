package common

import (
	"testing"
)

func TestAddressValidate(t *testing.T) {
	p2pkh := "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"
	addr, err := StringToAddress(p2pkh)

	if err != nil {
		t.Error(err)
	}
	t.Log(addr)
}
func TestAddressNotValidate(t *testing.T) {
	p2pkh := "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gj"
	addr, err := StringToAddress(p2pkh)

	if err != nil {
		t.Log(addr)
		t.Log(err)
	} else {
		t.Error("It must invalid, but pass, please check your code")
	}

}
