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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoS256_Sign(t *testing.T) {
	crypto := &CryptoS256{}
	t.Log("Message:	abc100")
	msg := []byte("abc100")
	h, err := crypto.Hash(msg)
	assert.Nil(t, err)
	t.Logf("Sha3 256 Hash:	%x,len:%d", h, len(h))
	privKey, err := crypto.KeyGen()
	assert.Nil(t, err)
	t.Logf("PrivateKey:	%x,len:%d", privKey, len(privKey))
	pubKey, err := crypto.PrivateKeyToPubKey(privKey)
	assert.Nil(t, err)
	t.Logf("Pubkey:	%x,len:%d", pubKey, len(pubKey))
	sign, err := crypto.Sign(privKey, msg)
	assert.Nil(t, err)
	t.Logf("Signature:	%x,len:%d", sign, len(sign))

	pass, err := crypto.Verify(pubKey, sign, msg)
	assert.Nil(t, err)
	assert.True(t, pass)
}
