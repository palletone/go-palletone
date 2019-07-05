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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 *
 */

package crypto

import (
	"github.com/palletone/go-palletone/common/crypto/gmsm/sm2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCryptoGm_Sign(t *testing.T) {
	crypto:=&CryptoGm{}
	//h,err:=crypto.Hash([]byte("ABC"))
	//assert.Nil(t,err)
	//t.Logf("Hash:%x",h)

	msg := []byte("ABC")
	privKey,err:= crypto.KeyGen()
	assert.Nil(t,err)
	t.Logf("PrivateKey:%x",privKey)
	sign,err:= crypto.Sign(privKey,msg)
	assert.Nil(t,err)
	t.Logf("Signature:%x,len:%d",sign,len(sign))
	pubKey,err:=crypto.PrivateKeyToPubKey(privKey)
	assert.Nil(t,err)
	t.Logf("Pubkey:%x,len:%d",pubKey,len(pubKey))
	pass,err:= crypto.Verify(pubKey,sign,msg)
	assert.Nil(t,err)
	assert.True(t,pass)
}


func Test_key(t *testing.T) {
	crypto:=&CryptoGm{}
	//h,err:=crypto.Hash([]byte("ABC"))
	msg := []byte("ABC")
	//assert.Nil(t,err)
	//t.Logf("Hash:%x",h)
	privKey,err:= crypto.KeyGen()
	assert.Nil(t,err)
	t.Logf("PrivateKey:%x",privKey)
	sign,err:= crypto.Sign(privKey,msg)
	assert.Nil(t,err)
	t.Logf("Signature:%x,len:%d",sign,len(sign))
	pubKey,err:=crypto.PrivateKeyToPubKey(privKey)
	assert.Nil(t,err)
	t.Logf("Pubkey:%x,len:%d",pubKey,len(pubKey))

//标准包验证
    prikey,err := sm2ToECDSA(privKey)
    assert.Nil(t,err)
	t.Logf("Prikey:%x",prikey)


	ok := prikey.Verify(msg, sign) // 密钥验证
	if ok != true {
		t.Logf("Verify error\n")
	} else {
		t.Logf("Verify ok\n")
	}

	pubKey1, _ := prikey.Public().(*sm2.PublicKey)
	t.Logf("Pubkey:%x",pubKey1)
	ok = pubKey1.Verify(msg, sign) // 公钥验证
	if ok != true {
		t.Logf("Verify error\n")
	} else {
		t.Logf("Verify ok\n")
	}

}
