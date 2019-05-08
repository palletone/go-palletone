package light

import (
	"time"

	"github.com/palletone/go-palletone/dag/modules"
	"sync"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/common/log"
	"errors"
	"fmt"
	"encoding/json"
	"github.com/palletone/go-palletone/common"
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
	addr  []byte    `json:"addr"`
	utxos [][][]byte  `json:"utxos"`
}

type utxosRespData struct {
	addr  string
	utxos map[modules.OutPoint]*modules.Utxo
}

func NewUtxosRespData()*utxosRespData  {
	return &utxosRespData{utxos:make(map[modules.OutPoint]*modules.Utxo)}
}

func (u *utxosRespData)encode()([][][]byte,error){
	//var datas lpsutxo
	//datas.addr = []byte(u.addr)
	var addrarr [][]byte
	var arrs [][][]byte
	addrarr = append(addrarr,[]byte(u.addr))
	arrs = append(arrs,addrarr)


	for outpoint,utxo:=range u.utxos{
		var data [][]byte
		d1,err:=json.Marshal(outpoint)
		if err!=nil{return arrs,err}
		log.Debug("Light PalletOne","utxosRespData encode outpoint",string(d1))
		data = append(data,d1)

		d2,err:=json.Marshal(utxo)
		if err!=nil{return arrs,err}
		log.Debug("Light PalletOne","utxosRespData encode utxo",string(d2))
		data = append(data,d2)
		arrs = append(arrs,data)
	}
	return arrs,nil
}

func (u *utxosRespData)decode(arrs [][][]byte)error{
	u.addr = string(arrs[0][0])

	for _,arr :=range arrs[1:]{
		var outpoint modules.OutPoint
		var utxo *modules.Utxo
		log.Debug("Light PalletOne","utxosRespData decode outpoint",string(arr[0]))
		log.Debug("Light PalletOne","utxosRespData decode utxo",string(arr[1]))
		if err:=json.Unmarshal(arr[0],&outpoint);err!=nil{return err}
		if err:=json.Unmarshal(arr[1],&utxo);err!=nil{return err}
		u.utxos[outpoint] = utxo
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

	address, err := common.StringToAddress(respdata.addr)
	if err != nil {
		log.Debug("Light PalletOne","SaveUtxoView err:",err,"addr",respdata.addr)
		return err
	}

	if err:=u.dag.ClearUtxo(address);err!=nil{
		log.Debug("Light PalletOne","SaveUtxoView ClearUtxo err:",err,"addr",respdata.addr)
		return err
	}
	if err:=u.dag.SaveUtxoView(respdata.utxos);err!=nil{
		log.Debug("Light PalletOne", "SaveUtxoView key err",err,"addr:", respdata.addr)
		return err
	}
	req.step <- OKUTXOsSync
	return nil
}