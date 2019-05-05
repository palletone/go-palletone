package list

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/palletone/go-palletone/common/log"
	"github.com/pkg/errors"
)

//var log = flogging.MustGetLogger("cclist")

type CCInfo struct {
	Id      []byte
	Name    string
	Path    string
	Version string
	TempleId []byte
	SysCC  bool
	//Enable bool
}

type chain struct {
	Version int
	CClist  map[string]*CCInfo //chainCodeId
}

var chains = struct {
	mu    sync.Mutex
	Clist map[string]*chain //chainId
}{Clist: make(map[string]*chain)}

func chainsInit() {
	chains.Clist = nil
	chains.Clist = make(map[string]*chain)
}

func addChainCodeInfo(c *chain, cc *CCInfo) error {
	if c == nil || cc == nil {
		return errors.New("chain or ccinfo is nil")
	}

	for k, v := range c.CClist {
		if k == cc.Name && v.Version == cc.Version {
			log.Error("addChainCodeInfo", "chaincode  already exit, name", cc.Name, "version", cc.Version)
			return errors.New("already exit chaincode")
		}
	}
	c.CClist[cc.Name] = cc

	return nil
}

func SetChaincode(cid string, version int, chaincode *CCInfo) error {
	chains.mu.Lock()
	defer chains.mu.Unlock()
	log.Info("SetChaincode", "chainId", cid, "version", version, "Name", chaincode.Name, "Id", chaincode.Id)
	for k, v := range chains.Clist {
		if k == cid {
			log.Info("SetChaincode", "chainId already exit, cid:", cid, "version", v)
			return addChainCodeInfo(v, chaincode)
		}
	}
	cNew := chain{
		Version: version,
		CClist:  make(map[string]*CCInfo),
	}
	chains.Clist[cid] = &cNew

	return addChainCodeInfo(&cNew, chaincode)
}

func GetChaincodeList(cid string) (*chain, error) {
	if cid == "" {
		return nil, errors.New("param is nil")
	}

	if chains.Clist[cid] != nil {
		return chains.Clist[cid], nil
	}
	errmsg := fmt.Sprintf("not find chainId[%s] in chains", cid)

	return nil, errors.New(errmsg)
}

func GetChaincode(cid string, deployId []byte) (*CCInfo, error) {
	if cid == "" {
		return nil, errors.New("param is nil")
	}
	if chains.Clist[cid] != nil {
		clist := chains.Clist[cid]
		for _, v := range clist.CClist {
			log.Info("GetChaincode", "find chaincode,name", v.Name, "list id", v.Id, "depId", deployId)
			if bytes.Equal(v.Id, deployId) {
				return v, nil
			}
		}
	}
	errmsg := fmt.Sprintf("not find chainId[%s], deployId[%x] in chains", cid, deployId)
	return nil, errors.New(errmsg)
}

func DelChaincode(cid string, ccName string, version string) error {
	chains.mu.Lock()
	defer chains.mu.Unlock()

	if cid == "" || ccName == "" {
		return errors.New("param is nil")
	}
	if chains.Clist[cid] != nil {
		delete(chains.Clist[cid].CClist, ccName)
		log.Info("DelChaincode", "delete chaincode", ccName)
		return nil
	}
	log.Info("DelChaincode", "not find chaincode", ccName)

	return nil
}

func init() {
	chainsInit()
}
