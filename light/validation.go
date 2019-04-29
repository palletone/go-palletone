package light

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/trie"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
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
	txhash common.Hash // Hash of the block being announced
	time   time.Time   // Timestamp of the announcement
	step   chan int    //0:ok   1:err  2:timeout
	valid  *Validation
}

func NewProofReq(txhash common.Hash, valid *Validation) *proofReq {
	return &proofReq{txhash: txhash, time: time.Now(), step: make(chan int), valid: valid}
}

func (req *proofReq) Wait() int {
	timeout := time.NewTicker(spvReqTimeout)
	defer timeout.Stop()
	for {
		select {
		case result := <-req.step:
			return result
		case <-timeout.C:
			req.valid.forgetHash(req.txhash)
			return 2
		}
	}
}

type Validation struct {
	preq     map[common.Hash]*proofReq //key:txhash  request queue
	preqLock sync.RWMutex

	queue *prque.Prque //recv validation path
	dag   dag.IDag

	//quit chan struct{}
}

func NewValidation(dag dag.IDag) *Validation {
	return &Validation{
		preq:  make(map[common.Hash]*proofReq),
		queue: prque.New(),
		//quit:  make(chan struct{}),
		dag: dag,
	}
}

func (v *Validation) Start() {
	//go v.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (v *Validation) Stop() {
	//close(v.quit)
}

func (v *Validation) forgetHash(txhash common.Hash) {
	v.preqLock.Lock()
	delete(v.preq, txhash)
	v.preqLock.Unlock()
}

func (v *Validation) Check(resp *proofsRespData) (int, error) {

	header, err := v.dag.GetHeaderByHash(resp.headerhash)
	if err != nil {
		log.Debug("Light PalletOne", "Validation->Check GetHeaderByHash err", err, "header hash", resp.headerhash)
		return 0, err
	}
	if header.TxRoot.String() != resp.txroothash.String() {
		return 0, errors.New("txroothash not equal")
	}
	log.Debug("Light PalletOne", "key", resp.key, "proof", resp.pathData)
	nodeSet := resp.pathData.NodeSet()
	_, err, _ = trie.VerifyProof(header.TxRoot, resp.key, nodeSet)
	if err != nil {
		log.Debug("Light PalletOne", "Validation->Check VerifyProof err", err)
		return 0, err
	}
	return 0, nil
}

func (v *Validation) AddSpvResp(resp *proofsRespData) error {
	v.preqLock.RLock()
	vreq, ok := v.preq[resp.txhash]
	if !ok {
		log.Debug("Light PalletOne", "Validation->Check key is not exist.key", resp.txhash)
		return errors.New("Key is not exist")
	}
	v.preqLock.RUnlock()
	_, err := v.Check(resp)
	if err != nil {
		vreq.step <- 1
		return err
	}
	vreq.step <- 0
	return nil
}

func (v *Validation) AddSpvReq(strhash string) (*proofReq, error) {
	hash := common.Hash{}
	hash.SetHexString(strhash)
	log.Debug("Light PalletOne ProtocolManager ReqProof", "strhash", strhash, "common hash", hash.String())

	v.preqLock.RLock()
	if _, ok := v.preq[hash]; ok {
		v.preqLock.RUnlock()
		return nil, errors.New("Key is exist")
	}
	v.preqLock.RUnlock()
	req := NewProofReq(hash, v)
	v.preqLock.Lock()
	v.preq[hash] = req
	v.preqLock.Unlock()
	return req, nil
}
