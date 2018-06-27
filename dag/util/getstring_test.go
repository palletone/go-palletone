package util

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"unicode/utf8"
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
	log.Println("ok object", arr, f2(s))
}

func f1(s string) int {
	return bytes.Count([]byte(s), nil) - 1
}

func f2(s string) int {
	return strings.Count(s, "") - 1
}

func f3(s string) int {
	return len([]rune(s))
}

func f4(s string) int {
	return utf8.RuneCountInString(s)
}

var s = "Hello, 世界"

func Benchmark1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f1(s)
	}
}

func Benchmark2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f2(s)
	}
}

func Benchmark3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f3(s)
	}
}

func Benchmark4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f4(s)
	}
}
