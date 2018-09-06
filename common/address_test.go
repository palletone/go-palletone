package common

import (
	"testing"
)

func TestAddressValidate(t *testing.T) {
	p2pkh := "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"
	addr := StringToAddress(p2pkh)
	addrt, err := addr.Validate()
	if err != nil {
		t.Error(err)
	}
	t.Log(addrt)
}
func TestAddressNotValidate(t *testing.T) {
	p2pkh := "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gj"
	addr := StringToAddress(p2pkh)
	addrt, err := addr.Validate()
	if err != nil {
		t.Log(addrt)
		t.Log(err)
	} else {
		t.Error("It must invalid, but pass, please check your code")
	}

}
