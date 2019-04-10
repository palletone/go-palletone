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
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
	"io/ioutil"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CertInfo struct {
	Issuer string
	Holder string
	Nonce  int // 不断加1的数，可以表示当前issuer发布的第几个证书。
	Cert   *x509.Certificate
}

type CertDBInfo struct {
	Holder string
	Raw    []byte
}

type CertHolderInfo struct {
	Holder   string
	IsServer bool // 是否是中间证书
	CertID   string
}

type CertState struct {
	CertID         string
	RecovationTime string
}

func (certHolderInfo *CertHolderInfo) Bytes() []byte {
	val, err := json.Marshal(certHolderInfo)
	if err != nil {
		return nil
	}
	return val
}

func (certHolderInfo *CertHolderInfo) SetBytes(data []byte) error {
	if err := json.Unmarshal(data, certHolderInfo); err != nil {
		return err
	}
	return nil
}

// put Cert state to write set
func setCert(certInfo *CertInfo, isServer bool, stub shim.ChaincodeStubInterface) error {
	// put {issuer, certid} state
	key := dagConstants.CERT_ISSUER_SYMBOL + certInfo.Issuer + dagConstants.CERT_SPLIT_CH + strconv.Itoa(certInfo.Nonce)
	certHolderInfo := CertHolderInfo{
		Holder:   certInfo.Holder,
		IsServer: isServer,
		CertID:   certInfo.Cert.SerialNumber.String(),
	}
	if err := stub.PutState(key, certHolderInfo.Bytes()); err != nil {
		return err
	}
	// put {holder, revocation} state
	if isServer {
		key = dagConstants.CERT_SERVER_SYMBOL
	} else {
		key = dagConstants.CERT_MEMBER_SYMBOL
	}
	key += certInfo.Holder + dagConstants.CERT_SPLIT_CH + certInfo.Cert.SerialNumber.String()
	recovationTime, _ := time.Time{}.MarshalBinary()
	if err := stub.PutState(key, recovationTime); err != nil {
		return err
	}
	// put {subject, certid} state
	key = dagConstants.CERT_SUBJECT_SYMBOL + certInfo.Cert.Subject.String()
	if err := stub.PutState(key, certInfo.Cert.SerialNumber.Bytes()); err != nil {
		return err
	}
	// put {certid, Cert bytes} state
	key = dagConstants.CERT_BYTES_SYMBOL + certInfo.Cert.SerialNumber.String()
	cerDBInfo := CertDBInfo{
		Holder: certInfo.Holder,
		Raw:    certInfo.Cert.Raw,
	}
	val, err := json.Marshal(cerDBInfo)
	if err != nil {
		return err
	}
	if err := stub.PutState(key, val); err != nil {
		return err
	}
	return nil
}

func getHolderCertIDs(addr string, stub shim.ChaincodeStubInterface) (serverCertStates []*CertState, memberCertStates []*CertState, err error) {
	// query server certificates
	serverCertStates, err = queryCertsIDs(dagConstants.CERT_SERVER_SYMBOL, addr, stub)
	if err != nil {
		return nil, nil, err
	}
	// query memmber certificates
	memberCertStates, err = queryCertsIDs(dagConstants.CERT_MEMBER_SYMBOL, addr, stub)
	if err != nil {
		return nil, nil, err
	}
	return serverCertStates, memberCertStates, nil
}

func getIssuerCertsInfo(issuer string, stub shim.ChaincodeStubInterface) (certHolderInfo []*CertHolderInfo, err error) {
	// query server certificates
	prefixKey := dagConstants.CERT_ISSUER_SYMBOL + issuer + dagConstants.CERT_SPLIT_CH
	KVs, err := stub.GetStateByPrefix(prefixKey)
	if err != nil {
		return nil, err
	}
	for _, data := range KVs {
		info := CertHolderInfo{}
		if err := info.SetBytes(data.Value); err != nil {
			return nil, err
		}
		certHolderInfo = append(certHolderInfo, &info)
	}
	return certHolderInfo, nil
}

func queryCertsIDs(symbol string, holder string, stub shim.ChaincodeStubInterface) (certstate []*CertState, err error) {
	prefixKey := symbol + holder + dagConstants.CERT_SPLIT_CH
	KVs, err := stub.GetStateByPrefix(prefixKey)
	if err != nil {
		return nil, err
	}
	certstate = []*CertState{}
	var revocationTime time.Time
	for _, data := range KVs {
		if err := revocationTime.UnmarshalBinary(data.Value); err != nil {
			return nil, err
		}
		certid, err := parseCertIDFrKey(data.Key)
		if err != nil {
			return nil, nil
		}
		state := CertState{
			CertID:         certid,
			RecovationTime: revocationTime.String(),
		}
		certstate = append(certstate, &state)
	}
	return certstate, nil
}

func loadCert(path string) ([]byte, error) {
	//加载PEM格式证书到字节数组
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadCertBytes(data)
}

func loadCertBytes(original []byte) ([]byte, error) {
	certDERBlock, _ := pem.Decode(original)
	if certDERBlock == nil {
		return nil, fmt.Errorf("get none Cert info")
	}

	return certDERBlock.Bytes, nil
}

func parseNondeFrKey(key string) (nonce int, err error) {
	ss := strings.Split(key, dagConstants.CERT_SPLIT_CH)
	if len(ss) != 2 {
		return -1, fmt.Errorf("get nonce from key error")
	}
	nonce, err = strconv.Atoi(ss[1])
	if err != nil {
		return -1, err
	}
	return nonce, nil
}

func parseCertIDFrKey(key string) (certId string, err error) {
	ss := strings.Split(key, dagConstants.CERT_SPLIT_CH)
	if len(ss) != 2 {
		return "", fmt.Errorf("get nonce from key error")
	}

	return ss[1], nil
}

func queryNonce(prefixSymbol string, issuer string, stub shim.ChaincodeStubInterface) (nonce int, err error) {
	prefixKey := prefixSymbol + issuer + dagConstants.CERT_SPLIT_CH
	KVs, err := stub.GetStateByPrefix(prefixKey)
	if err != nil {
		return -1, err
	}
	if len(KVs) <= 0 {
		return 0, nil
	}
	keys := []string{}
	for _, data := range KVs {
		keys = append(keys, data.Key)
	}
	// increasing order
	sort.Strings(keys)
	// the last one
	nonce, err = parseNondeFrKey(keys[len(keys)-1])
	if err != nil {
		return -1, err
	}
	return nonce, nil
}

func GetCertBytes(certid string, stub shim.ChaincodeStubInterface) (certBytes []byte, err error) {
	key := dagConstants.CERT_BYTES_SYMBOL + certid
	data, err := stub.GetState(key)
	if err != nil { // query none
		return nil, nil
	}
	certDBInfo := CertDBInfo{}
	if err := json.Unmarshal(data, &certDBInfo); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return certDBInfo.Raw, nil
}

func GetCertDBInfo(certid string, stub shim.ChaincodeStubInterface) (certDBInfo *CertDBInfo, err error) {
	key := dagConstants.CERT_BYTES_SYMBOL + certid
	data, err := stub.GetState(key)
	certDBInfo = &CertDBInfo{}
	if err := json.Unmarshal(data, certDBInfo); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return certDBInfo, nil
}

func setCRL(issuer string, crl *pkix.CertificateList, certHolderInfo []*CertHolderInfo, stub shim.ChaincodeStubInterface) error {
	var symbol string = ""
	for index, revokeCert := range crl.TBSCertList.RevokedCertificates {
		t, err := revokeCert.RevocationTime.MarshalBinary()
		if err != nil {
			return err
		}
		// update holder cert revocation
		if certHolderInfo[index].IsServer {
			symbol = dagConstants.CERT_SERVER_SYMBOL
		} else {
			symbol = dagConstants.CERT_MEMBER_SYMBOL
		}
		key := symbol + certHolderInfo[index].Holder + dagConstants.CERT_SPLIT_CH + certHolderInfo[index].CertID
		if err := stub.PutState(key, t); err != nil {
			return err
		}
		// update issuer crl bytes
		key = dagConstants.CRL_BYTES_SYMBOL + issuer
		if err := stub.PutState(key, crl.TBSCertList.Raw); err != nil {
			return err
		}
	}
	return nil
}

func getIssuerCRLBytes(issuer string, stub shim.ChaincodeStubInterface) ([]byte, error) {
	// query server certificates
	key := dagConstants.CRL_BYTES_SYMBOL + issuer
	data, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetIntermidateCertChains(cert *x509.Certificate, rootIssuer string, stub shim.ChaincodeStubInterface) (certChains []*x509.Certificate, err error) {
	subject := cert.Issuer.String()
	for {
		key := dagConstants.CERT_SUBJECT_SYMBOL + subject
		val, err := stub.GetState(key)
		if err != nil {
			return nil, err
		}
		// query chain done
		if val == nil {
			break
		}
		// parse certid
		certID := big.Int{}
		certID.SetBytes(val)
		// get cert bytes
		bytes, err := GetCertBytes(certID.String(), stub)
		if err != nil {
			return nil, err
		}
		// parse cert
		newCert, err := x509.ParseCertificate(bytes)
		if err != nil {
			return nil, err
		}
		certChains = append(certChains, newCert)
		subject = newCert.Issuer.String()
		if subject == rootIssuer {
			break
		}
	}

	return certChains, nil
}
