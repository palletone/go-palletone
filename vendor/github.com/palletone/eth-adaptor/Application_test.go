package ethadaptor

import (
	"encoding/hex"
	"github.com/palletone/adaptor"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	j1prvKey, _ = hex.DecodeString("e91c54bb8b68b19b77897fb32896b0bf31db76026ad02684612f7f7dfaeaae64")
	j1pubKey, _ = GetPublicKey(j1prvKey)
	j1Addr, _   = PubKeyToAddress(j1pubKey)
	j2prvKey, _ = hex.DecodeString("2f5a4e4d8f80c1a8069800d402a4bd17641ea1b4d2d20f4558016ca39b7ebbe5")
	j2pubKey, _ = GetPublicKey(j2prvKey)
	j2Addr, _   = PubKeyToAddress(j2pubKey)
	j3prvKey, _ = hex.DecodeString("59004d4b0b5a1f8e8bc75df36f337d21d8055f183ebf5db543de18729d869d1b")
	j3pubKey, _ = GetPublicKey(j3prvKey)
	j3Addr, _   = PubKeyToAddress(j3pubKey)
	j4prvKey, _ = hex.DecodeString("e128bfee7ca58ab329a1ec14892ec80f0ff563f4b8829b2f806c4412231c113c")
	j4pubKey, _ = GetPublicKey(j4prvKey)
	j4Addr, _   = PubKeyToAddress(j4pubKey)
	u1EthAddr   = "0x588eB98f8814aedB056D549C0bafD5Ef4963069C"
	ptnAddr     = "P1KJSodB2vzJ1A7jyqWVGb9NN7pCSt9ZZnN" //0xc8bee01c60a428f337406f9d10846d5b71c226e1
)

func TestDepositETH(t *testing.T) {
	var aeth adaptor.ICryptoCurrency = NewAdaptorETHTestnet()
	multiSignAddr:="0x4d736ed88459b2db85472aab13a9d0ce2a6ea676"
	input := &adaptor.CreateMultiSigAddressInput{}
	input.Keys = make([][]byte, 4)
	input.Keys[0] = j1pubKey
	input.Keys[1] = j2pubKey
	input.Keys[2] = j3pubKey
	input.Keys[3] = j4pubKey
	input.SignCount = 3
	_, err := aeth.CreateMultiSigAddress(input)
	t.Logf("Jury pub keys:%x", input.Keys)
	t.Logf("Jury addresss:%s", [...]string{j1Addr, j2Addr, j3Addr, j4Addr})
	assert.NotNil(t, err)

	t.Logf("MutiSign Address:%s", multiSignAddr)
	addrOut, err := aeth.GetPalletOneMappingAddress(
		&adaptor.GetPalletOneMappingAddressInput{
			ChainAddress: u1EthAddr,
		MappingDataSource:multiSignAddr})
	assert.Nil(t, err)
	t.Logf("PalletOne Address:%s,%x", addrOut.PalletOneAddress, []byte(addrOut.PalletOneAddress))
	//User1通过自己的ETH钱包转账到多签地址
	//接下来申请提PETH
	txHistoryOut, err := aeth.GetAddrTxHistory(&adaptor.GetAddrTxHistoryInput{FromAddress: u1EthAddr, ToAddress: multiSignAddr, PageSize: 5, AddressLogicAndOr: true, Asset: "ETH"})
	assert.Nil(t, err)
	for _, txHist := range txHistoryOut.Txs {
		t.Logf("History tx:%v", txHist.String())
		if txHist.IsInBlock && txHist.IsStable && txHist.IsSuccess {
			t.Logf("用户%s充值:%s对应Txid：%x,可以发放PETH", txHist.FromAddress, txHist.Amount.String(), txHist.TxID)
		}
	}
}

func TestDepositErc20(t *testing.T) {
	var aeth adaptor.ICryptoCurrency = NewAdaptorErc20Testnet()
	erc20Asset := "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9"
	multiSignAddr:="0x4d736ed88459b2db85472aab13a9d0ce2a6ea676"
	input := &adaptor.CreateMultiSigAddressInput{}
	input.Keys = make([][]byte, 4)
	input.Keys[0] = j1pubKey
	input.Keys[1] = j2pubKey
	input.Keys[2] = j3pubKey
	input.Keys[3] = j4pubKey
	input.SignCount = 3
	_, err := aeth.CreateMultiSigAddress(input)
	t.Logf("Jury pub keys:%x", input.Keys)
	t.Logf("Jury addresss:%s", [...]string{j1Addr, j2Addr, j3Addr, j4Addr})
	assert.NotNil(t, err)
	//multiSignAddr := output.Address
	t.Logf("MutiSign Address:%s", multiSignAddr)
	addrOut, err := aeth.GetPalletOneMappingAddress(
		&adaptor.GetPalletOneMappingAddressInput{
			ChainAddress: u1EthAddr,
		MappingDataSource:multiSignAddr})
	assert.Nil(t, err)
	t.Logf("PalletOne Address:%s,%x", addrOut.PalletOneAddress, []byte(addrOut.PalletOneAddress))
	//User1通过自己的ETH钱包转账到多签地址
	//接下来申请提PETH
	txHistoryOut, err := aeth.GetAddrTxHistory(&adaptor.GetAddrTxHistoryInput{FromAddress: u1EthAddr, ToAddress: multiSignAddr, PageSize: 5, AddressLogicAndOr: true, Asset: erc20Asset})
	assert.Nil(t, err)
	for _, txHist := range txHistoryOut.Txs {
		t.Logf("History tx:%v", txHist.String())
		if txHist.IsInBlock && txHist.IsStable && txHist.IsSuccess {
			t.Logf("用户%s充值:%s对应Txid：%x,可以发放ERC20", txHist.FromAddress, txHist.Amount.String(), txHist.TxID)
		}
	}
}
