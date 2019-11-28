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
 *  * @date 2018
 *
 */

package modules

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
)

type CertRawInfo struct {
	Issuer string
	Holder string
	Nonce  int // 不断加1的数，可以表示当前issuer发布的第几个证书。
	Cert   *x509.Certificate
}

type CertBytesInfo struct {
	Holder string
	Raw    []byte // 可以直接使用x509.ParseCertificate()接口获取证书信息
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

//func loadCert(path string) ([]byte, error) {
//	//加载PEM格式证书到字节数组
//	data, err := ioutil.ReadFile(path)
//	if err != nil {
//		return nil, err
//	}
//	return LoadCertBytes(data)
//}

func LoadCertBytes(original []byte) ([]byte, error) {
	certDERBlock, _ := pem.Decode(original)
	if certDERBlock == nil {
		return nil, fmt.Errorf("get none Cert info")
	}

	return certDERBlock.Bytes, nil
}
