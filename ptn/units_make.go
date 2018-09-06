package ptn

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"log"
	"time"
)

func MakeDags(Memdb ptndb.Database, unitAccount int) (*dag.Dag, error) {
	dag, _ := dag.NewDagForTest(Memdb)
	log.Println("开始创建 genesis unit===》》》")
	header := NewHeader([]common.Hash{}, []modules.IDType16{modules.PTNCOIN}, []byte{})
	header.Number.AssetID = modules.PTNCOIN
	header.Number.IsMain = true
	header.Number.Index = 0
	header.Authors = &modules.Authentifier{"", []byte{}, []byte{}, []byte{}}
	header.Witness = []*modules.Authentifier{&modules.Authentifier{"", []byte{}, []byte{}, []byte{}}}
	tx, _ := NewCoinbaseTransaction()
	txs := modules.Transactions{tx}
	genesisUnit := NewUnit(header, txs)
	err := SaveGenesis(dag.Db, genesisUnit)
	if err != nil {
		log.Println("SaveGenesis, err", err)
		return nil, err
	}
	log.Printf("--------genesis----unit----------------%#v\n", genesisUnit)
	log.Printf("--------genesis----unit.UnitHeader-----%#v\n", genesisUnit.UnitHeader)
	log.Printf("--------genesis----unit.Txs------------%#v\n", genesisUnit.Txs[0].Hash())
	log.Printf("--------genesis----unit.UnitHash-------%#v\n", genesisUnit.UnitHash)
	log.Printf("--------genesis----unit.UnitHeader.ParentsHash-----%#v\n", genesisUnit.UnitHeader.ParentsHash)
	log.Printf("--------genesis----unit.UnitHeader.Number.Index----%#v\n", genesisUnit.UnitHeader.Number.Index)
	log.Println("创建 genesis unit 完成并保存===》》》")
	log.Println()
	log.Println("开始创建其他 unit===》》》")
	units, _ := newDag(dag.Db, genesisUnit, unitAccount)
	log.Println("创建其他 unit 完成并保存===》》》")
	log.Println("全部unit的数量===》》》", len(units)+1)
	return dag, nil
}
func newDag(memdb ptndb.Database, gunit *modules.Unit, number int) (modules.Units, error) {
	units := make(modules.Units, number)
	par := gunit
	for i := 0; i < number; i++ {
		header := NewHeader([]common.Hash{par.UnitHash}, []modules.IDType16{modules.PTNCOIN}, []byte{})
		header.Number.AssetID = par.UnitHeader.Number.AssetID
		header.Number.IsMain = par.UnitHeader.Number.IsMain
		header.Number.Index = par.UnitHeader.Number.Index + 1
		header.Authors = &modules.Authentifier{"", []byte{}, []byte{}, []byte{}}
		header.Witness = []*modules.Authentifier{&modules.Authentifier{"", []byte{}, []byte{}, []byte{}}}
		tx, _ := NewCoinbaseTransaction()
		txs := modules.Transactions{tx}
		unit := NewUnit(header, txs)
		err := SaveUnit(memdb, unit, true)
		if err != nil {
			log.Println("保存其他unit出错===》》》", err)
			return nil, err
		}
		//log.Printf("--------第 %d 个unit====》》》----unit--------------%#v\n",i+1,unit)
		//log.Printf("--------第 %d 个unit====》》》----unit.UnitHeader---%#v\n",i+1,unit.UnitHeader)
		//log.Printf("--------第 %d 个unit====》》》----unit.Txs----------%#v\n",i+1,unit.Txs[0].Hash())
		//log.Printf("--------第 %d 个unit====》》》----unit.UnitHash-----%#v\n",i+1,unit.UnitHash)
		//log.Printf("--------第 %d 个unit====》》》----unit.UnitHeader.ParentsHash-----%#v\n",i+1,unit.UnitHeader.ParentsHash)
		//log.Printf("--------第 %d 个unit====》》》----unit.UnitHeader.Number.Index----%#v\n",i+1,unit.UnitHeader.Number.Index)
		units[i] = unit
		par = unit
	}
	return units, nil
}
func SaveGenesis(db ptndb.Database, unit *modules.Unit) error {
	if unit.NumberU64() != 0 {
		return fmt.Errorf("can't commit genesis unit with number > 0")
	}
	err := SaveUnit(db, unit, true)
	if err != nil {
		log.Println("SaveGenesis==", err)
		return err
	}
	return nil
}

func SaveUnit(db ptndb.Database, unit *modules.Unit, isGenesis bool) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Println("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	if unit.UnitSize != unit.Size() {
		log.Println("Validate size", "error", "Size is invalid")
		return modules.ErrUnit(-1)
	}
	_, isSuccess, err := common2.ValidateTransactions(db, &unit.Txs, isGenesis)
	if isSuccess != true {
		fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
		return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	}
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := storage.SaveHeader(db, unit.UnitHash, unit.UnitHeader); err != nil {
		log.Println("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step5. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	if err := storage.SaveNumberByHash(db, unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Println("SaveHashNumber:", "error", err.Error())
		return fmt.Errorf("Save unit hash and number error")
	}
	if err := storage.SaveHashByNumber(db, unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Println("SaveNumberByHash:", "error", err.Error())
		return fmt.Errorf("Save unit hash and number error")
	}
	if err := storage.SaveTxLookupEntry(db, unit); err != nil {
		return err
	}
	if err := storage.SaveTxLookupEntry(db, unit); err != nil {
		return err
	}
	if err := saveHashByIndex(db, unit.UnitHash, unit.UnitHeader.Number.Index); err != nil {
		return err
	}
	// update state
	storage.PutCanonicalHash(db, unit.UnitHash, unit.NumberU64())
	storage.PutHeadHeaderHash(db, unit.UnitHash)
	storage.PutHeadUnitHash(db, unit.UnitHash)
	storage.PutHeadFastUnitHash(db, unit.UnitHash)
	// todo send message to transaction pool to delete unit's transactions
	return nil
}
func NewUnit(header *modules.Header, txs modules.Transactions) *modules.Unit {
	u := &modules.Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	u.ReceivedAt = time.Now()
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}
func NewHeader(parents []common.Hash, asset []modules.IDType16, extra []byte) *modules.Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &modules.Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...), Creationdate: time.Now().Unix()}
}
func NewCoinbaseTransaction() (*modules.Transaction, error) {
	input := &modules.Input{}
	output := &modules.Output{}
	payload := modules.PaymentPayload{
		Input:  []*modules.Input{input},
		Output: []*modules.Output{output},
	}
	msg := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: payload,
	}
	coinbase := &modules.Transaction{
		TxMessages: []*modules.Message{&msg},
	}
	coinbase.TxHash = coinbase.Hash()
	return coinbase, nil
}

func saveHashByIndex(db ptndb.Database, hash common.Hash, index uint64) error {
	key := fmt.Sprintf("%s%v_", storage.HEADER_PREFIX, index)
	err := db.Put([]byte(key), hash.Bytes())
	return err
}
