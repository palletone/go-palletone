package light

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/light/les"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
	"sync"
	"time"
)

const (
	spvArriveTimeout = 500 * time.Millisecond // Time allowance before an announced block is explicitly requested
	spvGatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	spvReqTimeout    = 5 * time.Second        // Maximum allotted time to return an explicitly requested block
	spvMaxQueueDist  = 32                     // Maximum allowed distance from the chain head to queue
	spvHashLimit     = 256                    // Maximum number of unique blocks a peer may have announced
	spvReqLimit      = 64                     // Maximum number of unique blocks a peer may have delivered
)

type proofReq struct {
	txhash common.Hash    // Hash of the block being announced
	time   time.Time      // Timestamp of the announcement
	step   chan proofResp //0:ok   1:err  2:timeout
}

func (req *proofReq) Wait() proofResp {
	return <-req.step
}

type proofResp struct {
	txhash common.Hash
	resp   []les.NodeList
}

type Validation struct {
	preq     map[common.Hash]*proofReq //key:txhash  request queue
	preqLock sync.RWMutex

	queue *prque.Prque //recv validation path
	dag   dag.IDag

	quit chan struct{}
}

func NewValidation(dag dag.IDag) *Validation {
	return &Validation{
		preq:  make(map[common.Hash]*proofReq),
		queue: prque.New(),
		quit:  make(chan struct{}),
		dag:   dag,
	}
}

func (v *Validation) Start() {
	go v.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (v *Validation) Stop() {
	close(v.quit)
}

func (v *Validation) loop() {
	completeTimer := time.NewTimer(0)
	for {
		// Import any queued blocks that could potentially fit
		//var height uint64
		for !v.queue.Empty() {
			op := v.queue.PopItem().(*proofResp)
			v.process(op)

		}
		select {
		case <-v.quit:
			// loop terminating, abort all operations
			return
		case <-completeTimer.C:
		}
	}

}

func (v *Validation) forgetHash(hash common.Hash) {
	//send channel step   0:ok   1:err  2:timeout
	//delete(v.preq,hash)
}

func (v *Validation) process(op *proofResp) {
	go func() {
		//check hash.Is it exsit in preq queue.yes,save local leveldb and notice console.no,notice console
		//for hash, req := range v.preq {
		//	if time.Since(req.time) > fetchTimeout {
		//		v.forgetHash(hash)
		//	}
		//}
	}()
}

func (v *Validation) Check(resp proofsRespData) (int, error) {
	v.preqLock.RLock()
	if _, ok := v.preq[resp.txhash]; !ok {
		return 0, errors.New("Key is not exist")
	}
	v.preqLock.RUnlock()
	/*header*/ _, err := v.dag.GetHeaderByHash(resp.headerhash)
	if err != nil {
		log.Debug("===========")
		return 0, err
	}

	//proofs := resp.pathData
	//if len(proofs) != 1 {
	//	return 0, errInvalidEntryCount
	//}
	//nodeSet := proofs.NodeSet()
	//trie.VerifyProof(header.TxRoot, resp.key, nodeSet)
	//if err != nil {
	//	log.Errorf("VerifyProof error for key %x: %v\nraw proof: %v", resp.key, err, proofs)
	//	return 0, err
	//}
	log.Debug("+++++++++++Validation->Check++++++++++", "resp", resp)
	return 0, nil
}

func (v *Validation) AddSpvResp(resps []proofsRespData) (*proofReq, error) {
	for _, resp := range resps {
		v.Check(resp)
	}
	return nil, nil
}

func (v *Validation) AddSpvReq(strhash string) (*proofReq, error) {
	hash := common.Hash{}
	hash.SetHexString(strhash)
	log.Debug("Light PalletOne ProtocolManager ReqProof", "strhash", strhash, "common hash", hash.String())
	req := &proofReq{txhash: hash, time: time.Now(), step: make(chan proofResp)}

	v.preqLock.RLock()
	if _, ok := v.preq[hash]; ok {
		v.preqLock.RUnlock()
		return nil, errors.New("Key is exist")
	}
	v.preqLock.RUnlock()

	v.preqLock.Lock()
	v.preq[hash] = req
	v.preqLock.Unlock()
	return req, nil
}

//str := `{
//	"tx_hash": "0x162c55f36093da2ae4009521cdf80c5a9686c0662967e1cdc6e76fc2b55b270c",
//	"tx_size": 302,
//	"payment": {
//		"inputs": [{
//			"txid": "0x997e5eccf274fd6c612f508b7edfc5901cee88e7115ff9802cb9be0ba192759e",
//			"message_index": 0,
//			"out_index": 0,
//			"unlock_script": "5b98e83534080e9bd4d07d8a963d237661af438c940428af737248e30b58eab3185396e52f6bbd4dc7ce9075ea5657a5e2c884a06efbe83f8f9ae2e5f1f0377d01 031e976a58bb67b150a5df910715cd7c2dd4511c79f197832789f1601b6cbcef06",
//			"from_address": "P17GojddWTJr1wJbQqkRS3E3vtisfrxtQhp"
//		}],
//		"outputs": [{
//			"amount": 9900000000,
//			"asset": "PTN",
//			"to_address": "P1M1wwyGiBQAda5T25Rcuywnyxi2bxAZCWk",
//			"lock_script": "OP_DUP OP_HASH160 db90623dea2c87bff041811eeb2a501ceba6e98d OP_EQUALVERIFY OP_CHECKSIG"
//		}, {
//			"amount": 99999990099999980,
//			"asset": "PTN",
//			"to_address": "P17GojddWTJr1wJbQqkRS3E3vtisfrxtQhp",
//			"lock_script": "OP_DUP OP_HASH160 44ce01ae3f3d921cb77f9713c270096fb3292b59 OP_EQUALVERIFY OP_CHECKSIG"
//		}],
//		"locktime": 0
//	},
//	"account_state_update": null,
//	"data": null,
//	"contract_tpl": null,
//	"contract_deploy": null,
//	"contract_invoke": null,
//	"contract_stop": null,
//	"signature": null,
//	"install_request": null,
//	"deploy_request": null,
//	"invoke_request": null,
//	"stop_request": null
//}`
//
//tx := modules.Transaction{}
//if err := json.Unmarshal([]byte(str), &tx); err != nil {
//	log.Error("Light PalletOne ProtocolManager ReqProof", "Unmarshal err:", err)
//	return err
//}
//log.Debug("Light PalletOne ProtocolManager ReqProof", "tx.Hash()", tx.Hash().String(), "tx_hash", "0x162c55f36093da2ae4009521cdf80c5a9686c0662967e1cdc6e76fc2b55b270c")
