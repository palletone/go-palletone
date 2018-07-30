package common

import (
	"log"
	"testing"
)

func TestSaveConfig(t *testing.T) {
	confs := make(map[string]interface{})
	confs["ChainID"] = 1
	if err := SaveConfig(confs); err != nil {
		log.Println(err)
	}
}

func TestGetConfig(t *testing.T) {
	data := GetConfig([]byte("InitialActiveMediators"))
	if len(data) <= 0 {
		log.Println("Get config data error")
	} else {
		log.Println("Get Data:", data)
	}
}
