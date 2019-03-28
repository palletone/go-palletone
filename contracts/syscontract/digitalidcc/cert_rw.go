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
	"encoding/pem"
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
	"io/ioutil"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

type CertInfo struct {
	Issuer    string
	Nonce     int // 不断加1的数，可以表示当前issuer发布的第几个证书。
	cert      *x509.Certificate
	CertBytes []byte
}

// put cert state to write set
func setCert(certInfo *CertInfo, isServer bool, stub shim.ChaincodeStubInterface) error {
	var key string
	if isServer {
		key = dagConstants.CERT_SERVER_SYMBOL
	} else {
		key = dagConstants.CERT_MEMBER_SYMBOL
	}
	// put {issuer, certid} state
	key += certInfo.Issuer + dagConstants.CERT_SPLIT_CH + strconv.Itoa(certInfo.Nonce)
	if err := stub.PutState(key, certInfo.cert.SerialNumber.Bytes()); err != nil {
		return err
	}
	// put {certid, cert bytes} state
	key = dagConstants.CERT_BYTES_SYMBOL + certInfo.cert.SerialNumber.String()
	return stub.PutState(key, certInfo.CertBytes)
}

func getAddressCertIDs(addr string, stub shim.ChaincodeStubInterface) (serverCertIDs []string, memberCertIDs []string, err error) {
	// query server certificates
	serverCertIDs, err = queryCertsIDs(dagConstants.CERT_SERVER_SYMBOL, addr, stub)
	if err != nil {
		return nil, nil, err
	}
	// query memmber certificates
	memberCertIDs, err = queryCertsIDs(dagConstants.CERT_MEMBER_SYMBOL, addr, stub)
	if err != nil {
		return nil, nil, err
	}
	return serverCertIDs, memberCertIDs, nil
}

func queryCertsIDs(symbol string, issuer string, stub shim.ChaincodeStubInterface) (certids []string, err error) {
	prefixKey := symbol + issuer + dagConstants.CERT_SPLIT_CH
	KVs, err := stub.GetStateByPrefix(prefixKey)
	if err != nil {
		return nil, err
	}
	certids = []string{}
	for _, data := range KVs {
		var certID big.Int
		certID.SetBytes(data.Value)
		certids = append(certids, certID.String())
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

func queryNonce(isServer bool, issuer string, stub shim.ChaincodeStubInterface) (nonce int, err error) {
	var prefixKey string
	if isServer {
		prefixKey = dagConstants.CERT_SERVER_SYMBOL + issuer + dagConstants.CERT_SPLIT_CH
	} else {
		prefixKey = dagConstants.CERT_SERVER_SYMBOL + issuer + dagConstants.CERT_SPLIT_CH
	}
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
	if err != nil {
		return nil, err
	}
	return data, nil
}
