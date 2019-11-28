package jury

import (
	"time"
	"math/big"
	"encoding/hex"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/crypto"
)

/*
eg.
size(byte)
           req      txPool
install    1577     1691
deploy     373      1029
invoke     229      496
*/

const (
	ContractDefaultSignatureSize = 256.0
	ContractDefaultElectionSize  = 768.0
	ContractDefaultRWSize        = 512.0
)

func (p *Processor) getTxContractFee(tx *modules.Transaction, extDataSize float64, timeout uint32) (fee float64,
	size float64, tm uint32, err error) {
	if tx == nil {
		return 0, 0, 0, errors.New("getTxContractFee, param is nil")
	}
	reqId := tx.RequestHash()
	txType, err := getContractTxType(tx)
	if err != nil {
		log.Errorf("[%s]getTxContractFee,getContractTxType err:%s", shortId(reqId.String()), err.Error())
		return 0, 0, 0, err
	}
	allSize := tx.Size().Float64() + extDataSize
	timeFee, sizeFee := getContractTxNeedFee(p.dag, txType, float64(timeout), allSize) //todo  timeout
	log.Debugf("[%s]getTxContractFee, all txFee[%f],timeFee[%f],sizeFee[%f]", shortId(reqId.String()), timeFee+sizeFee, timeFee, sizeFee)
	return timeFee + sizeFee, allSize, timeout, nil
}

func (p *Processor) ContractInstallReqFee(from, to common.Address, daoAmount, daoFee uint64, tplName, path, version string,
	description, abi, language string, local bool, addrs []common.Address) (fee float64, size float64, tm uint32, err error) {
	addrHash := make([]common.Hash, 0)
	resultAddress := getValidAddress(addrs)
	for _, addr := range resultAddress {
		addrHash = append(addrHash, util.RlpHash(addr))
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_TPL_REQUEST,
		Payload: &modules.ContractInstallRequestPayload{
			TplName:        tplName,
			Path:           path,
			Version:        version,
			AddrHash:       addrHash,
			TplDescription: description,
			Abi:            abi,
			Language:       language,
			Creator:        from.String(),
		},
	}
	reqTx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.TxPool())
	if err != nil {
		log.Error("ContractInstallReqFee", "CreateGenericTransaction err:", err)
		return 0, 0, 0, err
	}
	msgs, err := runContractCmd(rwset.RwM, p.dag, p.contract, reqTx, nil, p.errMsgEnable)
	if err != nil {
		log.Error("ContractInstallReqFee", "runContractCmd err:", err)
		return 0, 0, 0, err
	}
	tx, err := gen.GenContractTransction(reqTx, msgs)
	if err != nil {
		log.Error("ContractInstallReqFee", "GenContractTransaction err:", err)
		return 0, 0, 0, err
	}
	return p.getTxContractFee(tx, ContractDefaultSignatureSize, 0)
}

func (p *Processor) ContractDeployReqFee(from, to common.Address, daoAmount, daoFee uint64, templateId []byte,
	args [][]byte, extData []byte, timeout time.Duration) (fee float64, size float64, tm uint32, err error) {
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_DEPLOY_REQUEST,
		Payload: &modules.ContractDeployRequestPayload{
			TemplateId: templateId,
			Args:       args,
			ExtData:    extData,
			Timeout:    uint32(timeout),
		},
	}
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.TxPool())
	if err != nil {
		log.Error("ContractDeployReqFee", "CreateGenericTransaction err:", err)
		return 0, 0, 0, err
	}
	return p.getTxContractFee(tx, ContractDefaultSignatureSize+ContractDefaultElectionSize, 0)
}

func (p *Processor) ContractInvokeReqFee(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	contractId common.Address, args [][]byte, timeout uint32) (fee float64, size float64, tm uint32, err error) {
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_INVOKE_REQUEST,
		Payload: &modules.ContractInvokeRequestPayload{
			ContractId: contractId.Bytes(),
			Args:       args,
			Timeout:    timeout,
		},
	}
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.TxPool())
	if err != nil {
		log.Error("ContractInvokeReqFee", "CreateGenericTransaction err:", err)
		return 0, 0, 0, err
	}
	return p.getTxContractFee(tx, ContractDefaultSignatureSize+ContractDefaultRWSize, timeout)
}

func (p *Processor) ContractStopReqFee(from, to common.Address, daoAmount, daoFee uint64,
	contractId common.Address, deleteImage bool) (fee float64, size float64, tm uint32, err error) {
	randNum, err := crypto.GetRandomNonce()
	if err != nil {
		return 0, 0, 0, errors.New("ContractStopReqFee, GetRandomNonce error")
	}
	msgReq := &modules.Message{
		App: modules.APP_CONTRACT_STOP_REQUEST,
		Payload: &modules.ContractStopRequestPayload{
			ContractId:  contractId.Bytes(),
			Txid:        hex.EncodeToString(randNum),
			DeleteImage: deleteImage,
		},
	}
	tx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.TxPool())
	if err != nil {
		log.Error("ContractStopReqFee", "CreateGenericTransaction err:", err)
		return 0, 0, 0, err
	}

	return p.getTxContractFee(tx, ContractDefaultSignatureSize, 0)
}
