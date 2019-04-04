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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package core

const (
	DefaultAlias                     = "PTN"
	DefaultTokenAmount               = "100000000000000000"
	DefaultTokenDecimal              = 8
	DefaultChainID                   = 1
	DefaultDepositRate               = "0.02"
	DefaultDepositPeriod             = "0"
	DefaultDepositAmountForMediator  = "200000000000"
	DefaultDepositAmountForJury      = "100000000000"
	DefaultDepositAmountForDeveloper = "80000000000"

	DefaultFoundationAddress = "P1LA8TkEWxU6FcMzkyeSbf9b9FwZwxrYRuF"
	DefaultTokenHolder       = "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"

	DefaultMediator = "P1Da7wwuvXgwqFm17GsLs4Cp4SLiPXZ6paF"
	DefaultNodeInfo = "pnode://4bdc1c533f6e3700a0a6cc346bf2364eace58a10d8a782762c8d2b27cf4d96c25827c82a15" +
		"684d348e88722b259f31abcccd4d0eaae0f52eeb85e1eb5342b862@127.0.0.1:30303"
	DefaultInitPubKey = "XmMwxWh6J71HtzndJy37gNDE9zcZqnHANkbxLHfBWYQwfBJyLeWq17kNRRR4bavoe3Brf5oGpWCYBy" +
		"MpbsWk45ymz4kmjU2AZo8Rm3mJ3MQHpdAgTo2nzWmqU3vCTW6qCfviPD1MKu3FJtmaWiLzdavLx831eCBXA1CdaiXAeU5MPcQ"

	DefaultJuryAddr = "P16bXzewsexHwhGYdt1c1qbzjBirCqDg8mN"

	DefaultMaxMediatorCount    = 255
	DefaultMediatorCount       = 5 //21
	DefaultMinMediatorCount    = 5 //21
	DefaultMinMediatorInterval = 1

	DefaultText = "姓名 丨 坐标 丨 简介   \r\n" +
		"孟岩丨北京丨通证派倡导者、CSDN副总裁、柏链道捷CEO.\r\n" +
		"刘百祥丨上海丨 GoC-lab发起人兼技术社群负责人,复旦大学计算机博士.\r\n" +
		"陈澄丨上海丨引力区开发者社区总理事,EOS超级节点负责人.\r\n" +
		"孙红景丨北京丨CTO、13年IT开发和管理经验.\r\n" +
		"kobegpfan丨北京丨世界500强企业技术总监.\r\n" +
		"余奎丨上海丨加密经济学研究员、产品研发经理.\r\n" +
		"Shangsong丨北京丨Fabric、 多链、 分片 、跨链技术.\r\n" +
		"郑广军丨上海丨区块链java应用开发.\r\n" +
		"钮祜禄虫丨北京丨大数据架构、Dapp开发.\r\n" +
		"彭敏丨四川丨计算机网络和系统集成十余年有经验.\r\n"

	DefaultRootCABytes = "MIID7TCCAtWgAwIBAgIUfEMGBtKUSmCHkv87i/HbY+dFhaQwDQYJKoZIhvcNAQEL" +
		"BQAwgYUxCzAJBgNVBAYTAkNOMRAwDgYDVQQIDAdCZWlqaW5nMRAwDgYDVQQHDAdC" +
		"ZWlqaW5nMQ0wCwYDVQQKDARITFlXMRIwEAYDVQQLDAlQYWxsZXRPbmUxDzANBgNV" +
		"BAMMBnlhbmd5dTEeMBwGCSqGSIb3DQEJARYPeWFuZ3l1QGhseXcuY29tMB4XDTE5" +
		"MDMyOTA3MzgwMloXDTI5MDMyNjA3MzgwMlowgYUxCzAJBgNVBAYTAkNOMRAwDgYD" +
		"VQQIDAdCZWlqaW5nMRAwDgYDVQQHDAdCZWlqaW5nMQ0wCwYDVQQKDARITFlXMRIw" +
		"EAYDVQQLDAlQYWxsZXRPbmUxDzANBgNVBAMMBnlhbmd5dTEeMBwGCSqGSIb3DQEJ" +
		"ARYPeWFuZ3l1QGhseXcuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC" +
		"AQEA381wb09v/UekCg4Y54mjioNnJ0FcHAEMlt9XSv5Kqr+5DzZcKlWI6+85bzYu" +
		"hzko5lFm/onb1zH1bVw0M70Us6/H//f5f0px4CDgJ5mApf8rTbMQqf8Xv7QUEuJX" +
		"x2rizHorxDrwWrZIIF4VhkQDIgRHVE1UUnHY8NM6xBaKAZEv6UrAE5+ImblA55Bk" +
		"2OSnKyYoSLNejs4uFtUBFwfgnIZ08fg6sHPFz0waD3GJtJKv8UISRvdts0JrTbdR" +
		"hRviJhXL+E0nQKwgJpmb/NO5uzAk3jv1Ezt1ZZQ4ey9bAGlU8HV0Lu1xAqYjWl3X" +
		"sG5WT56Wx7DrJ7WYE0c8Zl4jOwIDAQABo1MwUTAdBgNVHQ4EFgQUcFrfnln9ROaE" +
		"thXIpuHadFqDO6AwHwYDVR0jBBgwFoAUcFrfnln9ROaEthXIpuHadFqDO6AwDwYD" +
		"VR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAJQgFAZ2aXCYWXJ//VW0R" +
		"Uie5XCoglXy38ReaAboWFdHCM9YUV1T/42ptOypHqmoq/3fRSRlM2sXCK3/m2onU" +
		"xRG5jp8FV+kNSenZkg3aoVQgiudvEkydl6eBZbeOpVS64x0FsZ+NRj0PXEPL9cn7" +
		"Bb0Ne4caASTRyk07yZeA3QlOj6MCH4/P/4E6O5Ohac10/ZecLbSpImkaA7MFaPRu" +
		"G1b2ghWzkesKWiAWkWxBNJ5DMG01A0qfhUBSSUAqTUGvwYdzJ5xSN+yssjP1XnA3" +
		"vftAiyOpxD+V/8N3jO6ObMCWqbxYCsOjq1TFgNlDcWR/RQXYrblnCHHfyoJhPnOb"

	/* percentage fields are fixed point with a denominator of 10,000 */
	PalletOne100Percent            = 10000
	PalletOne1Percent              = PalletOne100Percent / 100
	PalletOneIrreversibleThreshold = 70 * PalletOne1Percent

	DefaultMediatorInterval    = 3       //5 /* seconds */
	DefaultMaintenanceInterval = 60 * 10 //60 * 60 * 24 // seconds, aka: 1 day

	DefaultMediatorCreateFee        = 5000
	DefaultAccountUpdateFee         = 20
	DefaultTransferPtnBaseFee       = 20
	DefaultTransferPtnPricePerKByte = 20
)
