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
	"encoding/pem"
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"io/ioutil"
	"strconv"
	"strings"
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
	// check issuer
	issuer, err := validateIssuer(stub)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode validate issuer error: %s", err.Error())
		return shim.Error(reqStr)
	}
	// load cert file
	pemBytes, err := loadCert(certPath)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode error: load cert file[%s] error", certPath)
		return shim.Error(reqStr)
	}
	// dump cert from bytes to Certificate struct
	certID, err := checkCertSignature(pemBytes)
	if err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode check signature error:", err.Error())
		return shim.Error(reqStr)
	}
	// TODO check certificate signature

	// put cert state to write set
	if err := setCert(issuer, certID, pemBytes, isServer, stub); err != nil {
		reqStr := fmt.Sprintf("DigitalIdentityChainCode add simulator error:", err.Error())
		return shim.Error(reqStr)
	}

	rspStr := fmt.Sprintf("Add Cert success")
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
	return shim.Success(cerIDsJson) //test
}

func validateIssuer(stub shim.ChaincodeStubInterface) (issuer string, err error) {
	if issuer, err := stub.GetInvokeAddress(); err != nil {
		return "", err
	} else {
		// check with root ca holder
		rootCAHolder, err := stub.GetSystemConfig("RootCaHolder")
		if err != nil {
			return "", err
		}
		// check in server list
		if issuer != rootCAHolder {
			// TODO query from server certs db
			return "", fmt.Errorf("Has no validate intermidate certificate")
		}
		return issuer, nil
	}
}

// put cert state to write set
func setCert(issuer string, certID string, certBytes []byte, isServer bool, stub shim.ChaincodeStubInterface) error {
	var key string
	if isServer {
		key = CERT_SERVER_SYMBOL
	} else {
		key = CERT_MEMBER_SYMBOL
	}
	key += issuer + SPLIT_CH + certID

	return stub.PutState(string(key), certBytes)
}

func getAddressCertIDs(addr string, stub shim.ChaincodeStubInterface) (serverCertIDs []string, memberCertIDs []string, err error) {
	// query server certificates
	serverCertIDs, err = queryCertsIDs(CERT_SERVER_SYMBOL, addr, stub)
	if err != nil {
		return nil, nil, err
	}
	// query memmber certificates
	memberCertIDs, err = queryCertsIDs(CERT_MEMBER_SYMBOL, addr, stub)
	if err != nil {
		return nil, nil, err
	}
	return serverCertIDs, memberCertIDs, nil
}

func queryCertsIDs(symbol string, issuer string, stub shim.ChaincodeStubInterface) (certids []string, err error) {
	prefixKey := symbol + issuer + SPLIT_CH
	KVs, err := stub.GetStateByPrefix(prefixKey)
	if err != nil {
		return nil, err
	}
	certids = []string{}
	for _, data := range KVs {
		certID, err := parseSerialFrKey(data.Key)
		if err != nil {
			continue
		}
		certids = append(certids, certID)
	}
	return certids, nil
}

func loadCert(path string) ([]byte, error) {
	//加载PEM格式证书到字节数组
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	certDERBlock, _ := pem.Decode(data)
	if certDERBlock == nil {
		return nil, fmt.Errorf("get none cert infor")
	}

	return certDERBlock.Bytes, nil
}

func checkCertSignature(pemBytes []byte) (certID string, err error) {
	cert, err := x509.ParseCertificate(pemBytes)
	if err != nil {
		return "", err
	}
	log.Debugf("cert serial number", cert.SerialNumber.String())
	return cert.SerialNumber.String(), nil
}

func parseSerialFrKey(key string) (certID string, err error) {
	ss := strings.Split(key, SPLIT_CH)
	if len(ss) != 2 {
		return "", fmt.Errorf("get none serial number from key")
	}
	return ss[1], nil
}
