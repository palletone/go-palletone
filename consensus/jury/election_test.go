package jury

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"github.com/palletone/go-palletone/dag/errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func electionOnce(index uint) {
	reqId := util.RlpHash(index)
	//log.Info("electionOnce", "index", index, "id_hash", reqId)
	va, err := createVrfCount()
	if err != nil {
		log.Error("electionOnce", "createVrfCount fail", err, "index", index)
		return
	}
	ele := elector{
		num:    10,
		weight: 4,
		total:  100,
		vrfAct: *va,
	}
	rand.Seed(time.Now().UnixNano())
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("gptn-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))
	ele.ks = keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)
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
		ok, err := ele.verifyVrfEc(proof, seedData)
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
	for i := 0; i < 100; i++ {
		electionOnce(uint(i))
	}
}
