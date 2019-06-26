package light

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/trie"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/errors"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
	"strconv"
	"sync"
	"time"
)

const (
	spvArriveTimeout = 500 * time.Millisecond // Time allowance before an announced block is explicitly requested
	spvGatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	spvReqTimeout    = 30 * time.Second       // Maximum allotted time to return an explicitly requested block
	spvMaxQueueDist  = 32                     // Maximum allowed distance from the chain head to queue
	spvHashLimit     = 256                    // Maximum number of unique blocks a peer may have announced
	spvReqLimit      = 64                     // Maximum number of unique blocks a peer may have delivered
	ERRSPVOTHERS     = 1
	ERRSPVTIMEOUT    = 2
)

type proofReq struct {
	strindex string
	txhash   common.Hash // Hash of the block being announced
	time     time.Time   // Timestamp of the announcement
	step     chan int    //0:ok   1:err  2:timeout
	valid    *Validation
}

func NewProofReq(txhash common.Hash, valid *Validation) *proofReq {
	t := time.Now().UnixNano()
	str := strconv.FormatInt(t, 10)
	return &proofReq{strindex: str + txhash.String(), txhash: txhash, time: time.Now(), step: make(chan int), valid: valid}
}

func (req *proofReq) Wait() int {
	timeout := time.NewTicker(spvReqTimeout)
	defer timeout.Stop()
	for {
		select {
		case result := <-req.step:
			req.valid.forgetHash(req.strindex)
			return result
		case <-timeout.C:
			req.valid.forgetHash(req.strindex)
			return ERRSPVTIMEOUT
		}
	}
}

type Validation struct {
	preq     map[string]*proofReq //key:txhash  request queue
	preqLock sync.RWMutex

	queue *prque.Prque //recv validation path
	dag   dag.IDag

	//quit chan struct{}
}

func NewValidation(dag dag.IDag) *Validation {
	return &Validation{
		preq:  make(map[string]*proofReq),
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

func (v *Validation) forgetHash(index string) {
	v.preqLock.Lock()
	delete(v.preq, index)
	v.preqLock.Unlock()
}

func (v *Validation) Check(resp *proofsRespData) (int, error) {
	header, err := v.dag.GetHeaderByHash(resp.headerhash)
	if err != nil {
		log.Debug("Light PalletOne", "Validation->Check GetHeaderByHash err", err, "header hash", resp.headerhash)
		return 0, err
	}
	//TODO recover
	//if header.TxRoot.String() != resp.txroothash.String() {
	//	return 0, errors.New("txroothash not equal")
	//}
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
	vreq, ok := v.preq[resp.index]
	if !ok {
		v.preqLock.RUnlock()

		//vreq.step <- ERRSPVOTHERS
		log.Debug("Light PalletOne", "Validation->Check key is not exist.key", resp.index)
		return errors.New("Key is not exist")
	}
	v.preqLock.RUnlock()
	_, err := v.Check(resp)
	if err != nil {
		vreq.step <- ERRSPVOTHERS
		return err
	}
	vreq.step <- 0
	return nil
}

func (v *Validation) AddSpvReq(strhash string) (*proofReq, error) {
	hash := common.Hash{}
	hash.SetHexString(strhash)
	log.Debug("Light PalletOne ProtocolManager ReqProof", "strhash", strhash, "common hash", hash.String())

	//TODO add limit console visit times
	//v.preqLock.RLock()
	//if _, ok := v.preq[hash]; ok {
	//	v.preqLock.RUnlock()
	//	return nil, errors.New("Key is exist")
	//}
	//v.preqLock.RUnlock()

	req := NewProofReq(hash, v)
	v.preqLock.Lock()
	v.preq[req.strindex] = req
	v.preqLock.Unlock()
	return req, nil
}
