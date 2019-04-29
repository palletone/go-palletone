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
	DefaultAlias       = "PTN"
	DefaultTokenAmount = "100000000000000000"
	//DefaultTokenDecimal              = 8
	DefaultChainID                   = 1
	DefaultDepositRate               = "0.02"
	DefaultTxCoinYearRate            = "0.01"
	DefaultGenerateUnitReward        = "100000000"
	DefaultDepositPeriod             = "0"
	DefaultDepositAmountForMediator  = "200000000000"
	DefaultDepositAmountForJury      = "100000000000"
	DefaultDepositAmountForDeveloper = "80000000000"
	DefaultFoundationAddress         = "P1LA8TkEWxU6FcMzkyeSbf9b9FwZwxrYRuF"

	DefaultUccMemory     = "1073741824" //物理内存  1073741824  1G
	DefaultUccMemorySwap = "1073741824" //内存交换区，不设置默认为memory的两倍
	DefaultUccCpuShares  = "1024"       //CPU占用率，相对的  CPU 利用率权重，默认为 1024
	DefaultCpuPeriod     = "50000"      // 限制CPU --cpu-period=50000 --cpu-quota=25000
	DefaultUccCpuQuota   = "25000"      //限制CPU 周期设为 50000，将容器在每个周期内的 CPU 配额设置为 25000，表示该容器每 50ms 可以得到 50% 的 CPU 运行时间
	DefaultUccCpuSetCpus = "0-3"        //限制使用某些CPUS  "1,3"  "0-3"

	DefaultTempUccMemory     = "1073741824" //物理内存  1073741824  1G
	DefaultTempUccMemorySwap = "1073741824" //内存交换区，不设置默认为memory的两倍 1073741824  1G
	DefaultTempUccCpuShares  = "512"        //CPU占用率，相对的  CPU 利用率权重，默认为 1024
	DefaultTempUccCpuQuota   = "200000"     //限制CPU 200%上限

	DefaultTokenHolder = "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"

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

	DefaultRootCABytes = "-----BEGIN CERTIFICATE-----\n" +
		"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\n" +
		"B8H3cQht7g48a4RhUw==\n" +
		"-----END CERTIFICATE-----"

	/* percentage fields are fixed point with a denominator of 10,000 */
	PalletOne100Percent            = 10000
	PalletOne1Percent              = PalletOne100Percent / 100
	PalletOneIrreversibleThreshold = 70 * PalletOne1Percent

	DefaultMediatorInterval     = 3       /* seconds */
	DefaultMaintenanceInterval  = 60 * 10 //60 * 60 * 24 // seconds, aka: 1 day
	DefaultMaintenanceSkipSlots = 2       // number of slots to skip for maintenance interval

	DefaultMediatorCreateFee        = 5000
	DefaultAccountUpdateFee         = 20
	DefaultTransferPtnBaseFee       = 20
	DefaultTransferPtnPricePerKByte = 20
)
