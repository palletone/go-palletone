package jury

import (
	"testing"
	"math/rand"
	"time"
	"path/filepath"
	"os"
	"fmt"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	alg "github.com/palletone/go-palletone/consensus/jury/algorithm"
)

func createVrfCount() (*vrfAccount, error) {
	c := elliptic.P256()
	key, err := ecdsa.GenerateKey(c, crand.Reader)
	if err != nil {
		log.Error("createVrfCount, GenerateKey fail")
		return nil, errors.New("GenerateKey fail")
	}
	va := vrfAccount{
		priKey: key,
		pubKey: &key.PublicKey,
	}
	return &va, nil
}

func electionOnce(index int, ks *keystore.KeyStore) {
	reqId := util.RlpHash(util.IntToBytes(index))
	//log.Info("electionOnce", "index", index, "id_hash", reqId)
	va, err := createVrfCount()
	if err != nil {
		log.Error("electionOnce", "createVrfCount fail", err, "index", index)
		return
	}
	ele := elector{
		num:    20,
		weight: 4,
		total:  100,
		vrfAct: *va,

		password: "1",
		ks:       ks,
	}
	acc, err := ele.ks.NewAccount(ele.password)
	if err != nil {
		log.Error("electionOnce", "NewAccount fail, index", index)
	}
	ele.addr = acc.Address
	ele.ks.Unlock(acc, "1")

	//h := util.RlpHash(acc)
	//log.Info("electionOnce", "account hash", h)

	seedData, err := getElectionSeedData(reqId)
	if err != nil {
		log.Error("electionOnce", "getElectionSeedData fail", err, "index", index)
		return
	}

	proof, err := ele.checkElected(seedData)
	if err != nil {
		log.Error("electionOnce", "checkElected fail", err, "index", index)
		return
	}
	//log.Info("electionOnce", "index", index, "seedData", seedData)

	if proof != nil {
		pk, err := ele.ks.GetPublicKey(ele.addr)
		if err != nil{
			log.Error("electionOnce GetPublicKey ", "error", err)
			return
		}
		ok, err := ele.verifyVrf(proof, seedData, pk)
		if err != nil {
			log.Error("electionOnce", "verifyVRF fail", err, "index", index)
			return
		}
		if ok {
			log.Info("electionOnce, election Ok", "index", index)
		}
	}
	//log.Info("electionOnce, election Fail", "index", index)
}

func TestElection(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("gptn-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))

	os.MkdirAll(dir, 0700)
	ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

	for i := 0; i < 100; i++ {
		electionOnce(i, ks)
	}
}

func TestContractProcess(t *testing.T) {
	for i := 0; i < 100; i++ {
		reqId := util.RlpHash(util.IntToBytes(rand.Int()))
		sel := alg.Selected(40, 4, 100, reqId[:])
		log.Info("TestContractProcess", "Selected", sel, "idx", i,  "reqId", reqId)
	}

}
