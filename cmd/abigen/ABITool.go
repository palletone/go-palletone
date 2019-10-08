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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const appendPackage = `
package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)
`
const appendMain = `
//import (
//	"encoding/json"
//	"fmt"
//	"reflect"
//	"strings"
//)

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
`

func GoRun(chainCodeName, fileContent string, mkdir bool) {
	path := ""
	if mkdir {
		path = "." + string(os.PathSeparator) + "ABITemp"
		err := os.Mkdir(path, 0600)
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(path)
		path = path + string(os.PathSeparator)
	}
	//
	file, err := os.OpenFile(path+chainCodeName+".abi.go", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	if _, err := file.Write([]byte(fileContent)); err != nil {
		panic(err)
	}
	file.Close()
	defer os.Remove(file.Name())
	//fmt.Println(file.Name())

	//
	stdout, err := os.OpenFile(chainCodeName+".abi.json", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer stdout.Close()

	//
	cmd := exec.Command("go", "run", file.Name())
	cmd.Stdout = stdout
	//
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	//
	if err := cmd.Run(); err != nil {
		fmt.Println(stdErr.String())
		fmt.Println(err.Error())
		return
	}
}

//const chainCodeCallPrefix = ".Start(new("
//const chainCodeCallEndfix = "))"
//
//func getUserChainCodeName(fileContent string) string {
//	index := strings.Index(fileContent, chainCodeCallPrefix)
//	if index == -1 {
//		return ""
//	}
//	index += len(chainCodeCallPrefix)
//	end := strings.Index(fileContent[index:], chainCodeCallEndfix)
//	if end == -1 {
//		return ""
//	}
//	return fileContent[index : end+index]
//}

func getChainCodeName(fileContent string) string {
	reg, err := regexp.Compile(`func(.*)+Invoke\(`)
	if err != nil {
		fmt.Println("errrrr")
		return ""
	}
	indexs := reg.FindIndex([]byte(fileContent))
	//fmt.Println(indexs)
	if len(indexs) != 2 {
		fmt.Println("FindIndex return empty")
		return ""
	}
	invokeStr := fileContent[indexs[0]:indexs[1]] //func (t *SimpleChaincode) Invoke(
	//fmt.Println(invokeStr)
	blankIndx := strings.Index(invokeStr[6:], " ") //t *SimpleChaincode) Invoke(
	blankIndx += 6                                 //add func (
	for '*' == invokeStr[blankIndx] || ' ' == invokeStr[blankIndx] {
		blankIndx += 1
	}

	return fileContent[indexs[0]+blankIndx : indexs[1]-9] //sub ) Invoke(
}

func getImportPath() string {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	//
	cmd := exec.Command("go", "list")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	//
	if err := cmd.Run(); err != nil {
		fmt.Println(stdErr.String())
		fmt.Println(err.Error())
		return ""
	}
	return stdOut.String()
}

func getPackage(inputFile string) string {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	//
	cmd := exec.Command("go", "list", "-f", "{{ .Name }}", inputFile)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	//
	if err := cmd.Run(); err != nil {
		fmt.Println(stdErr.String())
		fmt.Println(err.Error())
		return ""
	}
	return stdOut.String()
}

func getImports(inputFile string) string {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	//
	cmd := exec.Command("go", "list", "-f", "{{ join .Imports \"\\n\"}}", inputFile)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	//
	if err := cmd.Run(); err != nil {
		fmt.Println(stdErr.String())
		fmt.Println(err.Error())
		return ""
	}
	return stdOut.String()
}

func addImport(inputFile, fileStr string) (string, error) {
	importsList := getImports(inputFile)
	//fmt.Println(importsList)
	if len(importsList) == 0 {
		return "", fmt.Errorf("GetImports return empty")
	}

	imports := strings.Split(strings.Trim(importsList, "\n"), "\n")
	checkImports := []string{"encoding/json", "fmt", "reflect", "strings"}
	existImports := []string{}
	for i := range imports {
		if len(checkImports) == 0 {
			break
		}
		for j := range checkImports {
			if imports[i] == checkImports[j] {
				existImports = append(existImports, checkImports[j])
				if j < len(checkImports)-1 {
					checkImports = append(checkImports[:j], checkImports[j+1:]...)
				} else {
					checkImports = []string{}
				}
				break
			}
		}
	}
	for i := range checkImports {
		fileStr = strings.Replace(fileStr, "import (", "import (\n\""+checkImports[i]+"\"", 1)
	}
	for i := range existImports {
		fileStr = strings.Replace(fileStr, "_ \""+existImports[i]+"\"", "\""+existImports[i]+"\"", 1)
	}
	return fileStr, nil
}

func ReplaceContent(inputFile string) (string, string, bool, error) {
	//
	fileBuf, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return "", "", false, err
	}
	//
	chainCodeName := getChainCodeName(string(fileBuf))
	if chainCodeName == "" {
		return "", "", false, fmt.Errorf("GetChainCodeName failed, please check code, your chainCode file must exist Invoke")
	}
	//fmt.Println(chainCodeName)

	//
	if mainIndex := strings.Index(string(fileBuf), "func main"); mainIndex != -1 {
		//
		fileStr, err := addImport(inputFile, string(fileBuf))
		if err != nil {
			return "", "", false, err
		}
		fileStr = strings.Replace(fileStr, "func main", "func main_abi_bak", -1)
		fileStr += strings.Replace(appendMain, "SimpleChaincode", chainCodeName, 1)

		return fileStr, chainCodeName, false, nil
	} else {
		//
		importPath := getImportPath()
		//fmt.Println(importPath)
		if importPath == "" {
			return "", "", false, fmt.Errorf("getImportPath failed, please check code, is have syntax error ?")
		}
		if index := strings.Index(importPath, "\n"); index != -1 {
			importPath = importPath[:index]
		}
		//
		chainCodePackage := getPackage(inputFile)
		//fmt.Println(chainCodePackage)
		if chainCodePackage == "" {
			return "", "", false, fmt.Errorf("getPackage failed, please check code, your chainCode file no package")
		}
		if index := strings.Index(chainCodePackage, "\n"); index != -1 {
			chainCodePackage = chainCodePackage[:index]
		}
		fileStr := strings.Replace(appendPackage, "import (", "import (\n\""+importPath+"\"", 1)
		fileStr = fileStr + strings.Replace(appendMain, "SimpleChaincode", chainCodePackage+"."+chainCodeName, 1)
		return fileStr, chainCodePackage + "." + chainCodeName, true, nil
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Need one param.")
		fmt.Println("Usage: [your chainCode main file]")
		return
	}
	fileStr, chainCodeName, mkdir, err := ReplaceContent(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	GoRun(chainCodeName, fileStr, mkdir)
}
