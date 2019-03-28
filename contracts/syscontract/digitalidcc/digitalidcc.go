/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package digitalidcc

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
)

type DigitalIdentityChainCode struct {
}

func (d *DigitalIdentityChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//log.Info("***System digital contract init success ***")
	return shim.Success([]byte("ok"))
}

func (d *DigitalIdentityChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	// 添加中间Server证书
	case "addServerCert":
		return d.addCert(stub, args, true)
	// 添加用户证书
	case "addMemberCert":
		return d.addCert(stub, args, false)
	// 添加CRL证书
	case "addCRLCert":
		return d.addCRLCert(stub, args)
	case "getAddressCertIDs":
		return d.getAddressCertIDs(stub, args)
	case "getCertFormateInfo":
		return d.getCertFormateInfo(stub, args)
	case "getCertBytes":
		return d.getCertBytes(stub, args)
	default:
		return shim.Error("Invoke error")
	}
	return shim.Error("Invoke error")
}

func (d *DigitalIdentityChainCode) addCert(stub shim.ChaincodeStubInterface, args []string, isServer bool) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [cert path]")
		return shim.Error(reqStr)
	}
	certPath := args[0]
	var certInfo CertInfo
	// parse issuer
	issuer, err := stub.GetInvokeAddress()
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode parse issuer error:%s", err.Error())
		return shim.Error(reqStr)
	}
	// load cert file
	pemBytes, err := loadCert(certPath)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode load [%s] error: %s", certPath, err.Error())
		return shim.Error(reqStr)
	}
	// parse cert bytes to Certificate struct
	cert, err := x509.ParseCertificate(pemBytes)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode parse cert error: %s", certPath)
		return shim.Error(reqStr)
	}
	// validate certificate
	log.Debugf("cert serial number", cert.SerialNumber.String())
	// check issuer
	if err := ValidateCert(issuer, cert, stub); err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode validate error: %s", err.Error())
		return shim.Error(reqStr)
	}
	// query nonce
	nonce, err := queryNonce(isServer, issuer, stub)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode query nonce error: %s", err.Error())
		return shim.Error(reqStr)
	}
	certInfo.Issuer = issuer
	certInfo.CertBytes = pemBytes
	certInfo.cert = cert
	certInfo.Nonce = nonce + 1
	// put cert state to write set
	if err := setCert(&certInfo, isServer, stub); err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode add simulator error:%s", err.Error())
		return shim.Error(reqStr)
	}

	rspStr := fmt.Sprintf("------ Add Cert success ------")
	return shim.Success([]byte(rspStr))
}

func (d *DigitalIdentityChainCode) addCRLCert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	a, _ := strconv.Atoi(args[0])
	b, _ := strconv.Atoi(args[1])
	rspStr := fmt.Sprintf("Value:%d", a+b)
	return shim.Success([]byte(rspStr))
}

func (d *DigitalIdentityChainCode) getAddressCertIDs(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [issuer address]")
		return shim.Error(reqStr)
	}
	serverCertIDs, memberCertIDs, err := getAddressCertIDs(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("get address cert ids error:%s", err.Error())
		return shim.Error(reqStr)
	}

	certIDs := map[string][]string{
		"InterCertIDs":  serverCertIDs,
		"MemberCertIDs": memberCertIDs,
	}

	//return json
	cerIDsJson, err := json.Marshal(certIDs)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(cerIDsJson)
}

func (d *DigitalIdentityChainCode) getCertFormateInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [certificate serial number]")
		return shim.Error(reqStr)
	}
	data, err := GetCertBytes(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("Get cert byts error: %s", err.Error())
		return shim.Error(reqStr)
	}
	cert, err := x509.ParseCertificate(data)
	certInfoJson, err := json.Marshal(cert)
	if err != nil {
		reqStr := fmt.Sprintf("Get cert format info error: %s", err.Error())
		return shim.Error(reqStr)
	}
	return shim.Success(certInfoJson)
}

func (d *DigitalIdentityChainCode) getCertBytes(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [certificate serial number]")
		return shim.Error(reqStr)
	}
	data, err := GetCertBytes(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("Get cert byts error: %s", err.Error())
		return shim.Error(reqStr)
	}

	certInfoMap := map[string]interface{}{
		"CertID":    args[0],
		"CertBytes": data,
	}
	certInfoJson, err := json.Marshal(certInfoMap)
	if err != nil {
		reqStr := fmt.Sprintf("Get cert byts error: %s", err.Error())
		return shim.Error(reqStr)
	}
	return shim.Success(certInfoJson)
}
