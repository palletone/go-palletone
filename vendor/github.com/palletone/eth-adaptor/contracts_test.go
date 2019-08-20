package ethadaptor

import (
	"fmt"
	"testing"

	"github.com/palletone/adaptor"
)

func TestGenDeployContractTX(t *testing.T) {
	rpcParams := RPCParams{
		Rawurl: "https://ropsten.infura.io/", //"\\\\.\\pipe\\geth.ipc",
	}
	//multisig contract 2/3
	const PTNMapABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ptnToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmapPTN\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmap\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addrHex\",\"type\":\"address\"}],\"name\":\"encodeBase58\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_ptnhex\",\"type\":\"address\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getptnhex\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"
	const PTNMapBin = `0x6080604052633b9aca0060035534801561001857600080fd5b5060008054600160a060020a03191673a54880da9a63cdd2ddacf25af68daf31a1bcc0c9179055610d2e8061004e6000396000f3006080604052600436106100cf5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100d4578063095ea7b31461015e5780630ac74e721461019657806318160ddd146101c757806323b872dd146101ee578063313ce567146102185780634e11092f1461024357806370a08231146102645780638c5cecaa14610285578063927f526f146102a657806395d89b41146102c7578063a9059cbb146102dc578063dd62ed3e14610300578063e1a0cbd314610327575b600080fd5b3480156100e057600080fd5b506100e9610348565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561012357818101518382015260200161010b565b50505050905090810190601f1680156101505780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561016a57600080fd5b50610182600160a060020a036004351660243561037f565b604080519115158252519081900360200190f35b3480156101a257600080fd5b506101ab610388565b60408051600160a060020a039092168252519081900360200190f35b3480156101d357600080fd5b506101dc610397565b60408051918252519081900360200190f35b3480156101fa57600080fd5b50610182600160a060020a0360043581169060243516604435610435565b34801561022457600080fd5b5061022d61043e565b6040805160ff9092168252519081900360200190f35b34801561024f57600080fd5b506101ab600160a060020a0360043516610443565b34801561027057600080fd5b506101dc600160a060020a036004351661045e565b34801561029157600080fd5b506101ab600160a060020a03600435166104b3565b3480156102b257600080fd5b506100e9600160a060020a03600435166104ce565b3480156102d357600080fd5b506100e96104f3565b3480156102e857600080fd5b50610182600160a060020a036004351660243561052a565b34801561030c57600080fd5b506101dc600160a060020a036004358116906024351661037f565b34801561033357600080fd5b506100e9600160a060020a03600435166105df565b60408051808201909152600b81527f50544e204d617070696e67000000000000000000000000000000000000000000602082015281565b60005b92915050565b600054600160a060020a031681565b60008060009054906101000a9004600160a060020a0316600160a060020a03166318160ddd6040518163ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401602060405180830381600087803b15801561040457600080fd5b505af1158015610418573d6000803e3d6000fd5b505050506040513d602081101561042e57600080fd5b5051905090565b60009392505050565b600081565b600260205260009081526040902054600160a060020a031681565b600160a060020a0381811660009081526001602052604081205490911615156104aa57600160a060020a038281166000908152600260205260409020541615156104aa575060016104ae565b5060005b919050565b600160205260009081526040902054600160a060020a031681565b606081816104eb6104e66104e184610606565b6107de565b610977565b949350505050565b60408051808201909152600681527f50544e4d61700000000000000000000000000000000000000000000000000000602082015281565b33600090815260016020526040812054600160a060020a031615156105d7573360008181526001602090815260408083208054600160a060020a03891673ffffffffffffffffffffffffffffffffffffffff199182168117909255818552600284529382902080549094168517909355805186815290519293927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a3506001610382565b506000610382565b600160a060020a0380821660009081526001602052604090205460609161038291166104ce565b6040805160198082528183019092526060916c01000000000000000000000000840291839160009182918291906020820161032080388339505081519195506000918691508290811061065557fe5b906020010190600160f860020a031916908160001a905350600092505b60148360ff1610156106cb578460ff84166014811061068d57fe5b1a60f860020a02848460010160ff168151811015156106a857fe5b906020010190600160f860020a031916908160001a905350600190920191610672565b6040805160008082526bffffffffffffffffffffffff1988166001830152915160029283926015808201936020939092839003909101908290865af1158015610718573d6000803e3d6000fd5b5050506040513d602081101561072d57600080fd5b505160408051918252516020828101929091908190038201816000865af115801561075c573d6000803e3d6000fd5b5050506040513d602081101561077157600080fd5b50519150600090505b60048160ff1610156107d0578160ff82166020811061079557fe5b1a60f860020a02848260150160ff168151811015156107b057fe5b906020010190600160f860020a031916908160001a90535060010161077a565b8395505b5050505050919050565b6060806000806000808651600014156108075760408051602081019091526000815295506107d4565b6040805160288082526105208201909252906020820161050080388339019050509450600085600081518110151561083b57fe5b60ff90921660209283029091019091015260019350600092505b86518360ff16101561095257868360ff1681518110151561087257fe5b90602001015160f860020a900460f860020a0260f860020a900460ff169150600090505b8360ff168160ff16101561090757848160ff168151811015156108b557fe5b9060200190602002015160ff166101000282019150603a828115156108d657fe5b06858260ff168151811015156108e857fe5b60ff909216602092830290910190910152603a82049150600101610896565b600082111561094757603a8206858560ff1681518110151561092557fe5b60ff909216602092830290910190910152600190930192603a82049150610907565b826001019250610855565b61096c6109676109628787610ac4565b610b59565b610bef565b979650505050505050565b606080606060008085516002016040519080825280601f01601f1916602001820160405280156109b1578160200160208202803883390190505b508051909450849350600192507f500000000000000000000000000000000000000000000000000000000000000090849060009081106109ed57fe5b906020010190600160f860020a031916908160001a905350825160018301927f3100000000000000000000000000000000000000000000000000000000000000918591908110610a3957fe5b906020010190600160f860020a031916908160001a905350600090505b85518160ff161015610aba57858160ff16815181101515610a7357fe5b90602001015160f860020a900460f860020a028383806001019450815181101515610a9a57fe5b906020010190600160f860020a031916908160001a905350600101610a56565b5091949350505050565b60608060008360ff16604051908082528060200260200182016040528015610af6578160200160208202803883390190505b509150600090505b8360ff168160ff161015610b5157848160ff16815181101515610b1d57fe5b90602001906020020151828260ff16815181101515610b3857fe5b60ff909216602092830290910190910152600101610afe565b509392505050565b60608060008351604051908082528060200260200182016040528015610b89578160200160208202803883390190505b509150600090505b83518160ff161015610be8578351849060ff8316810360001901908110610bb457fe5b90602001906020020151828260ff16815181101515610bcf57fe5b60ff909216602092830290910190910152600101610b91565b5092915050565b606080600083516040519080825280601f01601f191660200182016040528015610c23578160200160208202803883390190505b509150600090505b83518160ff161015610be857606060405190810160405280603a81526020017f31323334353637383941424344454647484a4b4c4d4e5051525354555657585981526020017f5a6162636465666768696a6b6d6e6f707172737475767778797a000000000000815250848260ff16815181101515610ca557fe5b9060200190602002015160ff16815181101515610cbe57fe5b90602001015160f860020a900460f860020a02828260ff16815181101515610ce257fe5b906020010190600160f860020a031916908160001a905350600101610c2b5600a165627a7a72305820e3859334108c189e50db001d7a3605fb2d4ca63b2eae29be13fd188932d31a790029`
	//
	deployerAddr := "0x7d7116a8706ae08baa7f4909e26728fa7a5f0365"

	//
	var input adaptor.CreateContractInitialTxInput
	input.Address = deployerAddr
	input.Fee = &adaptor.AmountAsset{}
	input.Fee.Amount.SetString("21000000000000000", 10) //10000000000 10gwei*2100000
	input.Contract = []byte(PTNMapABI)
	input.Extra = Hex2Bytes(PTNMapBin[2:])
	//
	result, err := CreateContractInitialTx(&input, &rpcParams, NETID_TEST)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("%0x\n", result.RawTranaction)

		keyHex := "8e87ebb3b00565aaf3675e1f7d16ed68b300c7302267934f3831105b48e8a3e7"
		key := Hex2Bytes(keyHex)

		var input adaptor.SignTransactionInput
		input.PrivateKey = key
		//input.Transaction = Hex2Bytes("f9024981848203e883200b20946817cfb2c442693d850332c3b755b2342ec4afb280b902248c2e032100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000aaa919a7c465be9b053673c567d73be8603179630000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000407d7116a8706ae08baa7f4909e26728fa7a5f0365aaa919a7c465be9b053673c567d73be8603179636c7110482920e0af149a82189251f292a84148a85b7cd70d00000000000000000000000000000000000000000000000000000000000000417197961c5ae032ed6f33650f1f3a3ba111e8548a3dad14b3afa1cb6bc8f4601a6cb2b21aedcd575784e923942f3130f3290d56522ab2b28afca478e489426a4601000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041ae94b0e599ef0508ba7bec41db5b46d5a065b30d3d5c4b0a4c85ea2d4899d6607e80e3314ee0741049963d30fb3aceaa5506e13835a41ef54a8f44a04ef0f1e40100000000000000000000000000000000000000000000000000000000000000808080")
		input.Transaction = result.RawTranaction
		resultSign, err := SignTransaction(&input)
		if err != nil {
			fmt.Println("failed ", err.Error())
		} else {
			fmt.Printf("%x\n", resultSign.Signature)
			fmt.Printf("%x\n", resultSign.Extra)
		}
	}
}

func TestCreateContractInvokeTx(t *testing.T) {
	rpcParams := RPCParams{
		Rawurl: "https://ropsten.infura.io/", //"\\\\.\\pipe\\geth.ipc",
	}
	//multisig contract 2/3
	const PTNMapABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ptnToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmapPTN\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmap\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addrHex\",\"type\":\"address\"}],\"name\":\"encodeBase58\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_ptnhex\",\"type\":\"address\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getptnhex\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"
	//
	invokeAddr := "0x588eB98f8814aedB056D549C0bafD5Ef4963069C"

	//
	var input adaptor.CreateContractInvokeTxInput
	input.Address = invokeAddr
	input.Fee = &adaptor.AmountAsset{}
	input.Fee.Amount.SetString("21000000000000000", 10) //10000000000 10gwei*2100000
	input.ContractAddress = "0xef37aba8a6379a2aaad5a90f9b39c940265cda93"
	input.Function = "transfer"
	input.Args = append(input.Args, []byte("0x1a9ed32dec553511158595375d62a8aa8784bc5b"))
	input.Args = append(input.Args, []byte("1"))
	input.Extra = []byte(PTNMapABI)

	//
	result, err := CreateContractInvokeTx(&input, &rpcParams, NETID_TEST)
	if err != nil {
		fmt.Println("CreateContractInvokeTx failed: ", err.Error())
	} else {
		fmt.Printf("%0x\n", result.RawTranaction)

		keyHex := "BE2DA21D719E002A0035B52D36BA9137AFE7A67E4C92E0A342EC1632944CE806"
		key := Hex2Bytes(keyHex)

		var input adaptor.SignTransactionInput
		input.PrivateKey = key
		//input.Transaction = Hex2Bytes("f9024981848203e883200b20946817cfb2c442693d850332c3b755b2342ec4afb280b902248c2e032100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000aaa919a7c465be9b053673c567d73be8603179630000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000407d7116a8706ae08baa7f4909e26728fa7a5f0365aaa919a7c465be9b053673c567d73be8603179636c7110482920e0af149a82189251f292a84148a85b7cd70d00000000000000000000000000000000000000000000000000000000000000417197961c5ae032ed6f33650f1f3a3ba111e8548a3dad14b3afa1cb6bc8f4601a6cb2b21aedcd575784e923942f3130f3290d56522ab2b28afca478e489426a4601000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041ae94b0e599ef0508ba7bec41db5b46d5a065b30d3d5c4b0a4c85ea2d4899d6607e80e3314ee0741049963d30fb3aceaa5506e13835a41ef54a8f44a04ef0f1e40100000000000000000000000000000000000000000000000000000000000000808080")
		input.Transaction = result.RawTranaction
		resultSign, err := SignTransaction(&input)
		if err != nil {
			fmt.Println("failed ", err.Error())
		} else {
			fmt.Printf("%x\n", resultSign.Signature)
			fmt.Printf("%x\n", resultSign.Extra)
		}
	}
}

func TestQueryContract(t *testing.T) {
	rpcParams := RPCParams{
		Rawurl: "https://ropsten.infura.io/", //"\\\\.\\pipe\\geth.ipc",
	}
	//multisig contract 2/3
	const PTNMapABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ptnToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmapPTN\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmap\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addrHex\",\"type\":\"address\"}],\"name\":\"encodeBase58\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_ptnhex\",\"type\":\"address\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getptnhex\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"
	//
	invokeAddr := "0x588eB98f8814aedB056D549C0bafD5Ef4963069C"

	//
	var input adaptor.QueryContractInput
	//input.Address = invokeAddr
	//input.Fee = &adaptor.AmountAsset{}
	//input.Fee.Amount.SetString("21000000000000000", 10) //10000000000 10gwei*2100000

	//input.ContractAddress = "0xef37aba8a6379a2aaad5a90f9b39c940265cda93"
	//input.Function = "getptnhex"
	//input.Args = append(input.Args, []byte(invokeAddr))
	//input.Extra = []byte(PTNMapABI)

	const ERC20ABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"
	input.ContractAddress = "0xa54880da9a63cdd2ddacf25af68daf31a1bcc0c9"
	input.Function = "decimals"
	input.Extra = []byte(ERC20ABI)
	_ = invokeAddr

	//
	result, err := QueryContract(&input, &rpcParams)
	if err != nil {
		fmt.Println("QueryContract failed: ", err.Error())
	} else {
		fmt.Printf("%s\n", string(result.QueryResult))
	}
}
