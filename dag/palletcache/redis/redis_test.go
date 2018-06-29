package redis

import (
	"log"
	"testing"
)

func TestInit(t *testing.T) {
	log.Println("start test redis.")
	Init()
	// log.Println(red.ParseConfig("jay"), red.Init())
	Store("unit", "unit1", "hello1")

	val, ok := Get("unit", "123")
	log.Println("val:", (val), ok)

	if val, ok := GetString("unit", "unit1"); ok {
		log.Println("val", val)
	}
}
