protoc -I=./core/vmContractPub/protos/peer -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src ./core/vmContractPub/protos/peer/chaincode_event.proto
protoc -I=./core/vmContractPub/protos/peer -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src ./core/vmContractPub/protos/peer/chaincode_shim.proto
protoc -I=./core/vmContractPub/protos/peer -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src ./core/vmContractPub/protos/peer/chaincode.proto
protoc -I=./core/vmContractPub/protos/peer -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src ./core/vmContractPub/protos/peer/proposal_response.proto
protoc -I=./core/vmContractPub/protos/peer -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src ./core/vmContractPub/protos/peer/proposal.proto
protoc -I=./core/vmContractPub/protos/peer -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src ./core/vmContractPub/protos/peer/query.proto