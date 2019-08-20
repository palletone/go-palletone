package ethadaptor

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/palletone/adaptor"
)

func TestNewPrivateKey(t *testing.T) {
	key, _ := NewPrivateKey(NETID_TEST)
	fmt.Printf("%x\n", key)
}

func TestGetPublicKey(t *testing.T) {
	//	key := "5f3e017176835cee1ac8c90b702e9a45802784e3bce0404908f8d5ceaec97cad"
	//	testPubkey := "036b86475563e9c78c85ee8d4c007ff419b87968c3573f3d3405fd2a56e2f09679"
	keyHex := "8e87ebb3b00565aaf3675e1f7d16ed68b300c7302267934f3831105b48e8a3e7"
	testPubkey := "021c183161f5d96f59d6078d0123021876b5a0982b131ffa021b4437f49b93588a"

	key := Hex2Bytes(keyHex)
	pubkey, _ := GetPublicKey(key)
	pubkeyHex := hex.EncodeToString(pubkey)
	if testPubkey != pubkeyHex {
		t.Errorf("unexpected pubkey bytes - got: %s, "+
			"want: %s", pubkey, testPubkey)
	}
}

func TestGetAddress(t *testing.T) {
	//	key := "5f3e017176835cee1ac8c90b702e9a45802784e3bce0404908f8d5ceaec97cad"
	//	testAddr := "0xAc0a2917Dc6722a4554ea6C87ff0576105d7E26f"
	keyHex := "8e87ebb3b00565aaf3675e1f7d16ed68b300c7302267934f3831105b48e8a3e7"
	testAddr := "0x7d7116a8706ae08baa7f4909e26728fa7a5f0365"

	key := Hex2Bytes(keyHex)
	addr, _ := GetAddress(key, NETID_TEST)
	addrLower := strings.ToLower(addr)
	if addrLower != testAddr {
		t.Errorf("unexpected address - got: %s, "+
			"want: %s", addrLower, testAddr)
	}
}

func TestGetPalletOneMappingAddress(t *testing.T) {
	var input adaptor.GetPalletOneMappingAddressInput
	input.ChainAddress = "0x1a9ed32dec553511158595375d62a8aa8784bc5b"
	result, _ := GetPalletOneMappingAddress(&input)
	fmt.Println(result.PalletOneAddress)
}
