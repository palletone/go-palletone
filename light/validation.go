package light

import (
	"github.com/palletone/go-palletone/common"
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
	txhash common.Hash // Hash of the block being announced
	time   time.Time   // Timestamp of the announcement
	step   chan int    //0:ok   1:err  2:timeout
}

type proofResp struct {
	txhash common.Hash
	resp   []les.NodeList
}

type Validation struct {
	preq     map[common.Hash]*proofReq //key:txhash  request queue
	preqLock sync.RWMutex

	queue *prque.Prque //recv validation path

	quit chan struct{}
}

func NewValidation() *Validation {
	return &Validation{
		preq:  make(map[common.Hash]*proofReq),
		queue: prque.New(),
		quit:  make(chan struct{}),
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
	}
	select {
	case <-v.quit:
		// loop terminating, abort all operations
		return
	case <-completeTimer.C:
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
