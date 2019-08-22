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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package shim

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"
	"github.com/palletone/go-palletone/common/log"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/pkg/errors"
)

// PeerChaincodeStream interface for stream between Peer and chaincode instance.
type PeerChaincodeStream interface {
	Send(*pb.ChaincodeMessage) error
	Recv() (*pb.ChaincodeMessage, error)
	CloseSend() error
}

type nextStateInfo struct {
	msg      *pb.ChaincodeMessage
	sendToCC bool
}

func (handler *Handler) triggerNextState(msg *pb.ChaincodeMessage, send bool) {
	handler.nextState <- &nextStateInfo{msg, send}
}

// Handler handler implementation for shim side of chaincode.
type Handler struct {
	sync.RWMutex
	//shim to peer grpc serializer. User only in serialSend
	serialLock sync.Mutex
	To         string
	ChatStream PeerChaincodeStream
	FSM        *fsm.FSM
	cc         Chaincode
	// Multiple queries (and one transaction) with different txids can be executing in parallel for this chaincode
	// responseChannel is the channel on which responses are communicated by the shim to the chaincodeStub.
	responseChannel map[string]chan pb.ChaincodeMessage
	nextState       chan *nextStateInfo
}

//serialSend serializes msgs so gRPC will be happy
func (handler *Handler) serialSend(msg *pb.ChaincodeMessage) error {
	handler.serialLock.Lock()
	defer handler.serialLock.Unlock()

	err := handler.ChatStream.Send(msg)

	return err
}

//serialSendAsync serves the same purpose as serialSend (serialize msgs so gRPC will
//be happy). In addition, it is also asynchronous so send-remoterecv--localrecv loop
//can be nonblocking. Only errors need to be handled and these are handled by
//communication on supplied error channel. A typical use will be a non-blocking or
//nil channel
func (handler *Handler) serialSendAsync(msg *pb.ChaincodeMessage, errc chan error) {
	go func() {
		err := handler.serialSend(msg)
		if errc != nil {
			errc <- err
		}
	}()
}

//transaction context id should be composed of chainID and txid. While
//needed for CC-2-CC, it also allows users to concurrently send proposals
//with the same TXID to a CC on two multiple channels
func (handler *Handler) getTxCtxId(chainID string, txid string) string {
	return chainID + txid
}

func (handler *Handler) createChannel(channelID, txid string) (chan pb.ChaincodeMessage, error) {
	handler.Lock()
	defer handler.Unlock()
	if handler.responseChannel == nil {
		return nil, errors.Errorf("[%s]cannot create response channel", shorttxid(txid))
	}
	txCtxID := handler.getTxCtxId(channelID, txid)
	if handler.responseChannel[txCtxID] != nil {
		return nil, errors.Errorf("[%s]channel exists", shorttxid(txCtxID))
	}
	c := make(chan pb.ChaincodeMessage)
	handler.responseChannel[txCtxID] = c
	return c, nil
}

func (handler *Handler) sendChannel(msg *pb.ChaincodeMessage) error {
	handler.Lock()
	defer handler.Unlock()
	if handler.responseChannel == nil {
		return errors.Errorf("[%s]Cannot send message response channel", shorttxid(msg.Txid))
	}
	txCtxID := handler.getTxCtxId(msg.ChannelId, msg.Txid)
	if handler.responseChannel[txCtxID] == nil {
		return errors.Errorf("[%s]sendChannel does not exist", shorttxid(msg.Txid))
	}

	log.Debugf("[%s]before send", shorttxid(msg.Txid))
	handler.responseChannel[txCtxID] <- *msg
	log.Debugf("[%s]after send", shorttxid(msg.Txid))

	return nil
}

//sends a message and selects
func (handler *Handler) sendReceive(msg *pb.ChaincodeMessage, c chan pb.ChaincodeMessage) (pb.ChaincodeMessage,
	error) {
	errc := make(chan error, 1)
	handler.serialSendAsync(msg, errc)

	//the serialsend above will send an err or nil
	//the select filters that first error(or nil)
	//and continues to wait for the response
	//it is possible that the response triggers first
	//in which case the errc obviously worked and is
	//ignored
	for {
		select {
		case err := <-errc:
			if err == nil {
				continue
			}
			//would have been logged, return false
			return pb.ChaincodeMessage{}, err
		case outmsg, val := <-c:
			if !val {
				return pb.ChaincodeMessage{}, errors.New("unexpected failure on receive")
			}
			return outmsg, nil
		}
	}
}

func (handler *Handler) deleteChannel(channelID, txid string) {
	handler.Lock()
	defer handler.Unlock()
	if handler.responseChannel != nil {
		txCtxID := handler.getTxCtxId(channelID, txid)
		delete(handler.responseChannel, txCtxID)
	}
}

// beforeRegistered is called to handle the REGISTERED message.
func (handler *Handler) beforeRegistered(e *fsm.Event) {
	if _, ok := e.Args[0].(*pb.ChaincodeMessage); !ok {
		e.Cancel(errors.New("Received unexpected message type"))
		return
	}
	log.Debugf("Received %s, ready for invocations", pb.ChaincodeMessage_REGISTERED)
}

// handleInit handles request to initialize chaincode.
func (handler *Handler) handleInit(msg *pb.ChaincodeMessage) {
	// The defer followed by triggering a go routine dance is needed to ensure that the previous state transition
	// is completed before the next one is triggered. The previous state transition is deemed complete only when
	// the beforeInit function is exited. Interesting bug fix!!
	go func() {
		var nextStateMsg *pb.ChaincodeMessage

		send := true

		defer func() {
			handler.triggerNextState(nextStateMsg, send)
		}()

		errFunc := func(err error, payload []byte, ce *pb.ChaincodeEvent, errFmt string,
			args ...interface{}) *pb.ChaincodeMessage {
			if err != nil {
				// Send ERROR message to chaincode support and change state
				if payload == nil {
					payload = []byte(err.Error())
				}
				log.Errorf(errFmt, args...)
				return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: msg.Txid,
					ChaincodeEvent: ce, ChannelId: msg.ChannelId}
			}
			return nil
		}
		// Get the function and args from Payload
		input := &pb.ChaincodeInput{}
		unmarshalErr := proto.Unmarshal(msg.Payload, input)
		if nextStateMsg = errFunc(unmarshalErr, nil, nil, "[%s]Incorrect payload format. "+
			"Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Call chaincode's Run
		// Create the ChaincodeStub which the chaincode can use to callback
		stub := new(ChaincodeStub)
		err := stub.init(handler, msg.ContractId, msg.ChannelId, msg.Txid, input, msg.Proposal)
		if nextStateMsg = errFunc(err, nil, stub.chaincodeEvent, "[%s]Init get error response. "+
			"Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}
		for i, a := range stub.args {
			fmt.Println(i, a)
		}
		res := pb.Response{}
		if len(input.Args) != 0 {
			log.Infof("user contract deploy")
			res = handler.cc.Init(stub)
		} else {
			log.Infof("user contract restart")
			res = pb.Response{
				Status:  OK,
				Message: "Restart container",
				Payload: nil,
			}
		}
		log.Debugf("[%s]Init get response status: %d, payload len: %d", shorttxid(msg.Txid), res.Status,
			len(res.Payload))

		if res.Status >= ERROR {
			err = errors.New(res.Message)
			if nextStateMsg = errFunc(err, []byte(res.Message), stub.chaincodeEvent, "[%s]Init get error "+
				"response. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
				return
			}
		}

		resBytes, err := proto.Marshal(&res)
		if nextStateMsg = errFunc(err, nil, stub.chaincodeEvent, "[%s]Init marshal response error. "+
			"Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Send COMPLETED message to chaincode support and change state
		nextStateMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Payload: resBytes, Txid: msg.Txid,
			ChaincodeEvent: stub.chaincodeEvent, ChannelId: stub.ChannelId}
		log.Debugf("[%s]Init succeeded. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_COMPLETED)
	}()
}

// beforeInit will initialize the chaincode if entering init from established.
func (handler *Handler) beforeInit(e *fsm.Event) {
	log.Debugf("Entered state %s", handler.FSM.Current())
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(errors.New("received unexpected message type"))
		return
	}
	log.Debugf("[%s]Received %s, initializing chaincode", shorttxid(msg.Txid), msg.Type.String())
	if msg.Type.String() == pb.ChaincodeMessage_INIT.String() {
		// Call the chaincode's Run function to initialize
		handler.handleInit(msg)
	}
}

// handleTransaction Handles request to execute a transaction.
func (handler *Handler) handleTransaction(msg *pb.ChaincodeMessage) {
	// The defer followed by triggering a go routine dance is needed to ensure that the previous state transition
	// is completed before the next one is triggered. The previous state transition is deemed complete only when
	// the beforeInit function is exited. Interesting bug fix!!
	go func() {
		//better not be nil
		var nextStateMsg *pb.ChaincodeMessage

		send := true

		defer func() {
			handler.triggerNextState(nextStateMsg, send)
		}()

		errFunc := func(err error, ce *pb.ChaincodeEvent, errStr string, args ...interface{}) *pb.ChaincodeMessage {
			if err != nil {
				payload := []byte(err.Error())
				log.Errorf(errStr, args...)
				return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: msg.Txid,
					ChaincodeEvent: ce, ChannelId: msg.ChannelId}
			}
			return nil
		}

		// Get the function and args from Payload
		input := &pb.ChaincodeInput{}
		unmarshalErr := proto.Unmarshal(msg.Payload, input)
		if nextStateMsg = errFunc(unmarshalErr, nil, "[%s]Incorrect payload format. Sending %s",
			shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Call chaincode's Run
		// Create the ChaincodeStub which the chaincode can use to callback
		stub := new(ChaincodeStub)
		err := stub.init(handler, msg.ContractId, msg.ChannelId, msg.Txid, input, msg.Proposal)
		if nextStateMsg = errFunc(err, stub.chaincodeEvent, "[%s]Transaction execution failed. Sending %s",
			shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		res := handler.cc.Invoke(stub)

		// Endorser will handle error contained in Response.
		resBytes, err := proto.Marshal(&res)
		if nextStateMsg = errFunc(err, stub.chaincodeEvent, "[%s]Transaction execution failed. Sending %s",
			shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Send COMPLETED message to chaincode support and change state
		log.Debugf("[%s]Transaction completed. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_COMPLETED)
		nextStateMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Payload: resBytes, Txid: msg.Txid,
			ChaincodeEvent: stub.chaincodeEvent, ChannelId: stub.ChannelId}
	}()
}

// beforeTransaction will execute chaincode's Run if coming from a TRANSACTION event.
func (handler *Handler) beforeTransaction(e *fsm.Event) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(errors.New("Received unexpected message type"))
		return
	}
	log.Debugf("[%s]Received %s, invoking transaction on chaincode(Src:%s, Dst:%s)", shorttxid(msg.Txid),
		msg.Type.String(), e.Src, e.Dst)
	if msg.Type.String() == pb.ChaincodeMessage_TRANSACTION.String() {
		// Call the chaincode's Run function to invoke transaction
		handler.handleTransaction(msg)
	}
}

// afterResponse is called to deliver a response or error to the chaincode stub.
func (handler *Handler) afterResponse(e *fsm.Event) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(errors.New("received unexpected message type"))
		return
	}

	if err := handler.sendChannel(msg); err != nil {
		log.Errorf("[%s]error sending %s (state:%s): %+v", shorttxid(msg.Txid), msg.Type,
			handler.FSM.Current(), err)
	} else {
		log.Debugf("[%s]Received %s, communicated (state:%s)", shorttxid(msg.Txid), msg.Type,
			handler.FSM.Current())
	}
}

func (handler *Handler) afterError(e *fsm.Event) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(errors.New("Received unexpected message type"))
		return
	}

	/* TODO- revisit. This may no longer be needed with the serialized/streamlined messaging model
	 * There are two situations in which the ERROR event can be triggered:
	 * 1. When an error is encountered within handleInit or handleTransaction - some issue at the chaincode side; In this case there will be no responseChannel and the message has been sent to the peer.
	 * 2. The chaincode has initiated a request (get/put/del state) to the peer and is expecting a response on the responseChannel; If ERROR is received from peer, this needs to be notified on the responseChannel.
	 */
	if err := handler.sendChannel(msg); err == nil {
		log.Debugf("[%s]Error received from peer %s, communicated(state:%s)", shorttxid(msg.Txid),
			msg.Type, handler.FSM.Current())
	}
}

// callPeerWithChaincodeMsg sends a chaincode message (for e.g., GetState along with the key) to the peer for a given txid
// and receives the response.
func (handler *Handler) callPeerWithChaincodeMsg(msg *pb.ChaincodeMessage, channelID,
	txid string) (pb.ChaincodeMessage, error) {
	// Create the channel on which to communicate the response from the peer
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(channelID, txid); err != nil {
		return pb.ChaincodeMessage{}, err
	}

	defer handler.deleteChannel(channelID, txid)

	return handler.sendReceive(msg, respChan)
}

// TODO: Implement a method to get multiple keys at a time [FAB-1244]
// handleGetState communicates with the peer to fetch the requested state information from the ledger.
func (handler *Handler) handleGetState(collection string, key string, contractid []byte, channelId string,
	txid string) ([]byte, error) {
	// Construct payload for GET_STATE

	payloadBytes, _ := proto.Marshal(&pb.GetState{Collection: collection, ContractId: contractid, Key: key})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE)

	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending GET_STATE", shorttxid(txid)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]GetState received payload %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]GetState received error %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

// TODO: Implement a method to get multiple keys at a time [FAB-1244]
// handleGetState communicates with the peer to fetch the requested state information from the ledger.
func (handler *Handler) handelGetStateByPrefix(prefix string, contractId []byte, channelId string,
	txid string) ([]*modules.KeyValue, error) {
	// Construct payload for GET_STATE

	payloadBytes, _ := proto.Marshal(&pb.GetStateByPrefix{ContractId: contractId, Prefix: prefix})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_PREFIX, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE_BY_PREFIX)

	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending GET_STATE_BY_PREFIX",
			shorttxid(txid)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]GetState received payload %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		rows := []*modules.KeyValue{}
		err = json.Unmarshal(responseMsg.Payload, &rows)
		return rows, err
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]GetState received error %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}
func (handler *Handler) handleGetTimestamp(collection string, rangeNumber uint32, contractid []byte,
	channelId string, txid string) ([]byte, error) {
	// Construct payload for GET_STATE

	payloadBytes, _ := proto.Marshal(&pb.GetTimestamp{Collection: collection, RangeNumber: rangeNumber})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_TIMESTAMP, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_TIMESTAMP)

	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending GET_STATE", shorttxid(txid)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]GetState received payload %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]GetState received error %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

func (handler *Handler) handleGetTokenBalance(address string, token *modules.Asset, contractid []byte,
	channelId string, txid string) ([]*modules.InvokeTokens, error) {
	par := &pb.GetTokenBalance{Address: address}
	if token != nil {
		par.Asset = token.String()
	}
	payloadBytes, _ := proto.Marshal(par)

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_TOKEN_BALANCE, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_TOKEN_BALANCE)

	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending GET_TOKEN_BALANCE", shorttxid(txid)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]GET_TOKEN_BALANCE received payload %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		tokenList := []*modules.InvokeTokens{}
		err = rlp.DecodeBytes(responseMsg.Payload, &tokenList)
		if err != nil {
			return nil, err
		}

		return tokenList, nil

	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]GET_TOKEN_BALANCE received error %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_ERROR)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}
func (handler *Handler) handlePayOutToken(collection string, addr string, invokeTokens *modules.AmountAsset,
	lockTime uint32, contractid []byte, channelId string, txid string) error {
	log.Debugf("collection %s", collection)
	// Construct payload for PAY_OUT_TOKEN
	//TODO Devin
	payloadBytes, _ := proto.Marshal(&pb.PayOutToken{Asset: invokeTokens.Asset.Bytes(), Amount: invokeTokens.Amount,
		Address: addr, Locktime: lockTime})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PAY_OUT_TOKEN, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_PAY_OUT_TOKEN)

	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("[%s]error sending GET_STATE", shorttxid(txid)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]GetState received payload %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]GetState received error %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}
func (handler *Handler) handleDefineToken(tokenType byte, define []byte, creator string, contractid []byte,
	channelId string, txid string) error {
	par := &pb.DefineToken{TokenType: int32(tokenType), Define: define, Creator: creator}
	payloadBytes, _ := proto.Marshal(par)

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_DEFINE_TOKEN, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_DEFINE_TOKEN)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("[%s]error sending PUT_STATE", msg.Txid))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR,
			responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)

}
func (handler *Handler) handleSupplyToken(assetId []byte, uniqueId []byte, amt uint64, creator string,
	contractid []byte, channelId string, txid string) error {
	par := &pb.SupplyToken{AssetId: assetId, UniqueId: uniqueId, Amount: amt, Creator: creator}
	payloadBytes, _ := proto.Marshal(par)

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_SUPPLY_TOKEN, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_SUPPLY_TOKEN)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("[%s]error sending PUT_STATE", msg.Txid))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR,
			responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)

}

// TODO: Implement a method to set multiple keys at a time [FAB-1244]
// handlePutState communicates with the peer to put state information into the ledger.
func (handler *Handler) handlePutState(collection string, contractId []byte, key string, value []byte,
	channelId string, txid string) error {
	// Construct payload for PUT_STATE
	payloadBytes, _ := proto.Marshal(&pb.PutState{Collection: collection, ContractId: contractId, Key: key,
		Value: value})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_PUT_STATE)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("[%s]error sending PUT_STATE", msg.Txid))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR,
			responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

// handleOutCall communicates with the peer to put state information into the ledger.
func (handler *Handler) handleOutCall(collection string, outChainName string, method string, params []byte,
	channelId string, txid string) ([]byte, error) {
	// Construct payload for ChaincodeMessage_OUTCHAIN_CALL
	payloadBytes, _ := proto.Marshal(&pb.OutChainCall{Collection: collection, OutChainName: outChainName,
		Method: method, Params: params})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_OUTCHAIN_CALL, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_OUTCHAIN_CALL)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending OUTCHAIN_CALL", msg.Txid))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR,
			responseMsg.Payload)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

func (handler *Handler) handleSendJury(collection string, msgType uint32, consultContent []byte, myAnswer []byte,
	channelId string, txid string) ([]byte, error) {
	// Construct payload for PUT_STATE
	payloadBytes, _ := proto.Marshal(&pb.SendJury{Collection: collection, MsgType: msgType,
		ConsultContent: consultContent, MyAnswer: myAnswer})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_SEND_JURY, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_SEND_JURY)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending ChaincodeMessage_SEND_JURY",
			msg.Txid))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

func (handler *Handler) handleRecvJury(collection string, msgType uint32, consultContent []byte, timeout uint32,
	channelId string, txid string) ([]byte, error) {
	// Construct payload for PUT_STATE
	payloadBytes, _ := proto.Marshal(&pb.RecvJury{Collection: collection, MsgType: msgType,
		ConsultContent: consultContent, Timeout: timeout})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RECV_JURY, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_RECV_JURY)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error sending ChaincodeMessage_RECV_JURY",
			msg.Txid))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR,
			responseMsg.Payload)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

// handleDelState communicates with the peer to delete a key from the state in the ledger.
func (handler *Handler) handleDelState(collection string, contractId []byte, key string, channelId string,
	txid string) error {
	//payloadBytes, _ := proto.Marshal(&pb.GetState{Collection: collection, Key: key})
	payloadBytes, _ := proto.Marshal(&pb.DelState{Collection: collection, ContractId: contractId, Key: key})

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_DEL_STATE, Payload: payloadBytes, Txid: txid,
		ChannelId: channelId, ContractId: contractId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_DEL_STATE)

	// Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return errors.Errorf("[%s]error sending DEL_STATE %s", shorttxid(msg.Txid),
			pb.ChaincodeMessage_DEL_STATE)
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully deleted state", msg.Txid,
			pb.ChaincodeMessage_RESPONSE)
		return nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s. Payload: %s", msg.Txid, pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s or %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

func (handler *Handler) handleGetSystemConfig(channelId, txid string) (*modules.GlobalProperty, error) {
	//func (handler *Handler) handleGetSystemConfig(key, channelId, txid string) (string, error) {
	// Construct payload for PUT_STATE
	//payloadBytes, _ := proto.Marshal(&pb.KeyForSystemConfig{Key: key})
	payloadBytes, _ := proto.Marshal(&pb.KeyForSystemConfig{})
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_SYSTEM_CONFIG_REQUEST, Payload: payloadBytes,
		ChannelId: channelId, Txid: txid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_SYSTEM_CONFIG_REQUEST)

	//Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error GetSystemConfig ", msg.Txid))
	}

	//正确返回
	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		//Success response
		log.Debugf("[%s]Received %s. Successfully get deposit config", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)

		gp := &modules.GlobalProperty{}
		err = rlp.DecodeBytes(responseMsg.Payload, gp)
		if err != nil {
			log.Error("DecodeBytes ChainParameters err:", "error", err)
			return nil, err
		}
		return gp, nil
	}

	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE)
}

func (handler *Handler) handleGetContractAllState(channelId, txid string,
	contractid []byte) (map[string]*modules.ContractStateValue, error) {
	//定义一个pb.ChaincodeMessage_GET_ALL_SATE
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_CONTRACT_ALL_STATE, Payload: []byte(""),
		ChannelId: channelId, Txid: txid, ContractId: contractid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_CONTRACT_ALL_STATE)
	//Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error GetPayToContractTokens ", msg.Txid))
	}
	//正确返回
	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		//Success response
		log.Debugf("[%s]Received %s. Successfully get tokens of pay to contract ",
			shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		states := make(map[string]*modules.ContractStateValue)
		err = json.Unmarshal(responseMsg.Payload, &states)
		if err != nil {
			return nil, err
		}
		return states, nil
	}
	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE)
}

func (handler *Handler) createResponse(status int32, payload []byte) pb.Response {
	return pb.Response{Status: status, Payload: payload}
}

// handleInvokeChaincode communicates with the peer to invoke another chaincode.
func (handler *Handler) handleInvokeChaincode(chaincodeName string, args [][]byte, channelId string,
	txid string) pb.Response {
	//we constructed a valid object. No need to check for error
	payloadBytes, _ := proto.Marshal(&pb.ChaincodeSpec{ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input: &pb.ChaincodeInput{Args: args}})

	// Create the channel on which to communicate the response from validating peer
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(channelId, txid); err != nil {
		return handler.createResponse(ERROR, []byte(err.Error()))
	}

	defer handler.deleteChannel(channelId, txid)

	// Send INVOKE_CHAINCODE message to peer chaincode support
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_INVOKE_CHAINCODE, Payload: payloadBytes,
		Txid: txid, ChannelId: channelId}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_INVOKE_CHAINCODE)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return handler.createResponse(ERROR, []byte(fmt.Sprintf("[%s]error sending %s",
			shorttxid(msg.Txid), pb.ChaincodeMessage_INVOKE_CHAINCODE)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		log.Debugf("[%s]Received %s. Successfully invoked chaincode", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		respMsg := &pb.ChaincodeMessage{}
		if err = proto.Unmarshal(responseMsg.Payload, respMsg); err != nil {
			return handler.createResponse(ERROR, []byte(err.Error()))
		}
		if respMsg.Type == pb.ChaincodeMessage_COMPLETED {
			// Success response
			log.Debugf("[%s]Received %s. Successfully invoked chaincode", shorttxid(responseMsg.Txid),
				pb.ChaincodeMessage_RESPONSE)
			res := &pb.Response{}
			if err = proto.Unmarshal(respMsg.Payload, res); err != nil {
				return handler.createResponse(ERROR, []byte(err.Error()))
			}
			return *res
		}
		log.Errorf("[%s]Received %s. Error from chaincode", shorttxid(responseMsg.Txid), respMsg.Type.String())
		return handler.createResponse(ERROR, responseMsg.Payload)
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		log.Errorf("[%s]Received %s.", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return handler.createResponse(ERROR, responseMsg.Payload)
	}

	// Incorrect chaincode message received
	return handler.createResponse(ERROR, []byte(fmt.Sprintf("[%s]Incorrect chaincode message %s received."+
		" Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE,
		pb.ChaincodeMessage_ERROR)))
}

// handleMessage message handles loop for shim side of chaincode/peer stream.
func (handler *Handler) handleMessage(msg *pb.ChaincodeMessage) error {
	if msg.Type == pb.ChaincodeMessage_KEEPALIVE {
		// Received a keep alive message, we don't do anything with it for now
		// and it does not touch the state machine
		return nil
	}
	log.Debugf("[%s]Handling ChaincodeMessage of type: %s(state:%s)", shorttxid(msg.Txid), msg.Type,
		handler.FSM.Current())
	if handler.FSM.Cannot(msg.Type.String()) {
		err := errors.Errorf("[%s]chaincode handler FSM cannot handle message (%s) with payload size (%d) "+
			"while in state: %s", msg.Txid, msg.Type.String(), len(msg.Payload), handler.FSM.Current())
		handler.serialSend(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte(err.Error()),
			Txid: msg.Txid, ChannelId: msg.ChannelId})
		return err
	}
	err := handler.FSM.Event(msg.Type.String(), msg)
	return filterError(err)
}

// 根据证书ID获得证书字节数据
func (handler *Handler) handleGetCertState(key string, channelId string, txid string) (certBytes []byte, err error) {
	// Construct payload for PUT_STATE
	payloadBytes, _ := proto.Marshal(&pb.KeyForSystemConfig{Key: key})
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_CERT_STATE, Payload: payloadBytes, ChannelId: channelId,
		Txid: txid}
	log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_CERT_STATE)
	//Execute the request and get response
	responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)

	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("[%s]error GetRequesterCert ", msg.Txid))
	}
	//正确返回
	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		//Success response
		log.Debugf("[%s]Received %s. Successfully get cert bytes", shorttxid(responseMsg.Txid),
			pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}
	// Incorrect chaincode message received
	return nil, errors.Errorf("[%s]incorrect chaincode message %s received. Expecting %s",
		shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE)
}

// 获得root ca
func (handler *Handler) handleGetCACert(channelID string, txid string) (caCert *x509.Certificate, err error) {
	val, err := handler.handleGetCertState("RootCABytes", channelID, txid)
	if err != nil {
		return nil, err
	}
	bytes, err := modules.LoadCertBytes(val)
	if err != nil {
		return nil, err
	}
	caCert, err = x509.ParseCertificate(bytes)
	if err != nil {
		return nil, err
	}
	return caCert, nil
}

// 获取root ca holder
func (handler *Handler) handleGetCAHolder(channelID string, txid string) (caHolder string, err error) {
	val, err := handler.handleGetCertState("RootCAHolder", channelID, txid)
	if err != nil {
		return "", err
	}
	if len(val) <= 0 {
		return "", fmt.Errorf("query no ca holder")
	}
	return string(val), nil
}

// 获取证书链
func (handler *Handler) handleGetCertChain(rootIssuer string, cert *x509.Certificate,
	channelID string, txid string) (intermediates []*x509.Certificate, holders []string, err error) {
	intermediates = []*x509.Certificate{}
	holders = []string{}
	subject := cert.Issuer.CommonName
	for {
		key := dagConstants.CERT_SUBJECT_SYMBOL + subject
		val, err := handler.handleGetCertState(key, channelID, txid)
		if err != nil {
			return nil, nil, err
		}
		// query chain done
		if val == nil {
			break
		}
		// parse certid
		certID := big.Int{}
		certID.SetBytes(val)
		// get cert bytes
		key = dagConstants.CERT_BYTES_SYMBOL + certID.String()
		bytes, err := handler.handleGetCertState(key, channelID, txid)
		if err != nil {
			return nil, nil, err
		}
		certDBInfo := modules.CertBytesInfo{}
		if err := json.Unmarshal(bytes, &certDBInfo); err != nil {
			return nil, nil, err
		}
		// parse cert
		newCert, err := x509.ParseCertificate(certDBInfo.Raw)
		if err != nil {
			return nil, nil, err
		}
		intermediates = append(intermediates, newCert)
		holders = append(holders, certDBInfo.Holder)
		subject = newCert.Issuer.CommonName
		if subject == rootIssuer {
			break
		}
	}
	return intermediates, holders, nil
}

// 获取证书的吊销时间
func (handler *Handler) handleGetCertRevocationTime(key string, channelID string,
	txid string) (t time.Time, err error) {
	data, err := handler.handleGetCertState(key, channelID, txid)
	if err != nil {
		return time.Time{}, err
	}
	if len(data) <= 0 {
		return t, fmt.Errorf("have no time")
	}
	t = time.Time{}
	if err := t.UnmarshalBinary(data); err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (handler *Handler) handlerCheckCertValidation(caller string, certID []byte, channelId string,
	txid string) (bool, error) {
	intCertID := new(big.Int).SetBytes(certID)
	key := dagConstants.CERT_BYTES_SYMBOL + intCertID.String()

	// get root ca holder
	caHolder, err := handler.handleGetCAHolder(channelId, txid)
	if err != nil {
		return false, fmt.Errorf("get ca holder error(%s)", err.Error())
	}
	// check ca state
	caCert, err := handler.handleGetCACert(channelId, txid)
	if err != nil {
		return false, fmt.Errorf("query ca certificate error (%s)", err.Error())
	}
	if caCert.NotBefore.After(time.Now()) || caCert.NotAfter.Before(time.Now()) {
		return false, fmt.Errorf("ca certificate has expired")
	}
	// caller is ca holder
	if (caHolder == caller && caCert.SerialNumber.String() != intCertID.String()) || (intCertID.String() ==
		caCert.SerialNumber.String() && caHolder != caller) {
		return false, fmt.Errorf("you have no authority to call this certificate")
	} else if caller == caHolder && caCert.SerialNumber.String() == intCertID.String() {
		return true, nil
	}

	// get cert bytes
	resBytes, err := handler.handleGetCertState(key, channelId, txid)
	if err != nil {
		return false, fmt.Errorf("get certificate bytes by id error(%s)", err.Error())
	}
	if len(resBytes) <= 0 {
		return false, fmt.Errorf("certificate is not exist")
	}
	certDBInfo := modules.CertBytesInfo{}
	if err := json.Unmarshal(resBytes, &certDBInfo); err != nil {
		return false, fmt.Errorf("certificate bytes error")
	}
	// check caller
	if caller != certDBInfo.Holder {
		return false, fmt.Errorf("you have no authority to use this certificate")
	}
	// parse certificate
	cert, err := x509.ParseCertificate(certDBInfo.Raw)
	if err != nil {
		return false, fmt.Errorf("parse certificate error(%s)", err.Error())
	}
	// check revocation date, only user certificate could be used in contract
	key = dagConstants.CERT_SERVER_SYMBOL + caller + dagConstants.CERT_SPLIT_CH + cert.SerialNumber.String()
	revocation, err := handler.handleGetCertRevocationTime(key, channelId, txid)
	if err != nil {
		key = dagConstants.CERT_MEMBER_SYMBOL + caller + dagConstants.CERT_SPLIT_CH + cert.SerialNumber.String()
		revocation, err = handler.handleGetCertRevocationTime(key, channelId, txid)
		if err != nil {
			return false, fmt.Errorf("certificate is not existing")
		}
	}
	if revocation.Before(time.Now()) {
		return false, fmt.Errorf("certificate has been revocated at %s", revocation.String())
	}
	// Validity Period
	if cert.NotBefore.After(time.Now()) || cert.NotAfter.Before(time.Now()) {
		return false, fmt.Errorf("certificate has expired")
	}
	// check chain state ( get chain )
	intermediates, holders, err := handler.handleGetCertChain(caCert.Subject.CommonName, cert, channelId, txid)
	if err != nil {
		return false, fmt.Errorf("get certificate chain error (%s)", err.Error())
	}
	for index, c := range intermediates {
		if c.NotBefore.After(time.Now()) || c.NotAfter.Before(time.Now()) {
			return false, fmt.Errorf("intermediate certificate (%s) has expired", c.SerialNumber.String())
		}
		// check intermediate revocation
		key = dagConstants.CERT_SERVER_SYMBOL + holders[index] + dagConstants.CERT_SPLIT_CH + c.SerialNumber.String()
		tt, err := handler.handleGetCertRevocationTime(key, channelId, txid)
		if err != nil {
			return false, fmt.Errorf("get intermediate certificate revocation time error(%s)", err.Error())
		}
		if tt.Before(time.Now()) {
			return false, fmt.Errorf("intermediate certificate (%s) has been revocated",
				c.SerialNumber.String())
		}
	}
	// check certificate policy

	return true, nil
}

// filterError filters the errors to allow NoTransitionError and CanceledError to not propagate for cases where
// embedded Err == nil.
func filterError(errFromFSMEvent error) error {
	if errFromFSMEvent != nil {
		if noTransitionErr, ok := errFromFSMEvent.(*fsm.NoTransitionError); ok {
			if noTransitionErr.Err != nil {
				// Only allow NoTransitionError's, all others are considered true error.
				return errFromFSMEvent
			}
		}
		if canceledErr, ok := errFromFSMEvent.(*fsm.CanceledError); ok {
			if canceledErr.Err != nil {
				// Only allow NoTransitionError's, all others are considered true error.
				return canceledErr
				//t.Error("expected only 'NoTransitionError'")
			}
			log.Debugf("Ignoring CanceledError: %s", canceledErr)
		}
	}
	return nil
}
func shorttxid(txid string) string {
	if len(txid) < 8 {
		return txid
	}
	return txid[0:8]
}

// NewChaincodeHandler returns a new instance of the shim side handler.
func newChaincodeHandler(peerChatStream PeerChaincodeStream, chaincode Chaincode) *Handler {
	v := &Handler{
		ChatStream: peerChatStream,
		cc:         chaincode,
	}
	v.responseChannel = make(map[string]chan pb.ChaincodeMessage)
	v.nextState = make(chan *nextStateInfo)

	// Create the shim side FSM
	v.FSM = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: pb.ChaincodeMessage_REGISTERED.String(), Src: []string{"created"}, Dst: "established"},
			{Name: pb.ChaincodeMessage_READY.String(), Src: []string{"established"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_ERROR.String(), Src: []string{"init"}, Dst: "established"},
			{Name: pb.ChaincodeMessage_RESPONSE.String(), Src: []string{"init"}, Dst: "init"},
			{Name: pb.ChaincodeMessage_INIT.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_TRANSACTION.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_RESPONSE.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_ERROR.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_COMPLETED.String(), Src: []string{"init"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_COMPLETED.String(), Src: []string{"ready"}, Dst: "ready"},
		},
		fsm.Callbacks{
			"before_" + pb.ChaincodeMessage_REGISTERED.String():  func(e *fsm.Event) { v.beforeRegistered(e) },
			"after_" + pb.ChaincodeMessage_RESPONSE.String():     func(e *fsm.Event) { v.afterResponse(e) },
			"after_" + pb.ChaincodeMessage_ERROR.String():        func(e *fsm.Event) { v.afterError(e) },
			"before_" + pb.ChaincodeMessage_INIT.String():        func(e *fsm.Event) { v.beforeInit(e) },
			"before_" + pb.ChaincodeMessage_TRANSACTION.String(): func(e *fsm.Event) { v.beforeTransaction(e) },
		},
	)
	return v
}

//
//func (handler *Handler) handleGetQueryResult(collection string, query string, channelId string, txid string) (*pb.QueryResponse, error) {
// // Send GET_QUERY_RESULT message to peer chaincode support
// //we constructed a valid object. No need to check for error
// payloadBytes, _ := proto.Marshal(&pb.GetQueryResult{Collection: collection, Query: query})
//
// msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_QUERY_RESULT, Payload: payloadBytes, Txid: txid,
// ChannelId: channelId}
// log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_QUERY_RESULT)
//
// responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
// if err != nil {
//    return nil, errors.Errorf("[%s]error sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_QUERY_RESULT)
// }
//
// if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
//    // Success response
//    log.Debugf("[%s]Received %s. Successfully got range", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
//
//    executeQueryResponse := &pb.QueryResponse{}
//    if err = proto.Unmarshal(responseMsg.Payload, executeQueryResponse); err != nil {
//       return nil, errors.Errorf("[%s]unmarshal error", shorttxid(responseMsg.Txid))
//    }
//
//    return executeQueryResponse, nil
// }
// if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
//    // Error response
//    log.Errorf("[%s]Received %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
//    return nil, errors.New(string(responseMsg.Payload[:]))
// }
//
// // Incorrect chaincode message received
// return nil, errors.Errorf("incorrect chaincode message %s received. Expecting %s or %s", responseMsg.Type,
// pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
//}
/*
func (handler *Handler) handleGetHistoryForKey(key string, channelId string, txid string) (*pb.QueryResponse, error) {
   // Create the channel on which to communicate the response from validating peer
   var respChan chan pb.ChaincodeMessage
   var err error
   if respChan, err = handler.createChannel(channelId, txid); err != nil {
      return nil, err
   }

   defer handler.deleteChannel(channelId, txid)

   // Send GET_HISTORY_FOR_KEY message to peer chaincode support
   //we constructed a valid object. No need to check for error
   payloadBytes, _ := proto.Marshal(&pb.GetHistoryForKey{Key: key})

   msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_HISTORY_FOR_KEY, Payload: payloadBytes,
Txid: txid, ChannelId: channelId}
   log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_HISTORY_FOR_KEY)

   var responseMsg pb.ChaincodeMessage

   if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
      return nil, errors.Errorf("[%s]error sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_HISTORY_FOR_KEY)
   }

   if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
      // Success response
      log.Debugf("[%s]Received %s. Successfully got range", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)

      getHistoryForKeyResponse := &pb.QueryResponse{}
      if err = proto.Unmarshal(responseMsg.Payload, getHistoryForKeyResponse); err != nil {
         return nil, errors.Errorf("[%s]unmarshal error", shorttxid(responseMsg.Txid))
      }

      return getHistoryForKeyResponse, nil
   }
   if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
      // Error response
      log.Errorf("[%s]Received %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
      return nil, errors.New(string(responseMsg.Payload[:]))
   }

   // Incorrect chaincode message received
   return nil, errors.Errorf("incorrect chaincode message %s received. Expecting %s or %s", responseMsg.Type,
pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}
*/
/*
func (handler *Handler) handleGetStateByRange(collection, startKey, endKey string, channelId string,
txid string) (*pb.QueryResponse, error) {
   // Send GET_STATE_BY_RANGE message to peer chaincode support
   //we constructed a valid object. No need to check for error
   payloadBytes, _ := proto.Marshal(&pb.GetStateByRange{Collection: collection, StartKey: startKey, EndKey: endKey})

   msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_RANGE, Payload: payloadBytes,
Txid: txid, ChannelId: channelId}
   log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE_BY_RANGE)

   responseMsg, err := handler.callPeerWithChaincodeMsg(msg, channelId, txid)
   if err != nil {
      return nil, errors.Errorf("[%s]error sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE_BY_RANGE)
   }

   if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
      // Success response
      log.Debugf("[%s]Received %s. Successfully got range", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)

      rangeQueryResponse := &pb.QueryResponse{}
      if err = proto.Unmarshal(responseMsg.Payload, rangeQueryResponse); err != nil {
         return nil, errors.Errorf("[%s]GetStateByRangeResponse unmarshall error", shorttxid(responseMsg.Txid))
      }

      return rangeQueryResponse, nil
   }
   if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
      // Error response
      log.Errorf("[%s]Received %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
      return nil, errors.New(string(responseMsg.Payload[:]))
   }

   // Incorrect chaincode message received
   return nil, errors.Errorf("incorrect chaincode message %s received. Expecting %s or %s", responseMsg.Type,
pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

func (handler *Handler) handleQueryStateNext(id, channelId, txid string) (*pb.QueryResponse, error) {
   // Create the channel on which to communicate the response from validating peer
   var respChan chan pb.ChaincodeMessage
   var err error
   if respChan, err = handler.createChannel(channelId, txid); err != nil {
      return nil, err
   }

   defer handler.deleteChannel(channelId, txid)

   // Send QUERY_STATE_NEXT message to peer chaincode support
   //we constructed a valid object. No need to check for error
   payloadBytes, _ := proto.Marshal(&pb.QueryStateNext{Id: id})

   msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_NEXT, Payload: payloadBytes,
Txid: txid, ChannelId: channelId}
   log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_QUERY_STATE_NEXT)

   var responseMsg pb.ChaincodeMessage

   if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
      return nil, errors.Errorf("[%s]error sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_QUERY_STATE_NEXT)
   }

   if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
      // Success response
      log.Debugf("[%s]Received %s. Successfully got range", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)

      queryResponse := &pb.QueryResponse{}
      if err = proto.Unmarshal(responseMsg.Payload, queryResponse); err != nil {
         return nil, errors.Errorf("[%s]unmarshal error", shorttxid(responseMsg.Txid))
      }

      return queryResponse, nil
   }
   if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
      // Error response
      log.Errorf("[%s]Received %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
      return nil, errors.New(string(responseMsg.Payload[:]))
   }

   // Incorrect chaincode message received
   return nil, errors.Errorf("incorrect chaincode message %s received. Expecting %s or %s", responseMsg.Type,
pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}

func (handler *Handler) handleQueryStateClose(id, channelId, txid string) (*pb.QueryResponse, error) {
   // Create the channel on which to communicate the response from validating peer
   var respChan chan pb.ChaincodeMessage
   var err error
   if respChan, err = handler.createChannel(channelId, txid); err != nil {
      return nil, err
   }

   defer handler.deleteChannel(channelId, txid)

   // Send QUERY_STATE_CLOSE message to peer chaincode support
   //we constructed a valid object. No need to check for error
   payloadBytes, _ := proto.Marshal(&pb.QueryStateClose{Id: id})

   msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_CLOSE, Payload: payloadBytes,
Txid: txid, ChannelId: channelId}
   log.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_QUERY_STATE_CLOSE)

   var responseMsg pb.ChaincodeMessage

   if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
      return nil, errors.Errorf("[%s]error sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_QUERY_STATE_CLOSE)
   }

   if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
      // Success response
      log.Debugf("[%s]Received %s. Successfully got range", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)

      queryResponse := &pb.QueryResponse{}
      if err = proto.Unmarshal(responseMsg.Payload, queryResponse); err != nil {
         return nil, errors.Errorf("[%s]unmarshal error", shorttxid(responseMsg.Txid))
      }

      return queryResponse, nil
   }
   if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
      // Error response
      log.Errorf("[%s]Received %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
      return nil, errors.New(string(responseMsg.Payload[:]))
   }

   // Incorrect chaincode message received
   return nil, errors.Errorf("incorrect chaincode message %s received. Expecting %s or %s", responseMsg.Type,
pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)
}
*/
