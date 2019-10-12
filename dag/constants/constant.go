/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package constants

var (
	GENESIS_UNIT    string
	VERSION         string
	ALT             string
	COUNT_WITNESSES int
	// anti-spam limits
	MAX_AUTHORS_PER_UNIT                   int = 16
	MAX_PARENTS_PER_UNIT                   int = 16
	MAX_MESSAGES_PER_UNIT                  int = 128
	MAX_SPEND_PROOFS_PER_MESSAGE           int = 128
	MAX_INPUTS_PER_PAYMENT_MESSAGE         int = 128
	MAX_OUTPUTS_PER_PAYMENT_MESSAGE        int = 128
	MAX_CHOICES_PER_POLL                   int = 128
	MAX_DENOMINATIONS_PER_ASSET_DEFINITION int = 64
	MAX_ATTESTORS_PER_ASSET                int = 64
	MAX_DATA_FEED_NAME_LENGTH              int = 64
	MAX_DATA_FEED_VALUE_LENGTH             int = 64
	MAX_AUTHENTIFIER_LENGTH                int = 4096
	MAX_CAP                                int = 9e15
	MAX_COMPLEXITY                         int = 100

	MAX_PROFILE_FIELD_LENGTH int = 50
	MAX_PROFILE_VALUE_LENGTH int = 100

	TEXTCOIN_CLAIM_FEE                int = 548
	TEXTCOIN_ASSET_CLAIM_FEE          int = 750
	TEXTCOIN_ASSET_CLAIM_HEADER_FEE   int = 391
	TEXTCOIN_ASSET_CLAIM_MESSAGE_FEE  int = 209
	TEXTCOIN_ASSET_CLAIM_BASE_MSG_FEE int = 158
	VOTED_MEDIATORS                       = "VotedMediators"

	PledgeListLastDate = "PledgeListLastDate"
	PledgeList         = "PledgeList-"
	BlacklistAddress   = "BlacklistAddress"

	OldTestNetGenesisMediatorAndPubKey = make(map[string]string) // 测试网上genesis中定义的mediator
	OldMainNetGenesisMediatorAndPubKey = make(map[string]string) // 主网上genesis中定义的mediator
	OldMainNetMediatorAndPubKey        = make(map[string]string) // 1.0.3 版本之前主网上新申请的mediator

	TestNetGenesisHash = "0x6365f3bc9c197b8679821b998da5ee8f88b3db67fdb023250db3d1c2ae0ab1c6"
	MainNetGenesisHash = "0xfda8aea1b4b2920b1f4038fb10edb8fe510669ec1f574f5a13b20cad0f2294df"
)

func init() {
	VERSION = "1.0"
	ALT = "1"
	//	log.Println("start constant init...")
	if VERSION == "1.0" && ALT == "1" {
		GENESIS_UNIT = "TvqutGPz3T4Cs6oiChxFlclY92M2MvCvfXR5/FETato="

	} else {
		GENESIS_UNIT = "oj8yEksX9Ubq7lLc+p6F2uyHUuynugeVq4+ikT67X6E="

	}

	//  testnet  5个节点
	OldTestNetGenesisMediatorAndPubKey["P1HHQt1cMWrj3WBVmsTXYBsawwg2wrD6of6"] = "02d25b72c792d2aa83b45b07b4bfca3a6430dd3475bf855133742c51a40e81a2c8"
	OldTestNetGenesisMediatorAndPubKey["P1JxYp1dRpq2ZeYk58XEkSZJrptYEeuvZyq"] = "033f6c1422e15dd57ea960e957defad143b8be33e7c0ff055bdfdee00f6d094371"
	OldTestNetGenesisMediatorAndPubKey["P1KP5TZwTY8UowE7X3zSZ3gZDHqwCqcCThR"] = "02bbe16817006c18969ff559ea14522c67e4115d883fcd6dce9a86991ca097153a"
	OldTestNetGenesisMediatorAndPubKey["P1P3jZb43Y7stahiv8G3yvRUawGcWWvJBPt"] = "02c2f34e741a5840d49606f9533c1fb810ff4c5df7982773bbf32a058abfb4eaa4"
	OldTestNetGenesisMediatorAndPubKey["P1PuhsNTmpsSV36wyoEF49b5dhRdaTQYC2C"] = "03bcefd355332c533086f86570aba4ea470e36ada86788f55568f9e08e2f21c3bb"
	//  testnet

	//  mainnet 25个节点
	OldMainNetGenesisMediatorAndPubKey["P1JJoNc7EN5ekEduvvT7auJbRUsPeuSK2Rf"] = "03b3806a413f6d05df69eb5cd569f0b77593e4d095d006c7f87150f1e2f54f4f82"
	OldMainNetGenesisMediatorAndPubKey["P1AwgdqShDDWXq1Qm5m2MeaKmRFasFudo4Q"] = "03fc02969e762d2c78533ccd376ef9531800e143a7a15a18066df6140b81dc4be5"
	OldMainNetGenesisMediatorAndPubKey["P1Hduv11WD89kjNzatrwVbpLN6CCD9d81nM"] = "02890fd43039e398074487b8393de0c6130d0bd8f7f68aa12c5465a658df5adff0"
	OldMainNetGenesisMediatorAndPubKey["P14zD1EoQFwC527vvL8iqM74FeFxGU7Javx"] = "03da1351099fa5e31abf885d04198dbb580e6181b46183556364b324eda11057bc"
	OldMainNetGenesisMediatorAndPubKey["P1HAFPiPTq7APjqzWC9qvw8jFFWWxPjdBvH"] = "022dafe2ce408c1d6389767da353fa4daeb7409627e956dccd2cff8379c7ddb7a0"

	OldMainNetMediatorAndPubKey["P1ByTXr3sm42zLpRJb9avyFtfpGFzaq2TYc"] = "02d1ef149d47eccddbb670735429ff702f6cb6abb505ad5072d0e36d9c38263ebf"
	OldMainNetMediatorAndPubKey["P1JVoNN1VdGEYzw2ZuYJBJr78wNqfnhjQGH"] = "039149e808099dc3f59c3e8bb7bf3ba1e24639ef0a755921d6a78af72979762a7a"
	OldMainNetMediatorAndPubKey["P1FaqdTMUiuaZqnsa2wqB2ayQ7MjDqxeNZr"] = "0254e55902db3eb82c4288c82c8eaa94f060d7dcdb482947d7a27e1a0b95b6a2ca"
	OldMainNetMediatorAndPubKey["P13XFCo3n4UJjDLmGcwEVLqRGjU3awsF5kJ"] = "03d2869118ff41340f90863f7a4e6b2f024aae16ff4f1599659f0152efbdc91a55"
	OldMainNetMediatorAndPubKey["P1EdjUX1jVEAuQUG8Ev4MJ7mZhdbic4g5rB"] = "03573d69699190af6defb2c2d4212afb376c7942037bd3325ce88b52d8a57f9344"
	OldMainNetMediatorAndPubKey["P15c2tpiRj7AZgQi3i8SHUZGwwDNF7zZSD8"] = "03e80f0b8b1a039b0e51ec7d8ae168dce01c8f3127b46ccaea883467528638544e"
	OldMainNetMediatorAndPubKey["P1aNXZmxK1SdQSfB9y7jfpXc1E4SUPKvmv"] = "02047c2791f080264bbede8ddff970928182322f8721ecc8c61182e3c15e17fa30"
	OldMainNetMediatorAndPubKey["P1KJSodB2vzJ1A7jyqWVGb9NN7pCSt9ZZnN"] = "03cd92fd5c42506d798cb6dc13a88344195522ae6e2ddb2311395c00f466571714"
	OldMainNetMediatorAndPubKey["P1HSMef9d1LTDwohiFCsdFzNH89ywF9QhfD"] = "030021457c0cce06e19c897fef0d5d2e79f2717562b387ea4efb5f3da2e7f95cb8"
	OldMainNetMediatorAndPubKey["P1DAXndc7dDG9iEaECXpbxAbU4mBw8ZHsKh"] = "03b18f2099c593d575bd803e2d12e0212415bb864573bfa4e3e7515793a4643d95"
	OldMainNetMediatorAndPubKey["P18s59sBXd6iJMkpgHBgCPuQ8voP4KiYKbE"] = "03c751ca2d4811bf6d63009b770042b1d635320267faddbf10416d2a73406bbecd"
	OldMainNetMediatorAndPubKey["P1NPgdaVvvC9VXaZL5wc1AZJ4wrReKNu1Jd"] = "03581ae3a136ee8d622009492b1d64cfe98c05b4e3e768095cd4b79fe5eead1c61"
	OldMainNetMediatorAndPubKey["P13Ly89hV5UA5pwXimPXXSh3WDD7iutMKgJ"] = "028768b0e7edfbc85c53581a9026d9de4dcc81318d9b33a737e1d542f9c53d9153"
	OldMainNetMediatorAndPubKey["P1K3Tw5NRbU7Jd3PsoWw3wvPvvwPmPMTEpq"] = "02e2fc1e70ac4f4c097ac58192655164d17ed793da627456b2626678cd86669400"
	OldMainNetMediatorAndPubKey["P1LuuPivkdvUF76wcTeunfDirkoFQQfPEEj"] = "02cff3f49842b66b4c6bab82153b8dc1457645f587f4895afedcf8c570c1f4f01b"
	OldMainNetMediatorAndPubKey["P1NedpmrH4513E5jeM3soV4DyNxbgWGxZ4Y"] = "02e1f964897cbe7cd6a7c32b0d92f2e6fc9a4a12e6180fadd37e147d5e1536f1ed"
	OldMainNetMediatorAndPubKey["P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"] = "0326a0b144fd1df92f7d9e87d8d9929bc383059de4e0038b6d870f6f1d6ebb5219"
	OldMainNetMediatorAndPubKey["P17rPHgV9D61JKrbpcHtArktyuQrG7NM2xu"] = "0214e22f395f7af11ffe75cd7a992d15dc6478bb30abb0fd3dca997a9766e8ee5f"
	OldMainNetMediatorAndPubKey["P1FzmurBnMv15SEkcKKiVoshMrP3ngULmi"] = "022babc996955a7552919caecbb022ccb2fbd84d8162b8b4c7ee823dfd52b074fa"
	OldMainNetMediatorAndPubKey["P1EkrPBuLATsTyz4qda52HtcUcuCxdUeF9e"] = "03cdd9b7c57f19d4284bca24aa3725e7610f7cacf6b16624af62d7be2a9d300a96"
	OldMainNetMediatorAndPubKey["P1HD5bnL5JSHGghYcfkY19A6wgJR49kfszT"] = "02187443be44501236925c2ee322bfc9edfb0b359b9d5f139d33cae64ea4c12bf4"
	//  mainnet
}
