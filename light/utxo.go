package light

import (
	"time"

	"github.com/palletone/go-palletone/dag/modules"
	"sync"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/common/log"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	utxosArriveTimeout = 500 * time.Millisecond // Time allowance before an announced block is explicitly requested
	utxosGatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	utxosReqTimeout    = 5 * time.Second        // Maximum allotted time to return an explicitly requested block
	utxosMaxQueueDist  = 32                     // Maximum allowed distance from the chain head to queue
	utxosHashLimit     = 256                    // Maximum number of unique blocks a peer may have announced
	utxosReqLimit      = 64                     // Maximum number of unique blocks a peer may have delivered
	OKUTXOsSync         = 0
	ERRUTXOsOTHERS     = 1
	ERRUTXOsTIMEOUT    = 2
	ERRUTXOsExist      = 3
)

type lpsutxo struct {
	outpoint modules.OutPoint
	utxo modules.Utxo
}

type utxosRespData struct {
	addr  string
	utxos map[modules.OutPoint]*modules.Utxo
}

func NewUtxosRespData()*utxosRespData  {
	return &utxosRespData{utxos:make(map[modules.OutPoint]*modules.Utxo)}
}

func (u *utxosRespData)encode()([][]byte,error){
	var datas [][]byte
	datas = append(datas,[]byte(u.addr))
	for outpoint,utxo:=range u.utxos{
		lu :=lpsutxo{outpoint:outpoint,utxo:*utxo}
		data,err:=rlp.EncodeToBytes(lu)
		if err!=nil{
			return nil,err
		}
		datas = append(datas,data)
	}

	return datas,nil
}

func (u *utxosRespData)decode(datas [][]byte)error{
	u.addr = string(datas[0])
	for _,datalus:= range datas[1:]{
		lu :=lpsutxo{}
		if err:=rlp.DecodeBytes(datalus,&lu);err!=nil{
			return err
		}
		u.utxos[lu.outpoint] = &lu.utxo
	}
	return nil
}


type utxosReq struct {
	addr string
	time     time.Time   // Timestamp of the announcement
	step     chan int    //0:ok   1:err  2:timeout
	utxosync *UtxosSync
}

type UtxosSync struct {
	reqs map[string]*utxosReq //key:addr
	lock sync.RWMutex
	dag dag.IDag
}


func NewUTXOsReq(addr string,utxosync *UtxosSync) *utxosReq {
	return &utxosReq{addr:addr,time:time.Now(),step:make(chan int),utxosync:utxosync}
}

func (req *utxosReq) Wait() int {
	timeout := time.NewTicker(spvReqTimeout)
	defer timeout.Stop()
	for {
		select {
		case result := <-req.step:
			//req.valid.forgetHash(req.strindex)
			req.utxosync.forgetHash(req.addr)
			return result
		case <-timeout.C:
			req.utxosync.forgetHash(req.addr)
			return ERRSPVTIMEOUT
		}
	}
}

func NewUtxosSync(dag dag.IDag) *UtxosSync {
	return &UtxosSync{
		dag:dag,
		reqs:make(map[string]*utxosReq),
		}
}

func (u *UtxosSync)AddUtxoSyncReq(addr string) (*utxosReq,error) {
	u.lock.RLock()
	if req,ok:=u.reqs[addr];ok{
		u.lock.RUnlock()

		req.step <- ERRUTXOsExist
		log.Debug("Light PalletOne", "StartSyncByAddr key is exist. addr:", addr)
		return nil,errors.New("Key is not exist")
	}
	u.lock.RUnlock()

	req := NewUTXOsReq(addr,u)
	u.lock.Lock()
	u.reqs[addr] = req
	u.lock.Unlock()
	return req,nil
}

func (u *UtxosSync)forgetHash(addr string){
	u.lock.Lock()
	delete(u.reqs,addr)
	u.lock.Unlock()
}

func (u *UtxosSync)SaveUtxoView(respdata *utxosRespData)error{
	u.lock.RLock()
	req,ok:=u.reqs[respdata.addr]
	if !ok{
		u.lock.RUnlock()
		log.Debug("Light PalletOne", "SaveUtxoView key is not exist. addr:", respdata.addr)
		return errors.New(fmt.Sprintf("addr(%v) is not exist",respdata.addr))
	}
	u.lock.RUnlock()

	if err:=u.dag.SaveUtxoView(respdata.utxos);err!=nil{
		log.Debug("Light PalletOne", "SaveUtxoView key err",err,"addr:", respdata.addr)
		return err
	}
	req.step <- OKUTXOsSync
	return nil
}