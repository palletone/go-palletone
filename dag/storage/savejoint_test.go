package storage

import (
	"errors"
	"log"
	"strconv"
	"testing"
	"time"

	"gitee.com/sailinst/pallet_dag/palletdb"
	"github.com/palletone/go-palletone/dag/config"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestSaveJoint(t *testing.T) {
	if IsGenesisUnit("123") {
		log.Println("faile")
		t.Error("faild")
	} else {
		log.Println("success")
		t.Log("success")
	}
	log.Println(strconv.FormatInt(time.Now().Unix(), 10))
	err := SaveJoint(&modules.Joint{Unit: &modules.Unit{Unit: "123", CreationDate: time.Now()}, Ball: "ball", Skiplist: []string{"s1", "s2", "s3"}, CreationDate: time.Now()},
		&modules.ValidationState{}, func() { log.Println("ok") })
	log.Println("error:", err)
}

func TestAddUnitKey(t *testing.T) {
	keys := GetUnitKeys()
	// if len(keys) <= 0 {
	// 	return errors.New("null keys.")
	// }
	keys = append(keys, "unit1231526522017", "unit1231526521834")
	var err error
	if Dbconn == nil {
		config.DConfig.DbPath = "/Users/jay/code/gocode/src/palletone/bin/leveldb"
		Dbconn, err = palletdb.NewLDBDatabase(config.DConfig.DbPath, 0, 0)
		if err != nil {
			log.Println("new db error", err)
			t.Fatal("error1")
		}
	}
	if err := Dbconn.Put([]byte("array_units"), ConvertBytes(keys)); err != nil {
		log.Println("put error", err)
		t.Fatal("error2")
	}
	log.Println("success")
}

func TestGetUnitKeys(t *testing.T) {
	t0 := time.Now()

	keys := GetUnitKeys()
	var this []string
	for i, v := range keys {
		var exist bool
		for j := i + 1; j < len(keys); j++ {
			if v == keys[j] {
				log.Println("j:", j)
				exist = true
				log.Println("equal", v)
				break
			}
		}
		if !exist {
			// log.Println("i:", i)
			this = append(this, v)
		}
	}

	err := AddUnitKeys("unit1231526521834")
	if errors.New("key is already exist.").Error() == err.Error() {
		log.Println("success test add unit") // this
	} else {
		log.Println("failed test add  unit ")
	}
	log.Println("times:", (time.Now().UnixNano()-t0.UnixNano())/1e6)
}

func TestDBBatch(t *testing.T) {
	log.Println("db_path:", DBPath)
	table := palletdb.NewTable(Dbconn, "hehe")
	err0 := table.Put([]byte("jay"), []byte("baby"))
	log.Println("err0:", err0)

	b, err := table.Get([]byte("jay"))
	log.Println("b:", string(b), err)

	log.Println("table:", table.prefix)
}
