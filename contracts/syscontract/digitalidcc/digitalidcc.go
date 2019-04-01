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
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
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
	// 获得持有者的所有证书ID
	case "getHolderCertIDs":
		return d.getAddressCertIDs(stub, args)
	// 获得证书颁发机构颁发的所有证书信息（持有者，是否是中间证书，证书ID）
	case "getIssuerCertsInfo":
		return d.getIssuerCertsInfo(stub, args)
	case "getCertFormateInfo":
		return d.getCertFormateInfo(stub, args)
	case "getCertBytes":
		return d.getCertBytes(stub, args)
	case "getCertHolder":
		return d.getCertHolder(stub, args)
	// 添加CRL证书
	case "addCRL":
		return d.addCRLCert(stub, args)
	// 获得crl的byte数据
	case "getCRL":
		return d.getIssuerCRL(stub, args)
	default:
		return shim.Error("Invoke error")
	}
	return shim.Error("Invoke error")
}

func (d *DigitalIdentityChainCode) addCert(stub shim.ChaincodeStubInterface, args []string, isServer bool) pb.Response {
	if len(args) != 2 {
		reqStr := fmt.Sprintf("Need two args: [holder address][Cert path]")
		return shim.Error(reqStr)
	}
	certHolder := args[0]
	certPath := args[1]
	// parse issuer
	issuer, err := stub.GetInvokeAddress()
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode parse issuer error:%s", err.Error())
		return shim.Error(reqStr)
	}
	// load Cert file
	pemBytes, err := loadCert(certPath)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode load [%s] error: %s", certPath, err.Error())
		return shim.Error(reqStr)
	}
	// parse Cert bytes to Certificate struct
	cert, err := x509.ParseCertificate(pemBytes)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode parse Cert error: %s", certPath)
		return shim.Error(reqStr)
	}
	// validate certificate
	if err := ValidateCert(issuer, cert, stub); err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode validate error: %s", err.Error())
		return shim.Error(reqStr)
	}
	// query nonce
	nonce, err := queryNonce(dagConstants.CERT_ISSUER_SYMBOL, issuer, stub)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode query nonce error: %s", err.Error())
		return shim.Error(reqStr)
	}
	certInfo := CertInfo{
		Issuer: issuer,
		Holder: certHolder,
		Cert:   cert,
		Nonce:  nonce + 1,
	}

	// put Cert state to write set
	if err := setCert(&certInfo, isServer, stub); err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode add simulator error:%s", err.Error())
		return shim.Error(reqStr)
	}

	rspStr := fmt.Sprintf("------ Add Cert success ------")
	return shim.Success([]byte(rspStr))
}

func (d *DigitalIdentityChainCode) addCRLCert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need 1 arg:[CRL Cert path]")
		return shim.Error(reqStr)
	}
	// parse issuer
	issuer, err := stub.GetInvokeAddress()
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode addCRLCert parse issuer error:%s", err.Error())
		return shim.Error(reqStr)
	}
	// load crl file
	crlPath := args[0]
	pemBytes, err := loadCert(crlPath)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode addCRLCert load [%s] error: %s", crlPath, err.Error())
		return shim.Error(reqStr)
	}
	// parse crl bytes to CertificateList struct
	crl, err := x509.ParseCRL(pemBytes)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode addCRLCert parse Cert error: %s", crlPath)
		return shim.Error(reqStr)
	}
	// validate crl issuer
	certHolderInfo, err := ValidateCRLIssuer(issuer, crl, stub)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode addCRLCert validate error: %s", err.Error())
		return shim.Error(reqStr)
	}

	// handle state
	if err := setCRL(issuer, crl, certHolderInfo, stub); err != nil {
		return shim.Error(fmt.Sprintf("DigitalIdentityChainCode addCRLCert save state error: %s", err.Error()))
	}
	return shim.Success([]byte("---- Add CRL Success --- "))
}

func (d *DigitalIdentityChainCode) getAddressCertIDs(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [holder address]")
		return shim.Error(reqStr)
	}
	serverCertIDs, memberCertIDs, err := getHolderCertIDs(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("get address Cert ids error:%s", err.Error())
		return shim.Error(reqStr)
	}

	certIDs := map[string][]*CertState{
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

func (d *DigitalIdentityChainCode) getIssuerCertsInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [issuer address]")
		return shim.Error(reqStr)
	}
	issuerCertInfo, err := getIssuerCertsInfo(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("get issuer certs info error:%s", err.Error())
		return shim.Error(reqStr)
	}

	//return json
	cerIDsJson, err := json.Marshal(issuerCertInfo)
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
		reqStr := fmt.Sprintf("Get Cert byts error: %s", err.Error())
		return shim.Error(reqStr)
	}
	cert, err := x509.ParseCertificate(data)
	certInfoJson, err := json.Marshal(cert)
	if err != nil {
		reqStr := fmt.Sprintf("Get Cert format info error: %s", err.Error())
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
		reqStr := fmt.Sprintf("Get Cert byts error: %s", err.Error())
		return shim.Error(reqStr)
	}

	certInfoMap := map[string]interface{}{
		"CertID":    args[0],
		"CertBytes": data,
	}
	certInfoJson, err := json.Marshal(certInfoMap)
	if err != nil {
		reqStr := fmt.Sprintf("Get Cert byts error: %s", err.Error())
		return shim.Error(reqStr)
	}
	return shim.Success(certInfoJson)
}

func (d *DigitalIdentityChainCode) getCertHolder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [certificate serial number]")
		return shim.Error(reqStr)
	}
	data, err := GetCertDBInfo(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("Get Cert holder error: %s", err.Error())
		return shim.Error(reqStr)
	}
	certInfoJson, err := json.Marshal(data.Holder)
	if err != nil {
		reqStr := fmt.Sprintf("Get Cert holder error: %s", err.Error())
		return shim.Error(reqStr)
	}
	return shim.Success(certInfoJson)
}

func (d *DigitalIdentityChainCode) getIssuerCRL(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		reqStr := fmt.Sprintf("Need one args: [issuer address]")
		return shim.Error(reqStr)
	}
	crlInfo, err := getIssuerCRLBytes(args[0], stub)
	if err != nil {
		reqStr := fmt.Sprintf("get issuer crl info error:%s", err.Error())
		return shim.Error(reqStr)
	}

	//return json
	crlBytesJson, err := json.Marshal(crlInfo)
	if err != nil {
		return shim.Success([]byte(err.Error()))
	}
	return shim.Success(crlBytesJson)
}
