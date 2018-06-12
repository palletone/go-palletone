package modules

import (
	"log"
	"testing"
)

func TestNewUnit(t *testing.T) {
	unit := NewUnit()
	log.Println("unit", unit)
	if unit.IsOnMainChain {
		log.Println("Is On Main Chain:", true)
	} else {
		log.Println("Is On Main Chain", false)
	}
}

// test interface
type USB interface {
	Name() string
	Connect()
}
type PhoncConnecter struct {
	name string
}

func (pc PhoncConnecter) Name() string {
	return pc.name
}
func (pc PhoncConnecter) Connect() {
	log.Println(pc.name)
}
func TestInteface(t *testing.T) {
	// 第一种直接在声明结构时赋值
	var a USB
	a = PhoncConnecter{"PhoneC"}
	a.Connect()
	// 第二种，先给结构赋值后在将值给接口去调用
	var b = PhoncConnecter{}
	b.name = "b"
	var c USB
	c = b
	c.Connect()
}
