package packetcc

import (
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
