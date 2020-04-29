package packetcc

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"testing"
)

func TestPacket_GetPullAmount(t *testing.T) {
	p := &Packet{
		MinPacketAmount: 1,
		MaxPacketAmount: 100,
	}
	total := uint64(0)
	for i := 0; i < 1000; i++ {
		result := p.GetPullAmount(int64(i), 2000, 100)
		total += result
		t.Log(result)
	}
	t.Logf("Total:%d", total)

}
func TestNormalRandom(t *testing.T) {
	sum := uint64(0)
	for i := int64(0); i < 100; i++ {
		x := NormalRandom(i, 100, 90, 210)
		t.Log(x)
		sum += x
	}
	t.Logf("Total:%d", sum)
}

func TestOle2New(t *testing.T) {
	/*	// packet
		oldP := &OldPacket{
			PubKey:          []byte("old"),
			Creator:         common.Address{},
			Token:           dagconfig.DagConfig.GetGasToken().ToAsset(),
			Amount:          100,
			Count:           10,
			MinPacketAmount: 1,
			MaxPacketAmount: 10,
			ExpiredTime:     0,
			Remark:          "oldTest",
			Constant:        false,
		}
		fmt.Printf("old = %v\n",oldP)
		newP := OldPacket2New(oldP)
		fmt.Printf("new = %v\n",newP)
		fmt.Printf("new = %v\n",newP.Tokens)

		// record
		oldR := &OldPacketAllocationRecord{
			PubKey:      []byte("old"),
			Message:     "1",
			Amount:      1,
			Token:       dagconfig.DagConfig.GetGasToken().ToAsset(),
			ToAddress:   common.Address{},
			RequestHash: common.Hash{},
			Timestamp:   0,
		}
		fmt.Printf("old record = %v\n",oldR)
		newR := OldRecord2New(oldR)
		fmt.Printf("new record = %v\n",newR)
		fmt.Printf("new record = %v\n",newR.Tokens)*/

	//
	a, _ := common.StringToAddress("P16JiQ3U23zqGmpAhBZwH7gDksBz4ySzLT2")
	oP := &OldPacket{
		PubKey:          []byte("old"),
		Creator:         a,
		Token:           dagconfig.DagConfig.GetGasToken().ToAsset(),
		Amount:          90,
		Count:           10,
		MinPacketAmount: 1,
		MaxPacketAmount: 10,
		ExpiredTime:     0,
		Remark:          "remark",
		Constant:        false,
	}
	fmt.Printf("old = %v\n", oP)
	nP := OldPacket2New(oP,uint64(90),uint32(10))
	fmt.Printf("new = %v\n", nP)
	fmt.Printf("new = %#v\n", nP.Tokens[0])

	toA, _ := common.StringToAddress("P1CLYFd65W1YcHBmEBnUf8ACZsaqqworsMG")
	oR := &OldPacketAllocationRecord{
		PubKey:      []byte("old"),
		Message:     "1",
		Amount:      9,
		Token:       dagconfig.DagConfig.GetGasToken().ToAsset(),
		ToAddress:   toA,
		RequestHash: common.HexToHash("old"),
		Timestamp:   1588129287,
	}
	fmt.Printf("old = %v\n", oR)
	nR := OldRecord2New(oR)
	fmt.Printf("new = %v\n", nR)
	fmt.Printf("new = %#v\n", nR.Tokens[0])
}
