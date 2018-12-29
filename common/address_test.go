/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package common

import (
	"github.com/stretchr/testify/assert"
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
func TestHexToAddrString(t *testing.T) {

	addr := HexToAddress("0x00000000000000000000000000000000000000011C")
	t.Logf("0x1 contract address: %s", addr.String()) //PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM
	t.Logf("Is system contract:%t", addr.IsSystemContractAddress())
	newAddr, _ := StringToAddress(addr.String())
	t.Logf("contract hex is: %x", newAddr.Bytes())
	addr = HexToAddress("0x00000000000000000000000000000000000000021C")
	t.Logf("0x2 contract address: %s", addr.String()) //PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
	t.Logf("Is system contract:%t", addr.IsSystemContractAddress())
	addr = HexToAddress("0x00000000000000000000000000000000000000031C")
	t.Logf("0x3 contract address: %s", addr.String()) //PCGTta3M4t3yXu8uRgkKvaWd2d8DRLGbeyd
	t.Logf("Is system contract:%t", addr.IsSystemContractAddress())

	addr = HexToAddress("0x00000000000000000000000000000000000000081C")
	t.Logf("0x8 contract address: %s", addr.String()) //PCGTta3M4t3yXu8uRgkKvaWd2d8DRLGbeyd
	t.Logf("Is system contract:%t", addr.IsSystemContractAddress())

	addr = HexToAddress("0x00000000000000000000000000000000000095271C")
	t.Logf("0x9 contract address: %s", addr.String()) //PCGTta3M4t3yXu8uRgkKvaWd2d9Vgsc4zGX
	t.Logf("Is system contract:%t", addr.IsSystemContractAddress())

	addr = HexToAddress("0x000000000000000000000000000000000000000100")
	t.Logf("0x1 user address: %s", addr.String())

	addr = HexToAddress("0x000000000000000000000000000000000000000105")
	t.Logf("0x1 p2sh address: %s", addr.String())

	addr = HexToAddress("0x3c5a9cd1dc2437342692de6ed2b948c5cbb3174800")
	t.Logf("0x1 p2sh address: %s", addr.String())

	t.Logf("Is system contract:%t", addr.IsSystemContractAddress())
}
func TestBytesListToAddressList(t *testing.T) {
	str := "[\"PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM\",\"P1LWaK3KBCuPVsXUPHXkMZr2Cm5tZquRDK8\"]"
	addressList := BytesListToAddressList([]byte(str))
	assert.True(t, len(addressList) == 2)
	for _, addr := range addressList {
		t.Logf("Address:%s", addr.String())
	}
}
