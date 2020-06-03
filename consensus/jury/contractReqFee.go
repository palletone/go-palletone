package jury

import (
	"encoding/hex"
	"math/big"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/gen"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/tokenengine"
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
	ContractDefaultSignatureSize         = 256.0
	ContractDefaultElectionSize          = 768.0
	ContractDefaultRWSize                = 512.0
	ContractDefaultPayInputSignatureSize = 256.0
)

type Txo4Greedy struct {
	modules.OutPoint
	Amount uint64
}

func (txo *Txo4Greedy) GetAmount() uint64 {
	return txo.Amount
}
func newTxo4Greedy(outPoint modules.OutPoint, amount uint64) *Txo4Greedy {
	return &Txo4Greedy{
		OutPoint: outPoint,
		Amount:   amount,
	}
}

func (p *Processor) getTxContractFee(tx *modules.Transaction, extDataSize float64, timeout uint32) (fee float64,
	size float64, tm uint32, err error) {
	if tx == nil {
		return 0, 0, 0, errors.New("getTxContractFee, param is nil")
	}
	reqId := tx.RequestHash()
	txType := tx.GetContractTxType()
	if txType == modules.APP_UNKNOW {
		log.Error("[%s]getTxContractFee,getContractTxType APP_UNKNOW", reqId.ShortStr())
		return 0, 0, 0, err
	}
	allSize := tx.Size().Float64() + extDataSize
	timeFee, sizeFee := getContractTxNeedFee(p.dag, txType, float64(timeout), allSize) //todo  timeout
	log.Debugf("[%s]getTxContractFee, all txFee[%f],timeFee[%f],sizeFee[%f]", reqId.ShortStr(), timeFee+sizeFee, timeFee, sizeFee)
	return timeFee + sizeFee, allSize, timeout, nil
}
func (p *Processor) createBaseTransaction(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	enableGasFee bool) (*modules.Transaction, error) {
	//daoTotal := daoAmount + daoFee
	// 20200412  wzhyuan
	daoTotal := daoAmount
	if from.Equal(to) {
		if enableGasFee {
			daoTotal += daoFee
		} else {
			tx := &modules.Transaction{}
			certIDBytes := []byte{}
			if certID != nil {
				certIDBytes = certID.Bytes()
			}
			tx.SetCertId(certIDBytes)
			return tx, nil
		}
	} else {
		if enableGasFee {
			daoTotal += daoFee
		}
	}
	// 1. 获取转出账户所有的PTN utxo
	//allUtxos, err := dag.GetAddrUtxos(from)
	//  20200412   wzhyuan
	assetId := dagconfig.DagConfig.GetGasToken()
	coreUtxos, err := p.ptn.TxPool().GetAddrUtxos(from, assetId.ToAsset())
	if err != nil {
		return nil, err
	}
	//coreUtxos, err := b.GetPoolAddrUtxos(fromAddr, gasAsset.ToAsset())
	//if err != nil {
	//	return nil,  fmt.Errorf("GetPoolAddrUtxos err:%s", err.Error())
	//}
	if len(coreUtxos) == 0 {
		return nil, err
	}

	// 2. 利用贪心算法得到指定额度的utxo集合
	greedyUtxos := core.Utxos{}
	for outPoint, utxo := range coreUtxos {
		tg := newTxo4Greedy(outPoint, utxo.Amount)
		greedyUtxos = append(greedyUtxos, tg)
	}

	selUtxos, change, err := core.Select_utxo_Greedy(greedyUtxos, daoTotal)
	if err != nil {
		return nil, err
	}

	// 3. 构建PaymentPayload的Inputs
	pload := new(modules.PaymentPayload)
	pload.LockTime = 0

	for _, selTxo := range selUtxos {
		tg := selTxo.(*Txo4Greedy)
		txInput := modules.NewTxIn(&tg.OutPoint, []byte{})
		pload.AddTxIn(txInput)
	}

	// 4. 构建PaymentPayload的Outputs
	// 为了保证顺序， 将map改为结构体数组
	type OutAmount struct {
		addr   common.Address
		amount uint64
	}

	outAmounts := make([]*OutAmount, 1, 2)
	outAmount := &OutAmount{to, daoAmount}
	outAmounts[0] = outAmount

	if change > 0 {
		// 处理from和to是同一个地址的特殊情况
		if from.Equal(to) {
			outAmount.amount = outAmount.amount + change
			outAmounts[0] = outAmount
		} else {
			outAmounts = append(outAmounts, &OutAmount{from, change})
		}
	}
	asset := dagconfig.DagConfig.GetGasToken().ToAsset()
	for _, outAmount := range outAmounts {
		pkScript := tokenengine.Instance.GenerateLockScript(outAmount.addr)
		txOut := modules.NewTxOut(outAmount.amount, pkScript, asset)
		pload.AddTxOut(txOut)
	}

	// 5. 构建Transaction

	certIDBytes := []byte{}
	if certID != nil {
		certIDBytes = certID.Bytes()
	}
	tx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_PAYMENT, pload)})
	tx.SetCertId(certIDBytes)
	return tx, nil
}

func (p *Processor) CreateGenericTransaction(from, to common.Address, daoAmount, daoFee uint64, certID *big.Int,
	msg *modules.Message, enableGasFee bool) (*modules.Transaction, uint64, error) {
	// 如果是 text，则增加费用，以防止用户任意增加文本，导致网络负担加重
	if msg.App == modules.APP_DATA {
		size := float64(modules.CalcDateSize(msg.Payload))
		pricePerKByte := p.dag.GetChainParameters().TransferPtnPricePerKByte
		daoFee += uint64(size * float64(pricePerKByte) / 1024)
	}
	tx, err := p.createBaseTransaction(from, to, daoAmount, daoFee, certID, enableGasFee)
	if err != nil {
		return nil, 0, err
	}

	tx.AddMessage(msg)

	return tx, daoFee, nil
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
	//reqTx, _, err := p.dag.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.EnableGasFee())
	reqTx, _, err := p.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.EnableGasFee())
	if err != nil {
		log.Error("ContractInstallReqFee", "CreateGenericTransaction err:", err)
		return 0, 0, 0, err
	}
	ctx := &contracts.ContractProcessContext{
		RwM: rwset.RwM,
		Dag: p.dag,
		//TxPool: p.ptn.TxPool(),
		Contract:     p.contract,
		ErrMsgEnable: p.errMsgEnable}
	msgs, err := runContractCmd(ctx, reqTx)
	if err != nil {
		log.Error("ContractInstallReqFee", "RunContractCmd err:", err)
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
	tx, _, err := p.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.EnableGasFee())
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
	tx, _, err := p.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.EnableGasFee())
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
	tx, _, err := p.CreateGenericTransaction(from, to, daoAmount, daoFee, nil, msgReq, p.ptn.EnableGasFee())
	if err != nil {
		log.Error("ContractStopReqFee", "CreateGenericTransaction err:", err)
		return 0, 0, 0, err
	}

	return p.getTxContractFee(tx, ContractDefaultSignatureSize, 0)
}
