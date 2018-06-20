package getor

import (
	"log"
	"testing"
)

func TestGet(t *testing.T) {
	if m := GetPrefix([]byte("array")); m != nil {
		for k, v := range m {
			log.Println("key: ", k, "value: ", string(v))
		}
	}

	if m := GetPrefix([]byte("20")); m != nil {
		for k, v := range m {
			log.Println("key: ", k, "value: ", string(v))
		}
	}

	if m := GetPrefix([]byte("unit")); m != nil {
		for k, _ := range m {
			log.Println("key: ", k, "value: ", string(m[k]))
		}
	}
}
