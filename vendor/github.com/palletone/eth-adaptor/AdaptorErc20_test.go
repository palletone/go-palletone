package ethadaptor

import (
	"fmt"
	"testing"

	"github.com/palletone/adaptor"
	"github.com/stretchr/testify/assert"
)

func TestAdaptorErc20_NewPrivateKey(t *testing.T) {
	ada := newTestAdaptorErc20()
	output, err := ada.NewPrivateKey(nil)
	assert.Nil(t, err)
	t.Logf("New private key:%x,len:%d", output.PrivateKey, len(output.PrivateKey))
	pubKeyOutput, err := ada.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: output.PrivateKey})
	assert.Nil(t, err)
	t.Logf("Pub key:%x,len:%d", pubKeyOutput.PublicKey, len(pubKeyOutput.PublicKey))
	addrOutput, err := ada.GetAddress(&adaptor.GetAddressInput{Key: pubKeyOutput.PublicKey})
	assert.Nil(t, err)
	t.Logf("Address:%s", addrOutput.Address)
}
func newTestAdaptorErc20() *AdaptorErc20 {
	rpcParams := RPCParams{
		Rawurl: "https://ropsten.infura.io/", //"\\\\.\\pipe\\geth.ipc",//0xfb686ccee357012b8b8f338f8266a472f3c211c82f0a4c30a5d2e51176556546
	}
	return NewAdaptorErc20(NETID_TEST, rpcParams)
}

func TestGetAssetDecimal(t *testing.T) {
	ada := newTestAdaptorErc20()
	asset := &adaptor.GetAssetDecimalInput{Asset: "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9"}
	output, err := ada.GetAssetDecimal(asset)
	assert.Nil(t, err)
	fmt.Println(output.Decimal)
}

func TestAdaptorErc20_GetPalletOneMappingAddress(t *testing.T) {
	rpcParams := RPCParams{
		Rawurl: "https://ropsten.infura.io/", //"\\\\.\\pipe\\geth.ipc",//0xfb686ccee357012b8b8f338f8266a472f3c211c82f0a4c30a5d2e51176556546
	}

	ada := NewAdaptorErc20(NETID_TEST, rpcParams)//"0xa840d94b1ef4c326c370e84d108d539d31d52e84"

	addrETH := &adaptor.GetPalletOneMappingAddressInput{
		ChainAddress: "0x7D7116A8706Ae08bAA7F4909e26728fa7A5f0365",
		MappingDataSource:"0xa840d94b1ef4c326c370e84d108d539d31d52e84"}
	outputPTN, err := ada.GetPalletOneMappingAddress(addrETH)
	assert.Nil(t, err)
	t.Logf("Map PalletOne Address:%s", outputPTN.PalletOneAddress)

	addrPTNHex := &adaptor.GetPalletOneMappingAddressInput{
		PalletOneAddress: "P124gB1bXHDTXmox58g4hd4u13HV3e5vKie",
	MappingDataSource:"0xa840d94b1ef4c326c370e84d108d539d31d52e84"}
	outputETH, err := ada.GetPalletOneMappingAddress(addrPTNHex)
	assert.Nil(t, err)
	t.Logf("Map Chain Address:%s", outputETH.ChainAddress)
}

func TestAdaptorErc20_GetBalance(t *testing.T) {
	ada := newTestAdaptorErc20()

	input := &adaptor.GetBalanceInput{Address: "0x7D7116A8706Ae08bAA7F4909e26728fa7A5f0365", Asset: "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9"}
	output, err := ada.GetBalance(input)
	assert.Nil(t, err)
	fmt.Println(output.Balance.Amount.String())
}
