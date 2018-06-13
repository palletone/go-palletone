package util

import (
	"log"
	"testing"
)

func TestGetSourceString(t *testing.T) {
	log.Println("bool", GetSourceString(true))
	log.Println("int64", GetSourceString(int64(123)))
	log.Println("float", GetSourceString(123.182318278))

	log.Println("[]", GetSourceString([]string{"1", "2", "3"}))
	var mapobj map[string]string
	mapobj = make(map[string]string)
	mapobj["lan"] = "22"
	mapobj["jay"] = "23"
	mapobj["ma"] = "24"
	log.Println("map", GetSourceString(mapobj))
	type obj struct {
		Name string
		Age  int
	}
	o := obj{
		Name: "jay",
		Age:  20,
	}

	arr := GetSourceString(o)
	log.Println("ok object", arr)
}
