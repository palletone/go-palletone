// Copyright 2019 PalletOne
// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lint contains a linter for Go source code.
package lint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"

	"golang.org/x/tools/go/gcexportdata"
)

// A Linter lints Go source code.
type Linter struct {
}

// Problem represents a problem in some source code.
type Problem struct {
	Position   token.Position // position in source file
	Text       string         // the prose that describes the problem
	Link       string         // (optional) the link to the style guide for the problem
	Confidence float64        // a value in (0,1] estimating the confidence in this problem's correctness
	LineText   string         // the source line
	Category   string         // a short name for the general category of the problem

	// If the problem has a suggested fix (the minority case),
	// ReplacementLine is a full replacement for the relevant line of the source file.
	ReplacementLine string
}

// Lint lints src.
func (l *Linter) Lint(filename string, src []byte) ([]Problem, error) {
	return l.LintFiles(map[string][]byte{filename: src})
}

// LintFiles lints a set of files of a single package.
// The argument is a map of filename to source.
func (l *Linter) LintFiles(files map[string][]byte) ([]Problem, error) {
	pkg := &pkg{
		fset:  token.NewFileSet(),
		files: make(map[string]*file),
	}
	var pkgName string
	for filename, src := range files {
		if isGenerated(src) {
			continue // See issue #239
		}
		f, err := parser.ParseFile(pkg.fset, filename, src, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if pkgName == "" {
			pkgName = f.Name.Name
		} else if f.Name.Name != pkgName {
			return nil, fmt.Errorf("%s is in package %s, not %s", filename, f.Name.Name, pkgName)
		}
		pkg.files[filename] = &file{
			pkg:      pkg,
			f:        f,
			fset:     pkg.fset,
			src:      src,
			filename: filename,
		}
	}
	if len(pkg.files) == 0 {
		return nil, nil
	}
	return pkg.lint(), nil
}

var (
	genHdr = []byte("// Code generated ")
	genFtr = []byte(" DO NOT EDIT.")
)

// isGenerated reports whether the source file is generated code
// according the rules from https://golang.org/s/generatedcode.
func isGenerated(src []byte) bool {
	sc := bufio.NewScanner(bytes.NewReader(src))
	for sc.Scan() {
		b := sc.Bytes()
		if bytes.HasPrefix(b, genHdr) && bytes.HasSuffix(b, genFtr) && len(b) >= len(genHdr)+len(genFtr) {
			return true
		}
	}
	return false
}

// pkg represents a package being linted.
type pkg struct {
	fset  *token.FileSet
	files map[string]*file

	typesPkg  *types.Package
	typesInfo *types.Info

	// sortable is the set of types in the package that implement sort.Interface.
	//sortable map[string]bool
	// main is whether this is a "main" package.
	main bool

	problems []Problem
}

func (p *pkg) lint() []Problem {
	if err := p.typeCheck(); err != nil {
		fmt.Println(err.Error())
	}

	p.GenerateABI()

	p.main = p.isMain()

	return p.problems
}

// file represents a file being linted.
type file struct {
	pkg      *pkg
	f        *ast.File
	fset     *token.FileSet
	src      []byte
	filename string
}

//func (f *file) isTest() bool { return strings.HasSuffix(f.filename, "_test.go") }

var newImporter = func(fset *token.FileSet) types.ImporterFrom {
	return gcexportdata.NewImporter(fset, make(map[string]*types.Package))
}

func (p *pkg) typeCheck() error {
	config := &types.Config{
		// By setting a no-op error reporter, the type checker does as much work as possible.
		Error:    func(error) {},
		Importer: newImporter(p.fset),
	}
	info := &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	var anyFile *file
	astFiles := make([]*ast.File, 0)
	for _, f := range p.files {
		anyFile = f
		astFiles = append(astFiles, f.f)
	}
	pkg, err := config.Check(anyFile.f.Name.Name, p.fset, astFiles, info)
	// Remember the typechecking info, even if config.Check failed,
	// since we will get partial information.
	p.typesPkg = pkg
	p.typesInfo = info
	return err
}

func (p *pkg) typeOf(expr ast.Expr) types.Type {
	if p.typesInfo == nil {
		return nil
	}
	return p.typesInfo.TypeOf(expr)
}

//func (p *pkg) isNamedType(typ types.Type, importPath, name string) bool {
//	n, ok := typ.(*types.Named)
//	if !ok {
//		return false
//	}
//	tn := n.Obj()
//	return tn != nil && tn.Pkg() != nil && tn.Pkg().Path() == importPath && tn.Name() == name
//}

// scopeOf returns the tightest scope encompassing id.
//func (p *pkg) scopeOf(id *ast.Ident) *types.Scope {
//	var scope *types.Scope
//	if obj := p.typesInfo.ObjectOf(id); obj != nil {
//		scope = obj.Parent()
//	}
//	if scope == p.typesPkg.Scope() {
//		// We were given a top-level identifier.
//		// Use the file-level scope instead of the package-level scope.
//		pos := id.Pos()
//		for _, f := range p.files {
//			if f.f.Pos() <= pos && pos < f.f.End() {
//				scope = p.typesInfo.Scopes[f.f]
//				break
//			}
//		}
//	}
//	return scope
//}

type ABI_Param struct {
	Components string `json:"components,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}
type ABI_Function struct {
	Constant        bool        `json:"constant"`
	Inputs          []ABI_Param `json:"inputs"`
	Name            string      `json:"name"`
	Outputs         []ABI_Param `json:"outputs"`
	Payable         bool        `json:"payable"`
	StateMutability string      `json:"stateMutability"`
	Type            string      `json:"type"`
}

func (p *pkg) getChainCodeName() string {
	chainCodeName := ""
	for _, f := range p.files {
		f.walk(func(n ast.Node) bool {
			if chainCodeName != "" {
				//if n != nil {
				//	fmt.Println(n.Pos())
				//}
				return false
			}
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				return true
			}
			//
			if fn.Name.Name == "Invoke" {
				chainCodeName = receiverType(fn)
				return false
			}
			return true
		})
	}
	return chainCodeName
}

var (
	typePalletOne = map[string]string{
		"github.com/palletone/go-palletone/contracts/shim.ChaincodeStubInterface":   "",
		"github.com/palletone/go-palletone/core/vmContractPub/protos/peer.Response": "",
		"error": "",
		"github.com/palletone/go-palletone/common.Address":                                  "Address",
		"*github.com/palletone/go-palletone/common.Address":                                 "Address",
		"[]github.com/palletone/go-palletone/common.Address":                                "[]Address",
		"[]*github.com/palletone/go-palletone/common.Address":                               "[]Address",
		"github.com/palletone/go-palletone/common.Hash":                                     "Hash",
		"*github.com/palletone/go-palletone/common.Hash":                                    "Hash",
		"[]github.com/palletone/go-palletone/common.Hash":                                   "[]Hash",
		"[]*github.com/palletone/go-palletone/common.Hash":                                  "[]Hash",
		"github.com/palletone/go-palletone/vendor/github.com/shopspring/decimal.Decimal":    "Decimal",
		"*github.com/palletone/go-palletone/vendor/github.com/shopspring/decimal.Decimal":   "Decimal",
		"[]github.com/palletone/go-palletone/vendor/github.com/shopspring/decimal.Decimal":  "[]Decimal",
		"[]*github.com/palletone/go-palletone/vendor/github.com/shopspring/decimal.Decimal": "[]Decimal",
		"github.com/shopspring/decimal.Decimal":                                             "Decimal",
		"*github.com/shopspring/decimal.Decimal":                                            "Decimal",
		"[]github.com/shopspring/decimal.Decimal":                                           "[]Decimal",
		"[]*github.com/shopspring/decimal.Decimal":                                          "[]Decimal",
		"github.com/palletone/go-palletone/dag/modules.Asset":                               "Asset",
		"*github.com/palletone/go-palletone/dag/modules.Asset":                              "Asset",
		"[]github.com/palletone/go-palletone/dag/modules.Asset":                             "[]Asset",
		"[]*github.com/palletone/go-palletone/dag/modules.Asset":                            "[]Asset",
		"github.com/palletone/go-palletone/dag/modules.AssetId":                             "AssetId",
		"*github.com/palletone/go-palletone/dag/modules.AssetId":                            "AssetId",
		"[]github.com/palletone/go-palletone/dag/modules.AssetId":                           "[]AssetId",
		"[]*github.com/palletone/go-palletone/dag/modules.AssetId":                          "[]AssetId",
	}
)

func (p *pkg) getTypeName(paramType types.Type) (string, string) {
	typeName := paramType.String()
	//fmt.Println(typeName)
	typeNameShow, exist := typePalletOne[typeName]
	if exist {
		return typeNameShow, ""
	}

	switch paramType.Underlying().(type) {
	case *types.Basic:
		//fmt.Println("++++ ++++ Basic ", typeName)
		return typeName, ""
	case *types.Struct:
		testingT := paramType.Underlying().(*types.Struct)
		//fmt.Println("++++ ++++ Struct ", typeName)
		//fmt.Println(testingT.String())
		inputs := make([]ABI_Param, 0, testingT.NumFields())
		for i := 0; i < testingT.NumFields(); i++ {
			//fmt.Print(testingT.Field(i).Name(), "-", testingT.Field(i).Type().String(), "\n")
			typeName, detailName := p.getTypeName(testingT.Field(i).Type())
			inputs = append(inputs, ABI_Param{Name: testingT.Field(i).Name(), Type: typeName, Components: detailName})
		}
		abiJSON, err := json.Marshal(inputs)
		if err != nil {
			fmt.Println("types.Struct json.Marshal failed") //todo
			return "", ""
		}
		return "tuple", string(abiJSON) //strings.Replace(string(abiJSON), "\\", "", -1)
	case *types.Pointer:
		pointer := paramType.Underlying().(*types.Pointer)
		underlyType := pointer.Elem()
		testingT, ok := underlyType.Underlying().(*types.Struct)
		//fmt.Println("++++ ++++ Pointer", typeName)
		//fmt.Println(testingT.String())
		if ok {
			inputs := make([]ABI_Param, 0, testingT.NumFields())
			for i := 0; i < testingT.NumFields(); i++ {
				//fmt.Print(testingT.Field(i).Name(), "-", testingT.Field(i).Type().String(), "\n")
				typeName, detailName := p.getTypeName(testingT.Field(i).Type())
				inputs = append(inputs, ABI_Param{Name: testingT.Field(i).Name(), Type: typeName, Components: detailName})
			}
			abiJSON, err := json.Marshal(inputs)
			if err != nil {
				fmt.Println("types.Pointer json.Marshal failed") //todo
				return "", ""
			}
			return "tuple", string(abiJSON) //strings.Replace(string(abiJSON), "\\", "", -1)
		} else {
			fmt.Println("Not struct Pointer is not support") //todo
			return "", ""
		}
	case *types.Array:
		testingT := paramType.Underlying().(*types.Array)
		underlyType := testingT.Elem()
		//fmt.Println("++++ ++++ []Struct", typeName)
		//fmt.Println(underlyType.String())
		switch underlyType.Underlying().(type) {
		case *types.Basic:
			index := strings.Index(typeName, underlyType.String())
			if index < 0 {
				fmt.Println(typeName, "index ", index)
			}
			return underlyType.String() + typeName[:index], ""
		}
		_, detailName := p.getTypeName(underlyType)
		return "tuple[]", strings.Replace(detailName, "\\", "", -1)
	case *types.Slice:
		testingT := paramType.Underlying().(*types.Slice)
		underlyType := testingT.Elem()
		//fmt.Println("++++ ++++ []Slice", typeName)
		//fmt.Println(underlyType.String())
		switch underlyType.Underlying().(type) {
		case *types.Basic:
			return underlyType.String() + "[]", ""
		}
		_, detailName := p.getTypeName(underlyType)
		return "tuple[]", strings.Replace(detailName, "\\", "", -1)
	}
	//fmt.Println("Underlying() ==== ", underlyType.String(), paramType.String())
	return typeName, ""
}

func (p *pkg) GenerateABI() (string, error) {
	//
	allFunc := make([]ABI_Function, 0)

	//find Invoke
	chainCodeName := p.getChainCodeName()
	if chainCodeName == "" {
		return "", fmt.Errorf("Not find chainCode")
	}
	//fmt.Println("find: ", chainCodeName)

	//
	for _, f := range p.files {
		f.walk(func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				return true
			}
			//
			recv := receiverType(fn)
			if recv != chainCodeName {
				return true
			}

			//fmt.Println("==== ====.", recv, fn.Name.Name)
			funcName := fn.Name.Name

			if funcName == "Init" || funcName == "Invoke" {
				return true
			}
			oneFunc := ABI_Function{Name: strings.ToLower(funcName[0:1]) + funcName[1:], Type: "function",
				Inputs: make([]ABI_Param, 0), Outputs: make([]ABI_Param, 0)}

			if strings.HasPrefix(funcName, "Query") || strings.HasPrefix(funcName, "Find") ||
				strings.HasPrefix(funcName, "Get") {
				oneFunc.Constant = true
				oneFunc.Payable = false
				oneFunc.StateMutability = "view"
			} else {
				oneFunc.Constant = false //todo more, Payable & StateMutability
				if strings.HasSuffix(funcName, "Pay") {
					oneFunc.Payable = true
				}
				oneFunc.StateMutability = "nonpayable"
			}

			//paramsCount := 0
			for _, field := range fn.Type.Params.List {
				if len(field.Names) == 0 { //str1, str2 string
					fmt.Println("Invalid Input ==== ==== ==== ====")
				} else {
					typeName, detailName := p.getTypeName(p.typeOf(field.Type))
					if typeName == "" {
						continue
					}
					for _, name := range field.Names {
						//fmt.Print(name.Name, " ", typeName)
						oneInput := ABI_Param{Name: name.Name, Type: typeName, Components: detailName}
						oneFunc.Inputs = append(oneFunc.Inputs, oneInput)
					}

				}
				//fmt.Println("")
			}
			//fmt.Println("Params len ", paramsCount)
			//fmt.Println("params over")

			if nil != fn.Type.Results {
				//fmt.Println("Results len ", len(fn.Type.Results.List))
				for _, field := range fn.Type.Results.List {
					if len(field.Names) == 0 { //str1, str2 string
						resultName := ""
						typeName, detailName := p.getTypeName(p.typeOf(field.Type))
						if typeName != "" {
							//fmt.Print(resultName, " ", typeName)
							oneOutput := ABI_Param{Name: resultName, Type: typeName, Components: detailName}
							oneFunc.Outputs = append(oneFunc.Outputs, oneOutput)
						}
					} else {
						typeName, detailName := p.getTypeName(p.typeOf(field.Type))
						if typeName == "" {
							continue
						}
						for _, name := range field.Names {
							//fmt.Print(name.Name, " ", typeName)
							oneOutput := ABI_Param{Name: name.Name, Type: typeName, Components: detailName}
							oneFunc.Outputs = append(oneFunc.Outputs, oneOutput)
						}
					}
					//fmt.Println("")
				}
			} else {
				fmt.Println("Results len 0")
			}

			allFunc = append(allFunc, oneFunc)
			return true
		})
	}

	abiJSON, err := json.Marshal(allFunc)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	//fmt.Println(string(abiJSON))

	resultJSON := string(abiJSON)
	resultJSON = strings.Replace(resultJSON, "\\", "", -1)
	resultJSON = strings.Replace(resultJSON, "components\":\"[", "components\":[", -1)
	resultJSON = strings.Replace(resultJSON, "}]\"", "}]", -1)
	fmt.Println(resultJSON)
	WriteToFile(chainCodeName+".abi.json", resultJSON)

	return resultJSON, nil
}

func WriteToFile(fileName string, content string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("file create failed. err: " + err.Error())
	} else {
		// offset
		//os.Truncate(filename, 0) //clear
		n, _ := f.Seek(0, io.SeekEnd)
		_, err = f.WriteAt([]byte(content), n)
		//fmt.Println("write succeed!")
		defer f.Close()
	}
	return err
}

func (p *pkg) isMain() bool {
	for _, f := range p.files {
		if f.isMain() {
			return true
		}
	}
	return false
}

func (f *file) isMain() bool {
	if f.f.Name.Name == "main" {
		return true
	}
	return false
}

// exportedType reports whether typ is an exported type.
// It is imprecise, and will err on the side of returning true,
// such as for composite types.
//func exportedType(typ types.Type) bool {
//	switch T := typ.(type) {
//	case *types.Named:
//		// Builtin types have no package.
//		return T.Obj().Pkg() == nil || T.Obj().Exported()
//	case *types.Map:
//		return exportedType(T.Key()) && exportedType(T.Elem())
//	case interface {
//		Elem() types.Type
//	}: // array, slice, pointer, chan
//		return exportedType(T.Elem())
//	}
//	// Be conservative about other types, such as struct, interface, etc.
//	return true
//}

// receiverType returns the named type of the method receiver, sans "*",
// or "invalid-type" if fn.Recv is ill formed.
func receiverType(fn *ast.FuncDecl) string {
	switch e := fn.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		if id, ok := e.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	// The parser accepts much more than just the legal forms.
	return "invalid-type"
}

func (f *file) walk(fn func(ast.Node) bool) {
	ast.Walk(walker(fn), f.f)
}

// walker adapts a function to satisfy the ast.Visitor interface.
// The function return whether the walk should proceed into the node's children.
type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}
