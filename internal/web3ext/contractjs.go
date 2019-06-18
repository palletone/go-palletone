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
 *  * @date 2018
 *
 */

package web3ext

func init() {
	Modules["contract"] = Contract_JS
}

const Contract_JS = `
web3._extend({
	property: 'contract',
	methods: [
	   //ccTx
		new web3._extend.Method({
			name: 'ccinstalltx',
        	call: 'contract_ccinstalltx',
        	params: 7, //from, to , daoAmount, daoFee , tplName, path, version
			inputFormatter: [null, null, null,null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccdeploytx',
        	call: 'contract_ccdeploytx',
        	params: 6, //from, to , daoAmount, daoFee , templateId , args  
			inputFormatter: [null, null, null,null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccinvoketx',
        	call: 'contract_ccinvoketx',
        	params: 8, //from, to, daoAmount, daoFee , contractAddr, args[]string------>["fun", "key", "value"], certid, timeout
			inputFormatter: [null, null, null,null, null, null, null, null]
		}),
        new web3._extend.Method({
			name: 'ccinvoketxPass',
			call: 'contract_ccinvoketxPass',
			params: 9, //from, to, daoAmount, daoFee , contractAddr, args[]string------>["fun", "key", "value"],passwd,duration, certid
			inputFormatter: [null, null, null,null, null, null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccinvokeToken',
        	call: 'contract_ccinvokeToken',
        	params: 9, //from, to, toToken, daoAmount, daoFee, daoAmountToken, assetToken, contractAddr, args[]string------>["fun", "key", "value"]
			inputFormatter: [null, null, null,null, null, null,null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccstoptx',
        	call: 'contract_ccstoptx',
        	params: 6, //from, to, daoAmount, daoFee, contractId, deleteImage
			inputFormatter: [null, null, null, null, null, null]
		}),
		//cc
		new web3._extend.Method({
			name: 'ccinstall',
        	call: 'contract_ccinstall',
        	params: 3, //ccName string, ccPath string, ccVersion string
			inputFormatter: [null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccdeploy',
        	call: 'contract_ccdeploy',
        	params: 3, //templateId, args []string ---->["init", "a", "1", "b", 10], timeout uint32
			inputFormatter: [null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccinvoke',
        	call: 'contract_ccinvoke',
        	params: 3, // contractAddr, args[]string------>["fun", "key", "value"], timeout uint32
			inputFormatter: [null, null, null]
		}),
		new web3._extend.Method({
			name: 'ccquery',
			call: 'contract_ccquery',
			params: 3, //contractAddr,args[]string---->["func","arg1","arg2","..."], timeout uint32
			inputFormatter: [null,null, null]
		}),
		new web3._extend.Method({
			name: 'ccstop',
        	call: 'contract_ccstop',
        	params: 1, //contractId
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'depositContractInvoke',
        	call: 'contract_depositContractInvoke',
        	params: 5, //from, to, daoAmount, daoFee,param[]string
			inputFormatter: [null, null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'depositContractQuery',
        	call: 'contract_depositContractQuery',
        	params: 1, //param[]string
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getContractsByTpl',
        	call: 'contract_getContractsByTpl',
        	params: 1, //param[]string
			inputFormatter: [null]
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'listAllContractTemplates',
			getter: 'contract_listAllContractTemplates'
		}),
		new web3._extend.Property({
			name: 'listAllContracts',
			getter: 'contract_listAllContracts'
		}),
	]
});
`
