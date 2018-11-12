package createToken

import ("github.com/palletone/go-palletone/contracts/shim"
pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer")

type CreateTokenChainCode struct {
	TokenName string

}

func (t *CreateTokenChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return pb.Response{}
}

// Invoke gets the supplied key and if it exists, updates the key with the newly
// supplied value.
func (t *CreateTokenChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return pb.Response{}
}
