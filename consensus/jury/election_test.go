package jury

import (
	"testing"
	"math/rand"
	"time"
	"fmt"
	"math"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	alg "github.com/palletone/go-palletone/consensus/jury/vrf/algorithm"
)

//func createVrfCount() (*vrfAccount, error) {
//	c := elliptic.P256()
//	key, err := ecdsa.GenerateKey(c, crand.Reader)
//	if err != nil {
//		log.Error("createVrfCount, GenerateKey fail")
//		return nil, errors.New("GenerateKey fail")
//	}
//	va := vrfAccount{
//		priKey: key,
//		pubKey: &key.PublicKey,
//	}
//	return &va, nil
//}

//func electionOnce(index int, ks *keystore.KeyStore) {
//	reqId := util.RlpHash(util.IntToBytes(index))
//	//log.Info("electionOnce", "index", index, "id_hash", reqId)
//	_, err := createVrfCount()
//	if err != nil {
//		log.Error("electionOnce", "createVrfCount fail", err, "index", index)
//		return
//	}
//	ele := elector{
//		num:    4,
//		weight: 1,
//		total:  100,
//		//vrfAct: *va,
//
//		password: "1",
//		ks:       ks,
//	}
//	acc, err := ele.ks.NewAccount(ele.password)
//	if err != nil {
//		log.Error("electionOnce", "NewAccount fail, index", index)
//	}
//	ele.addr = acc.Address
//	ele.ks.Unlock(acc, "1")
//
//	//h := util.RlpHash(acc)
//	//log.Info("electionOnce", "account hash", h)
//
//	seedData:= getElectionSeedData(reqId)
//	proof, err := ele.checkElected(seedData)
//	if err != nil {
//		log.Error("electionOnce", "checkElected fail", err, "index", index)
//		return
//	}
//	//log.Info("electionOnce", "index", index, "seedData", seedData)
//
//	if proof != nil {
//		pk, err := ele.ks.GetPublicKey(ele.addr)
//		if err != nil {
//			log.Error("electionOnce GetPublicKey ", "error", err)
//			return
//		}
//		ok, err := ele.verifyVrf(proof, seedData, pk)
//		if err != nil {
//			log.Error("electionOnce", "verifyVRF fail", err, "index", index)
//			return
//		}
//		if ok {
//			log.Info("electionOnce, election Ok", "index", index)
//		}
//	}
//	//log.Info("electionOnce, election Fail", "index", index)
//}

//func TestElection(t *testing.T) {
//	rand.Seed(time.Now().UnixNano())
//	dir := filepath.Join(os.TempDir(), fmt.Sprintf("gptn-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))
//
//	os.MkdirAll(dir, 0700)
//	ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)
//
//	for i := 0; i < 100; i++ {
//		electionOnce(i, ks)
//	}
//}

func TestContractProcess(t *testing.T) {
	for i := 0; i < 10; i++ {
		reqId := util.RlpHash(util.IntToBytes(rand.Int()))
		//sel := alg.Selected(40, 4, 100, reqId[:])
		sel := alg.Selected(4, 10, 100, reqId[:])
		if sel > 0 {
			log.Info("sel ok")
		} else {
			fmt.Print(".")
			//log.Info("TestContractProcess", "Selected", sel, "idx", i, "reqId", reqId)
		}
	}
}

func eone(expectNm uint, weight, total uint64) (num int) {
	cnt := 0
	rand.Seed(time.Now().Unix())
	for i := 0; i < int(total); i++ {
		reqId := util.RlpHash(util.IntToBytes(rand.Int()))
		//	log.Info("election", "reqId", reqId)

		sel := alg.Selected(expectNm, weight, total, reqId[:])
		if sel > 0 {
			//log.Info("sel ok")
			cnt ++
		} else {
			//fmt.Print(".")
			//log.Info("TestContractProcess", "Selected", sel, "idx", i, "reqId", reqId)
		}
	}
	return cnt
}

/*
Preliminary test conclusion
expectNum:4
use:
total    weight    num
20         4        5
50         7        7
100        8        5
200        15       5
500        17       6
*/

func Test_Election_Optimal(t *testing.T) {
	order_weight := 1 //+1
	def_expNm := uint(4)
	def_weight := 1
	max_weight := 20

	total_list := []uint64{20, 50, 100} //20, 50, 100, 200, 500, 1000

	for _, total := range total_list {
		//log.Info("Election", "total", total)
		mixCnt := 100000
		for w := def_weight; w <= max_weight; w += order_weight {
			cnt := eone(def_expNm, uint64(w), total)
			if cnt > int(def_expNm) {
				tp := math.Abs(float64(cnt) - float64(def_expNm))
				if int(tp) <= mixCnt {
					mixCnt = int(tp)
					log.Info("Election", "total", total, "weight", w, "mixCnt", mixCnt, "cnt", int(cnt))
				}else{
					//log.Info("Election", "total", total, "weight", w, "cnt", int(cnt))
				}
			}
		}
	}
}
