/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018-2020
 */
package fabricadaptor

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/palletone/fabric-adaptor/pkg/client/channel"
	"github.com/palletone/fabric-adaptor/pkg/client/ledger"
	"github.com/palletone/fabric-adaptor/pkg/client/msp"
	"github.com/palletone/fabric-adaptor/pkg/client/resmgmt"
	"github.com/palletone/fabric-adaptor/pkg/common/providers/fab"
	"github.com/palletone/fabric-adaptor/pkg/context"
	"github.com/palletone/fabric-adaptor/pkg/core/config"
	"github.com/palletone/fabric-adaptor/pkg/core/cryptosuite"
	"github.com/palletone/fabric-adaptor/pkg/fab/ccpackager/gopackager"
	"github.com/palletone/fabric-adaptor/pkg/fabsdk"
	"github.com/palletone/fabric-adaptor/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	"github.com/palletone/fabric-adaptor/pkg/fab/txn"

	cb "github.com/hyperledger/fabric-protos-go/common"

	"github.com/palletone/adaptor"
)

type SDKParams struct {
	ConfigFile  string	//sdk的配置文件路径
	UserName string		//组织的普通用户
	ChannelID string	//通道id
	OrgAdmin string		//组织的管理员用户
	OrgName  string		//组织名字 organizations --- org1
	OrgID    string		//组织id organizations --- org1 --- mspid: Org1MSP
	EnvGoPath string	//GOPATH
}

type AdaptorFabric struct {
	SDKParams
	Sdk        *fabsdk.FabricSDK	//保存实例化后的sdk
	LedgerClient *ledger.Client
	ChannelClient *channel.Client
	ResClient *resmgmt.Client
	MspClient *msp.Client
}

func NewAdaptorFabric(configFile string) *AdaptorFabric {
	return &AdaptorFabric{SDKParams: SDKParams{ConfigFile: configFile},}
}

func InitSDK(afab *AdaptorFabric) error {
	if afab.Sdk != nil {
		return  nil
	}
	sdk, err := fabsdk.New(config.FromFile(afab.ConfigFile))
	if err != nil {
		//fmt.Println("fabsdk.New",err.Error())
		return err
	}
	//fmt.Println("fabsdk.New success")
	afab.Sdk = sdk
	return nil
}

func InitLedger(afab *AdaptorFabric) error  {
	if afab.LedgerClient != nil {
		return  nil
	}
	// 创建上下文
	clientContext := afab.Sdk.ChannelContext(afab.ChannelID, fabsdk.WithUser(afab.UserName))
	// 创建 ledger 客户端
	ledgerClient, err := ledger.New(clientContext)
	if err != nil {
		//fmt.Println(err.Error())
		return err
	}
	afab.LedgerClient = ledgerClient
	return nil
}

func InitChannel(afab *AdaptorFabric) error  {
	if afab.ChannelClient != nil {
		return  nil
	}
	// 创建上下文
	clientContext := afab.Sdk.ChannelContext(afab.ChannelID, fabsdk.WithUser(afab.UserName))
	// 创建channel客户端
	channelClient, err := channel.New(clientContext)
	if err != nil {
		//fmt.Println("channel.New",err.Error())
		return err
	}
	afab.ChannelClient = channelClient
	return nil
}

func InitResmgmt(afab *AdaptorFabric) error  {
	if afab.ResClient != nil {
		return  nil
	}
	afab.EnvGoPath = os.Getenv("GOPATH")
	//根据实例创建资源管理客户端
	//resCliProvider := afab.Sdk.Context(fabsdk.WithUser(afab.OrgAdmin),fabsdk.WithOrg(afab.OrgName))
	resCliProvider := afab.Sdk.ContextZxl(fabsdk.WithUser(afab.OrgAdmin),fabsdk.WithOrg(afab.OrgName))
	resClient, err := resmgmt.New(resCliProvider)
	if err != nil {
		//fmt.Println("resmgmt.New",err.Error())
		return err
	}
	afab.ResClient = resClient
	return nil
}

func InitMsp(afab *AdaptorFabric) error  {
	if afab.MspClient != nil {
		return  nil
	}
	mspClient, err := msp.New(afab.Sdk.Context(), msp.WithOrg(afab.OrgName))
	if err != nil {
		//fmt.Println("msp.New()", err.Error())
		return err
	}
	afab.MspClient = mspClient
	return nil
}

//use afab.UserName
func getClientContext(afab *AdaptorFabric) (fab.ClientContext, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	//providerFunc := afab.Sdk.Context(fabsdk.WithUser(afab.UserName))
	providerFunc := afab.Sdk.ContextZxl(fabsdk.WithUser(afab.UserName))
	clientProvider,err := providerFunc()
	if err != nil {
		//fmt.Println("providerFunc", err.Error())
		return nil, err
	}
	if nil == clientProvider {
		return nil, fmt.Errorf("clientProvider is nil")
	}
	return clientProvider, nil
}
func getClientContextNoUser(afab *AdaptorFabric) (fab.ClientContext, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	providerFunc := afab.Sdk.Context(fabsdk.WithUser(afab.UserName))
	//providerFunc := afab.Sdk.ContextZxl(fabsdk.WithUser(afab.UserName))
	clientProvider,err := providerFunc()
	if err != nil {
		//fmt.Println("providerFunc", err.Error())
		return nil, err
	}
	if nil == clientProvider {
		return nil, fmt.Errorf("clientProvider is nil")
	}
	return clientProvider, nil
}

/*IUtility*/
//创建一个新的私钥
func (afab *AdaptorFabric) NewPrivateKey(input *adaptor.NewPrivateKeyInput) (*adaptor.NewPrivateKeyOutput, error) {
	return nil, errors.New("not implement")
}

//根据私钥创建公钥
func (afab *AdaptorFabric) GetPublicKey(input *adaptor.GetPublicKeyInput) (
	*adaptor.GetPublicKeyOutput, error) {
	return nil, errors.New("not implement")
}

//根据Key创建地址
func (afab *AdaptorFabric) GetAddress(key *adaptor.GetAddressInput) (
	*adaptor.GetAddressOutput, error) {
	return nil, errors.New("not implement")
}
func (afab *AdaptorFabric) GetPalletOneMappingAddress(addrInput *adaptor.GetPalletOneMappingAddressInput) (
	*adaptor.GetPalletOneMappingAddressOutput, error) {
	return nil, errors.New("not implement")
}

//HashMessage have implement
func (afab *AdaptorFabric) HashMessage(input *adaptor.HashMessageInput) (*adaptor.HashMessageOutput, error) {
	clientContext, err := getClientContextNoUser(afab)
	if err != nil {
		return nil, err
	}
	ctxCryptoSuite := clientContext.CryptoSuite()
	if nil == ctxCryptoSuite {
		return nil, fmt.Errorf("ctxCryptoSuite is nil")
	}
	hash,err := ctxCryptoSuite.Hash(input.Message, cryptosuite.GetSHA256Opts())
	if err != nil {
		//fmt.Println("ctxCryptoSuite.Hash()", err.Error())
		return nil, err
	}
	//fmt.Println("hash", hex.EncodeToString(hash))

	var output adaptor.HashMessageOutput
	output.Hash = hash
	return &output, nil
}

//对一条消息进行签名 //SignMessage have implement
func (afab *AdaptorFabric) SignMessage(input *adaptor.SignMessageInput) (
	*adaptor.SignMessageOutput, error) {
	clientContext, err := getClientContext(afab)
	if err != nil {
		return nil, err
	}

	sig,err := clientContext.SigningManager().Sign(input.Message,
		clientContext.PrivateKey())//Zxl panic has fix
	if err != nil {
		//fmt.Println("clientContext.Sign()", err.Error())
		return nil, err
	}
	//fmt.Println("sig", hex.EncodeToString(sig))

	var output adaptor.SignMessageOutput
	output.Signature = sig
	return &output, nil
}

//对签名进行验证 //VerifySignature have implement
func (afab *AdaptorFabric) VerifySignature(input *adaptor.VerifySignatureInput) (
	*adaptor.VerifySignatureOutput, error) {
	clientContext, err := getClientContext(afab)
	if err != nil {
		return nil, err
	}
	var output adaptor.VerifySignatureOutput
	//err = clientContext.Verify(input.Message, input.Signature)
	//if err != nil {
	//	fmt.Println("clientContext.Verify()", err.Error())
	//	output.Pass = false
	//} else {
	//	output.Pass = true
	//}

	ctxCryptoSuite := clientContext.CryptoSuite()
	hash,err := ctxCryptoSuite.Hash(input.Message, cryptosuite.GetSHA256Opts())
	if err != nil {
		//fmt.Println("ctxCryptoSuite.Hash()", err.Error())
		return nil, err
	}
	valid,err := ctxCryptoSuite.Verify(clientContext.PrivateKey(), input.Signature,
		hash, nil)
	if err != nil {
		//fmt.Println("ctxCryptoSuite.Verify()", err.Error())
		output.Pass = false
	} else {
		output.Pass = valid
	}
	return &output, nil
}

//对一条交易进行签名，并返回签名结果
func (afab *AdaptorFabric) SignTransaction(input *adaptor.SignTransactionInput) (
	*adaptor.SignTransactionOutput, error) {
	clientContext, err := getClientContext(afab)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.NewRequest(clientContext)
	var proposal fab.TransactionProposal
	err = json.Unmarshal(input.Transaction, &proposal)
	if err != nil {
		return nil, err
	}
	processProposalRequest, err := txn.SignProposal(ctx, &proposal)
	if err != nil {
		//fmt.Println("txn.SignProposal", err.Error())
		return nil, err
	}

	resultJSON, err := json.Marshal(*processProposalRequest)
	if err != nil {
		//fmt.Println("json.Marshal(processProposalRequest)", err.Error())
		return nil, err
	}

	var output adaptor.SignTransactionOutput
	output.SignedTx = resultJSON
	return &output, nil
}

//将未签名的原始交易与签名进行绑定，返回一个签名后的交易
func (afab *AdaptorFabric) BindTxAndSignature(input *adaptor.BindTxAndSignatureInput) (
	*adaptor.BindTxAndSignatureOutput,
	error) {
	return nil, errors.New("not implement")
}

//根据交易内容，计算交易Hash
func (afab *AdaptorFabric) CalcTxHash(input *adaptor.CalcTxHashInput) (*adaptor.CalcTxHashOutput, error) {
	return nil, errors.New("not implement")
}

//将签名后的交易广播到网络中,如果发送交易需要手续费，指定最多支付的手续费
func (afab *AdaptorFabric) SendTransaction(input *adaptor.SendTransactionInput) (*adaptor.SendTransactionOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}

	txType := string(input.Extra)
	if "" != txType {
		switch txType {
		case "install":
			err = InitResmgmt(afab)
			if err != nil {
				return nil, err
			}
			var processProposalRequest fab.ProcessProposalRequest
			err = json.Unmarshal(input.Transaction, &processProposalRequest)
			if err != nil {
				return nil, err
			}
			_, txID, err := afab.ResClient.InstallCCBroadcastZxl(&processProposalRequest)
			if err != nil {
				//fmt.Println("ResClient.InstallCCBroadcastZxl", err.Error())
				return nil, err
			}
			//fmt.Println(len(rspArr))
			//fmt.Println(result)
			//fmt.Println("SendTransaction return", txID)

			var output adaptor.SendTransactionOutput // todo
			output.TxID = []byte(txID)
			return &output, nil
		case "init":
			err = InitResmgmt(afab)
			if err != nil {
				return nil, err
			}
			var processProposalRequest fab.ProcessProposalRequest
			err = json.Unmarshal(input.Transaction, &processProposalRequest)
			if err != nil {
				return nil, err
			}
			txID, err := afab.ResClient.InstantiateCCBroadcasZxl(afab.ChannelID, &processProposalRequest)
			if err != nil {
				//fmt.Println("ResClient.Broadcast", err.Error())
				return nil, err
			}
			var output adaptor.SendTransactionOutput // todo
			output.TxID = []byte(txID)
			return &output,nil
		case "invoke":
			err = InitChannel(afab)
			if err != nil {
				return nil, err
			}
			var processProposalRequest fab.ProcessProposalRequest
			err = json.Unmarshal(input.Transaction, &processProposalRequest)
			if err != nil {
				return nil, err
			}

			//写入之前,要创建请求:
			req := channel.Request{
				ChaincodeID:processProposalRequest.ChaincodeID,
				ProposalReq: &processProposalRequest,
			}
			resp, err := afab.ChannelClient.ExecuteBrocadcastZxl(req)
			if err != nil {
				//fmt.Println("ResClient.ExecuteBrocadcastFirstZxl", err.Error())
				return nil, err
			}
			var output adaptor.SendTransactionOutput // todo
			output.TxID = []byte(resp.TransactionID)
			return &output,nil

		}
	}

	return nil, fmt.Errorf("invalid extra")
}

//根据交易ID获得交易的基本信息
func (afab *AdaptorFabric) GetTxBasicInfo(input *adaptor.GetTxBasicInfoInput) (*adaptor.GetTxBasicInfoOutput, error) {
	return nil, errors.New("not implement")
}

//获取最新区块头 //GetBlockInfo have implement
func (afab *AdaptorFabric) GetBlockInfo(input *adaptor.GetBlockInfoInput) (*adaptor.GetBlockInfoOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	err = InitLedger(afab)
	if err != nil {
		return nil, err
	}
	var output adaptor.GetBlockInfoOutput
	if input.Latest {
		rsp, err := afab.LedgerClient.QueryInfo()
		if err != nil {
			//fmt.Println(err.Error())
			return nil, err
		}
		if rsp == nil {
			//fmt.Println("rsp is nil")
			return nil, fmt.Errorf("LedgerClient.QueryInfo response is nil")
		}
		//fmt.Println("rsp.Status", rsp.Status, "rsp.Endorser", rsp.Endorser)
		if rsp.BCI != nil {
			//fmt.Println("rsp.BCI.Height", rsp.BCI.Height)
			//fmt.Println("rsp.BCI.CurrentBlockHash", hex.EncodeToString(rsp.BCI.CurrentBlockHash))
			//fmt.Println("rsp.BCI.PreviousBlockHash", hex.EncodeToString(rsp.BCI.PreviousBlockHash))
			output.Block.IsStable = false //todo
			output.Block.BlockHeight = uint(rsp.BCI.Height)
			output.Block.BlockID = rsp.BCI.CurrentBlockHash
			output.Block.ParentBlockID = rsp.BCI.PreviousBlockHash
		} else {
			//fmt.Println("rsp.BCI is nil")
			return nil, fmt.Errorf("LedgerClient.QueryInfo response.BCI is ni")
		}
	} else {
		var rsp *cb.Block
		if len(input.BlockID) > 0 {
			rsp, err = afab.LedgerClient.QueryBlockByHash(input.BlockID)
		} else {
			rsp, err = afab.LedgerClient.QueryBlock(input.Height)
		}
		if err != nil {
			//fmt.Println(err.Error())
			return nil, err
		}
		if nil == rsp {
			//fmt.Println("rsp is nil")
			return nil, fmt.Errorf("LedgerClient.QueryBlock response is nil")
		} else if nil == rsp.Header {
			//fmt.Println("rsp.Header is nil")
			return nil, fmt.Errorf("LedgerClient.QueryBlock response.Header is nil")
		} else {
			//fmt.Println("rsp.Header.Number", rsp.Header.Number)
			//fmt.Println("rsp.Header.DataHash", hex.EncodeToString(rsp.Header.DataHash))
			//fmt.Println("rsp.Header.PreviousHash", hex.EncodeToString(rsp.Header.PreviousHash))
			output.Block.IsStable = false //todo
			output.Block.BlockHeight = uint(rsp.Header.Number)
			output.Block.BlockID = rsp.Header.DataHash //todo
			output.Block.ParentBlockID = rsp.Header.PreviousHash
		}
	}
	return &output, nil
}

/*ICryptoCurrency*/
//获取某地址下持有某资产的数量,返回数量为该资产的最小单位
func (afab *AdaptorFabric) GetBalance(input *adaptor.GetBalanceInput) (*adaptor.GetBalanceOutput, error) {
	return nil, errors.New("not implement")
}

//获取某资产的小数点位数
func (afab *AdaptorFabric) GetAssetDecimal(asset *adaptor.GetAssetDecimalInput) (
	*adaptor.GetAssetDecimalOutput, error) {
	return nil, errors.New("not implement")
}

//创建一个转账交易，但是未签名
func (afab *AdaptorFabric) CreateTransferTokenTx(input *adaptor.CreateTransferTokenTxInput) (
	*adaptor.CreateTransferTokenTxOutput, error) {
	return nil, errors.New("not implement")
}

//获取某个地址对某种Token的交易历史,支持分页和升序降序排列
func (afab *AdaptorFabric) GetAddrTxHistory(input *adaptor.GetAddrTxHistoryInput) (
	*adaptor.GetAddrTxHistoryOutput, error) {
	return nil, errors.New("not implement")
}

//根据交易ID获得对应的转账交易
func (afab *AdaptorFabric) GetTransferTx(input *adaptor.GetTransferTxInput) (*adaptor.GetTransferTxOutput, error) {
	return nil, errors.New("not implement")
}

//创建一个多签地址，该地址必须要满足signCount个签名才能解锁
func (afab *AdaptorFabric) CreateMultiSigAddress(input *adaptor.CreateMultiSigAddressInput) (
	*adaptor.CreateMultiSigAddressOutput, error) {
	return nil, errors.New("not implement")
}

//构造一个从多签地址付出Token的交易
func (aerc20 *AdaptorFabric) CreateMultiSigPayoutTx(input *adaptor.CreateMultiSigPayoutTxInput) (
	*adaptor.CreateMultiSigPayoutTxOutput, error) {
	return nil, errors.New("not implement")
}

/*ISmartContract*/
//创建一个安装合约的交易，未签名 //fabric 产生安装合约模板交易， need implement
func (afab *AdaptorFabric) CreateContractInstallTx(input *adaptor.CreateContractInstallTxInput) (
	*adaptor.CreateContractInstallTxOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	err = InitResmgmt(afab)
	if err != nil {
		return nil, err
	}
	//创建请求之前,需要使用 gopackager.NewCCPackage 方法生成一个resource.CCPackage 对象,传递两个参数,一个是链码的路径(相对于工程的路径), 一个是GOPATH的路径.
	ccp, err := gopackager.NewCCPackage(string(input.Extra), afab.EnvGoPath)
	if err != nil {
		//fmt.Println(err.Error())
		return nil, err
	}
	//fmt.Println("435", afab.EnvGoPath, string(input.Extra))
	//安装链码之前需要创建请求
	installCCRequest := resmgmt.InstallCCRequest{
		Name:string(input.Contract),//链码名称
		Path:string(input.Extra),//链码在工程中的路径
		Version:"0",
		Package:ccp,
	}
	//安装链码
	//result, err := afab.ResClient.InstallCC(installCCRequest)
	result, err := afab.ResClient.InstallCCZxl(installCCRequest)
	if err != nil {
		//fmt.Println("ResClient.InstallCCZxl", err.Error())
		return nil, err
	}
	//fmt.Println(len(rspArr))
	//fmt.Println(result)
	//fmt.Println("InstallCC return", result.TxnID)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		//fmt.Println("json.Marshal(result)", err.Error())
		return nil, err
	}
	var output adaptor.CreateContractInstallTxOutput // todo
	output.RawTransaction = resultJSON
	return &output, nil
}

//查询合约安装的结果的交易 //fabric 获取安装模板交易， not implement
func (afab *AdaptorFabric) GetContractInstallTx(input *adaptor.GetContractInstallTxInput) (
	*adaptor.GetContractInstallTxOutput, error) {
	return nil, errors.New("not implement")
}

//初始化合约实例 //fabric 产生合约实例化交易， need implement
func (afab *AdaptorFabric) CreateContractInitialTx(input *adaptor.CreateContractInitialTxInput) (
	*adaptor.CreateContractInitialTxOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	err = InitResmgmt(afab)
	if err != nil {
		return nil, err
	}

	//创建请求之前,需要生成一个*cb.SignaturePolicyEnvelope类型的对象,
	// 使用 third_party/github.com/hyperledger/fabric/common/cauthdsl/cauthdsl_builder.go
	// 文件中的方法即可,提供了好几个方法, 使用任意一个即可.
	// 这里使用 SignedByAnyMember 方法: 需要传入所属组织ID
	ccpolity := cauthdsl.SignedByAnyMember([]string{afab.OrgID})//不能为空

	//实例化链码之前需要创建请求
	req := resmgmt.InstantiateCCRequest{
		Name:string(input.Contract),//链码名称
		Path:string(input.Extra),//链码在工程中的路径
		Version:"0",
		Args:input.Args,
		Policy:ccpolity,
	}

	//实例化链码
	//result,err := afab.ResClient.InstantiateCC(afab.ChannelID, req)
	result,err := afab.ResClient.InstantiateCCCreateInitZxl(afab.ChannelID, req)
	if err != nil {
		//fmt.Println("resClient.InstantiateCC",err.Error())
		return nil, err
	}
	//fmt.Println("InstantiateCC return", result.TransactionID)
	//fmt.Println("InstantiateCC return", result.TxnID)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		//fmt.Println("json.Marshal(result)", err.Error())
		return nil, err
	}
	var output adaptor.CreateContractInitialTxOutput // todo
	output.RawTransaction = resultJSON
	return &output,nil
}

//查询初始化合约实例的交易 //fabric 获取合约实例化交易， not implement
func (afab *AdaptorFabric) GetContractInitialTx(input *adaptor.GetContractInitialTxInput) (
	*adaptor.GetContractInitialTxOutput, error) {
	return nil, errors.New("not implement")
}

//调用合约方法 //fabric 产生合约调用交易， need implement
func (afab *AdaptorFabric) CreateContractInvokeTx(input *adaptor.CreateContractInvokeTxInput) (
	*adaptor.CreateContractInvokeTxOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	err = InitChannel(afab)
	if err != nil {
		return nil, err
	}
	//写入之前,要创建请求:
	req := channel.Request{
		ChaincodeID: input.ContractAddress, // 链码名称
		Fcn: input.Function, // 方法名:  createCar
		Args:input.Args, // 传递的参数,是一个二维字符切片
	}

	//pmsp.Context().SigningManager().Sign()
	//使用 pkg/client/channel/chclient.go 中的 Execute()方法,来进行数据写入的操作:
	//rsp, err := afab.ChannelClient.Execute(req)
	rsp, err := afab.ChannelClient.ExecuteZxl(req)
	if err != nil {
		//fmt.Println(err.Error())
		return nil, err
	}
	//fmt.Println("Execute retrun", rsp.TransactionID)

	if nil == rsp.Proposal {
		return nil, fmt.Errorf("CreateContractInvokeTx failed")
	}
	resultJSON, err := json.Marshal(rsp.Proposal)
	if err != nil {
		//fmt.Println("json.Marshal(result)", err.Error())
		return nil, err
	}
	var output adaptor.CreateContractInvokeTxOutput // todo
	output.RawTransaction = resultJSON
	output.Extra = []byte(rsp.TransactionID)
	return &output, nil
}

//查询调用合约方法的交易 //fabric 获取合约调用交易，查询交易细节， need implement
func (afab *AdaptorFabric) GetContractInvokeTx(input *adaptor.GetContractInvokeTxInput) (
	*adaptor.GetContractInvokeTxOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	err = InitLedger(afab)
	if err != nil {
		return nil, err
	}
	rsp,err := afab.LedgerClient.QueryTransaction(fab.TransactionID(hex.EncodeToString(input.TxID)))
	if err != nil {
		//fmt.Println(err.Error())
		return nil, err
	}
	if rsp == nil{
		//fmt.Println("rsp is nil")
		return nil, err
	}
	//fmt.Println("rsp.ValidationCode",rsp.ValidationCode)
	var output adaptor.GetContractInvokeTxOutput
	if rsp.TransactionEnvelope == nil{
		//fmt.Println("rsp.TransactionEnvelope is nil")
		return nil, fmt.Errorf("rsp.TransactionEnvelope is nil")
	}
	txDetails, err := GetTransactionInfo(rsp)
	if err != nil {
		//fmt.Println("GetTransactionInfoFromData failed", err.Error())
		return nil, err
	}
	//fmt.Println("txDetails.TransactionId", txDetails.TransactionId)
	//fmt.Println("txDetails.Args", txDetails.Args)
	txID,_ := hex.DecodeString(string(txDetails.TransactionId))
	output.TxID = txID
	output.TargetAddress = txDetails.ChaincodeID
	argsJSON,_:=json.Marshal(txDetails.Args)
	output.TxRawData = argsJSON

	if 0 == rsp.ValidationCode {
		output.IsSuccess = true
	} else {
		output.IsSuccess = false
	}
	output.UpdateStateSuccess = output.IsSuccess//todo
	if nil != rsp.TransactionEnvelope {
		output.InvokeResult = rsp.TransactionEnvelope.Payload
	}

	return &output, nil
}

//调用合约的查询方法 //fabric 合约查询
func (afab *AdaptorFabric) QueryContract(input *adaptor.QueryContractInput) (
	*adaptor.QueryContractOutput, error) {
	err := InitSDK(afab)
	if err != nil {
		return nil, err
	}
	err = InitChannel(afab)
	if err != nil {
		return nil, err
	}
	//使用 pkg/client/channel/chclient.go 中的 Query()方法,来进行数据查询的操作:
	// 查询之前,同样需要创建请求.
	reqQuery := channel.Request{
		ChaincodeID: input.ContractAddress,
		Fcn: input.Function,
		Args: input.Args,
	}
	resp, err := afab.ChannelClient.Query(reqQuery)
	if err != nil {
		//fmt.Println("channelClient.Query",err.Error())
		return nil, err
	}
	//fmt.Println(resp.TransactionID)
	//fmt.Println(string(resp.Payload))

	var output adaptor.QueryContractOutput
	output.QueryResult = resp.Payload
	return &output, nil
}
