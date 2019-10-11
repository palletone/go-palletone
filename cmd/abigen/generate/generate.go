package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

//
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_ = stub
	return shim.Success(nil)
}
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()
	_ = args
	switch f {
	case "add":
		t.Add(stub, 1, 2)
	case "getAdd":
		t.GetAdd(stub, 1, 2)
	}
	return shim.Success(nil)

}

func (t *SimpleChaincode) Add(stub shim.ChaincodeStubInterface, a int, b int) int {
	_ = stub
	return a + b
}
func (t *SimpleChaincode) GetAdd(stub shim.ChaincodeStubInterface, a int, b int) int {
	_ = stub
	return t.add(a, b)
}
func (t *SimpleChaincode) add(a int, b int) int {
	return a + b
}

////////////////////

type ABI_Input struct {
	Name string
	Type string
}
type ABI_Output struct {
	Name string
	Type string
}
type ABI_Function struct {
	Constant        bool
	Inputs          []ABI_Input
	Name            string
	Outputs         []ABI_Output
	Payable         bool
	StateMutability string
	Type            string
}

func GenerateABI(chainCode interface{}) (string, error) {
	val := reflect.ValueOf(chainCode)
	paramKind := val.Kind().String()
	if paramKind == "struct" {
		val = reflect.New(val.Type())
	} else if paramKind != "ptr" {
		return "", fmt.Errorf("Invalid chainCode input,need struct or ptr")
	}

	//
	numOfMethod := val.NumMethod()

	//
	allFunc := make([]ABI_Function, 0, numOfMethod)

	//
	iStructType := val.Type()

	//
	for i := 0; i < numOfMethod; i++ {
		iMethod := val.Method(i)
		//fmt.Println(iMethod.String())//proto

		funcName := iStructType.Method(i).Name
		//fmt.Println(iStructType.Method(i).Name)
		if funcName == "Init" || funcName == "Invoke" {
			continue
		}

		oneFunc := ABI_Function{Name: strings.ToLower(funcName[0:1]) + funcName[1:], Type: "function",
			Inputs: make([]ABI_Input, 0), Outputs: make([]ABI_Output, 0)}

		if strings.HasPrefix(funcName, "Query") || strings.HasPrefix(funcName, "Find") ||
			strings.HasPrefix(funcName, "Get") {
			oneFunc.Constant = true
			oneFunc.Payable = false
			oneFunc.StateMutability = "nonpayable"
		} else {
			oneFunc.Constant = false //todo more
		}

		iMethodType := iMethod.Type()
		//fmt.Print(iMethod.Type().NumIn(), " in params")
		numIn := iMethodType.NumIn()
		cntIgnore := 0
		for j := 0; j < numIn; j++ {
			inType := iMethodType.In(j).String()
			if strings.HasPrefix(inType, "shim") || strings.HasPrefix(inType, "peer") {
				cntIgnore += 1
				continue
			}
			oneInput := ABI_Input{Name: "Args" + fmt.Sprintf("%d", j-cntIgnore+1), Type: inType}
			//fmt.Print(", ", iMethodType.In(j))
			oneFunc.Inputs = append(oneFunc.Inputs, oneInput)
		}

		//fmt.Print(iMethodType.NumOut(), " out params")
		numOut := iMethodType.NumOut()
		cntIgnore = 0
		for j := 0; j < numOut; j++ {
			outType := iMethodType.Out(j).String()
			if outType == "error" {
				continue
			}
			if strings.HasPrefix(outType, "shim") || strings.HasPrefix(outType, "peer") {
				cntIgnore += 1
				continue
			}
			oneOutput := ABI_Output{Name: "Result" + fmt.Sprintf("%d", j-cntIgnore+1), Type: outType}
			//fmt.Print(", ", iMethodType.Out(j))
			oneFunc.Outputs = append(oneFunc.Outputs, oneOutput)
		}
		allFunc = append(allFunc, oneFunc)
	}

	abiJSON, err := json.Marshal(allFunc)
	if err != nil {
		return "", err
	}
	return string(abiJSON), nil
}

func main() {
	result, err := GenerateABI(&SimpleChaincode{})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result)
	}
}
