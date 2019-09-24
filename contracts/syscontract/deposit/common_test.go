///*
//	This file is part of go-palletone.
//	go-palletone is free software: you can redistribute it and/or modify
//	it under the terms of the GNU General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//	go-palletone is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.
//	You should have received a copy of the GNU General Public License
//	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
//*/
//
package deposit

import (
	"encoding/hex"
	"fmt"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	// 通过当前时间格式化
	now := time.Now().UTC()
	fmt.Println(now)
	l1 := now.Format("2006-01-02 15")
	fmt.Println(l1)
	l2 := now.Format("2006-01-02 15:04")
	fmt.Println(l2)
	l3 := now.Format("2006-01-02 15:04:05")
	fmt.Println(l3)
	//
	t1, _ := time.Parse("2006-01-02 15", l1)
	fmt.Println(t1)
	t2, _ := time.Parse("2006-01-02 15:04", l2)
	fmt.Println(t2)
	t3, _ := time.Parse("2006-01-02 15:04:05", l3)
	fmt.Println(t3)
	fmt.Println(t1.String())
}

func TestArray(t *testing.T) {
	var arr []string
	for i, v := range arr {
		fmt.Println(i, v)
	}
}

func TestLaa(t *testing.T) {
	//mainnetAddrAndPubKey["P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"] = "0326a0b144fd1df92f7d9e87d8d9929bc383059de4e0038b6d870f6f1d6ebb5219"
	b, _ := hex.DecodeString("0326a0b144fd1df92f7d9e87d8d9929bc383059de4e0038b6d870f6f1d6ebb5219")
	fmt.Println(crypto.PubkeyBytesToAddress(b).String())
	addr := "P1J7o5ri49ed1SNCw66A2UsmeZ1oRHiZCo7"
	encode := "0x03a3412f5ec867d575f01af8c60c73180ce6d00d0717f03c4c094a038acde0832b"
	fmt.Println(len(encode))
	byte, _ := hexutil.Decode(encode)
	fmt.Println(len(byte))
	if crypto.PubkeyBytesToAddress(byte).String() == addr {
		t.Log("success")
		return
	}
	t.Error("failed")
}

func TestL(t *testing.T) {
	addrAndPubKey := make(map[string]string)
	addrAndPubKey["P1HHQt1cMWrj3WBVmsTXYBsawwg2wrD6of6"] = "02d25b72c792d2aa83b45b07b4bfca3a6430dd3475bf855133742c51a40e81a2c8"
	addrAndPubKey["P1JxYp1dRpq2ZeYk58XEkSZJrptYEeuvZyq"] = "033f6c1422e15dd57ea960e957defad143b8be33e7c0ff055bdfdee00f6d094371"
	addrAndPubKey["P1KP5TZwTY8UowE7X3zSZ3gZDHqwCqcCThR"] = "02bbe16817006c18969ff559ea14522c67e4115d883fcd6dce9a86991ca097153a"
	addrAndPubKey["P1P3jZb43Y7stahiv8G3yvRUawGcWWvJBPt"] = "02c2f34e741a5840d49606f9533c1fb810ff4c5df7982773bbf32a058abfb4eaa4"
	addrAndPubKey["P1PuhsNTmpsSV36wyoEF49b5dhRdaTQYC2C"] = "03bcefd355332c533086f86570aba4ea470e36ada86788f55568f9e08e2f21c3bb"
	for a, p := range addrAndPubKey {
		b, _ := hex.DecodeString(p)
		if crypto.PubkeyBytesToAddress(b).String() != a {
			t.Errorf("error")
		}
	}
	//t.Log("success")

	mainnetAddrAndPubKey := make(map[string]string)
	mainnetAddrAndPubKey["P1JJoNc7EN5ekEduvvT7auJbRUsPeuSK2Rf"] = "03b3806a413f6d05df69eb5cd569f0b77593e4d095d006c7f87150f1e2f54f4f82"
	mainnetAddrAndPubKey["P1ByTXr3sm42zLpRJb9avyFtfpGFzaq2TYc"] = "02d1ef149d47eccddbb670735429ff702f6cb6abb505ad5072d0e36d9c38263ebf"
	mainnetAddrAndPubKey["P1JVoNN1VdGEYzw2ZuYJBJr78wNqfnhjQGH"] = "039149e808099dc3f59c3e8bb7bf3ba1e24639ef0a755921d6a78af72979762a7a"
	mainnetAddrAndPubKey["P1FaqdTMUiuaZqnsa2wqB2ayQ7MjDqxeNZr"] = "0254e55902db3eb82c4288c82c8eaa94f060d7dcdb482947d7a27e1a0b95b6a2ca"
	mainnetAddrAndPubKey["P13XFCo3n4UJjDLmGcwEVLqRGjU3awsF5kJ"] = "03d2869118ff41340f90863f7a4e6b2f024aae16ff4f1599659f0152efbdc91a55"
	mainnetAddrAndPubKey["P1AwgdqShDDWXq1Qm5m2MeaKmRFasFudo4Q"] = "03fc02969e762d2c78533ccd376ef9531800e143a7a15a18066df6140b81dc4be5"
	mainnetAddrAndPubKey["P1Hduv11WD89kjNzatrwVbpLN6CCD9d81nM"] = "02890fd43039e398074487b8393de0c6130d0bd8f7f68aa12c5465a658df5adff0"
	mainnetAddrAndPubKey["P1EdjUX1jVEAuQUG8Ev4MJ7mZhdbic4g5rB"] = "03573d69699190af6defb2c2d4212afb376c7942037bd3325ce88b52d8a57f9344"
	mainnetAddrAndPubKey["P15c2tpiRj7AZgQi3i8SHUZGwwDNF7zZSD8"] = "03e80f0b8b1a039b0e51ec7d8ae168dce01c8f3127b46ccaea883467528638544e"
	mainnetAddrAndPubKey["P1aNXZmxK1SdQSfB9y7jfpXc1E4SUPKvmv"] = "02047c2791f080264bbede8ddff970928182322f8721ecc8c61182e3c15e17fa30"

	mainnetAddrAndPubKey["P14zD1EoQFwC527vvL8iqM74FeFxGU7Javx"] = "03da1351099fa5e31abf885d04198dbb580e6181b46183556364b324eda11057bc"
	mainnetAddrAndPubKey["P1KJSodB2vzJ1A7jyqWVGb9NN7pCSt9ZZnN"] = "03cd92fd5c42506d798cb6dc13a88344195522ae6e2ddb2311395c00f466571714"
	mainnetAddrAndPubKey["P1HSMef9d1LTDwohiFCsdFzNH89ywF9QhfD"] = "030021457c0cce06e19c897fef0d5d2e79f2717562b387ea4efb5f3da2e7f95cb8"
	mainnetAddrAndPubKey["P1DAXndc7dDG9iEaECXpbxAbU4mBw8ZHsKh"] = "03b18f2099c593d575bd803e2d12e0212415bb864573bfa4e3e7515793a4643d95"
	mainnetAddrAndPubKey["P18s59sBXd6iJMkpgHBgCPuQ8voP4KiYKbE"] = "03c751ca2d4811bf6d63009b770042b1d635320267faddbf10416d2a73406bbecd"
	mainnetAddrAndPubKey["P1NPgdaVvvC9VXaZL5wc1AZJ4wrReKNu1Jd"] = "03581ae3a136ee8d622009492b1d64cfe98c05b4e3e768095cd4b79fe5eead1c61"
	mainnetAddrAndPubKey["P13Ly89hV5UA5pwXimPXXSh3WDD7iutMKgJ"] = "028768b0e7edfbc85c53581a9026d9de4dcc81318d9b33a737e1d542f9c53d9153"
	mainnetAddrAndPubKey["P1K3Tw5NRbU7Jd3PsoWw3wvPvvwPmPMTEpq"] = "02e2fc1e70ac4f4c097ac58192655164d17ed793da627456b2626678cd86669400"
	mainnetAddrAndPubKey["P1LuuPivkdvUF76wcTeunfDirkoFQQfPEEj"] = "02cff3f49842b66b4c6bab82153b8dc1457645f587f4895afedcf8c570c1f4f01b"
	mainnetAddrAndPubKey["P1NedpmrH4513E5jeM3soV4DyNxbgWGxZ4Y"] = "02e1f964897cbe7cd6a7c32b0d92f2e6fc9a4a12e6180fadd37e147d5e1536f1ed"

	mainnetAddrAndPubKey["P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"] = "0326a0b144fd1df92f7d9e87d8d9929bc383059de4e0038b6d870f6f1d6ebb5219"
	mainnetAddrAndPubKey["P17rPHgV9D61JKrbpcHtArktyuQrG7NM2xu"] = "0214e22f395f7af11ffe75cd7a992d15dc6478bb30abb0fd3dca997a9766e8ee5f"
	mainnetAddrAndPubKey["P1FzmurBnMv15SEkcKKiVoshMrP3ngULmi"] = "022babc996955a7552919caecbb022ccb2fbd84d8162b8b4c7ee823dfd52b074fa"
	mainnetAddrAndPubKey["P1EkrPBuLATsTyz4qda52HtcUcuCxdUeF9e"] = "03cdd9b7c57f19d4284bca24aa3725e7610f7cacf6b16624af62d7be2a9d300a96"
	mainnetAddrAndPubKey["P1HAFPiPTq7APjqzWC9qvw8jFFWWxPjdBvH"] = "022dafe2ce408c1d6389767da353fa4daeb7409627e956dccd2cff8379c7ddb7a0"

	for a, p := range mainnetAddrAndPubKey {
		b, _ := hex.DecodeString(p)
		if crypto.PubkeyBytesToAddress(b).String() != a {
			t.Errorf("error")
		}
	}
	t.Log("success")
}
