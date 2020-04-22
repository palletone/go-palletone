package fabricadaptor

import (
	_ "fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/palletone/fabric-adaptor/protoutil"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// TransactionDetail获取了交易的具体信息
type TransactionDetail struct {
	TransactionId string
	CreateTime    string
	ChaincodeID   string
	Args          []string
}

//从SDK中Block.BlockDara.Data中提取交易具体信息
func GetTransactionInfo(ptx *pb.ProcessedTransaction) (*TransactionDetail, error) {
	env := ptx.TransactionEnvelope

	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		return nil, err
	}
	channelHeaderBytes := payl.Header.ChannelHeader
	channelHeader := &cb.ChannelHeader{}
	if err := proto.Unmarshal(channelHeaderBytes, channelHeader); err != nil {
		return nil, errors.Wrap(err, "error Unmarshal ChannelHeader from payload")
	}

	//chaincodeAction,err:=putilsI.GetActionFromEnvelopeMsg(env)
	//_=chaincodeAction

	var args []string
	tx, err := protoutil.UnmarshalTransaction(payl.Data)
	if err != nil {
		return nil, err
	}
	chaincodeActionPayload,_, err := protoutil.GetPayloads(tx.Actions[0])
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling chaincode action payload")
	}
	propPayload := &pb.ChaincodeProposalPayload{}
	if err := proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, propPayload); err != nil {
		return nil, errors.Wrap(err, "error Unmarshal ChaincodeProposalPayload from payload")
	}
	invokeSpec := &pb.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(propPayload.Input, invokeSpec)
	if err != nil {
		return nil, errors.Wrap(err, "error Unmarshal ChaincodeInvocationSpec from payload")
	}
	if nil == invokeSpec.ChaincodeSpec {
		return nil, errors.Wrap(err, "invokeSpec.ChaincodeSpec is nil")
	}
	for _, v := range invokeSpec.ChaincodeSpec.Input.Args {
		args = append(args, string(v))
	}

	result := &TransactionDetail{
		TransactionId: channelHeader.TxId,
		ChaincodeID: invokeSpec.ChaincodeSpec.ChaincodeId.Name,
		Args:          args,
		CreateTime:    time.Unix(channelHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05"),
	}
	return result, nil
}
