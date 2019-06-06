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

package web3ext

func init() {
	Modules["personal"] = Personal_JS
}

const Personal_JS = `
web3._extend({
	property: 'personal',
	methods: [
		new web3._extend.Method({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
		new web3._extend.Method({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, null, null]
		}),
		new web3._extend.Method({
			name: 'openWallet',
			call: 'personal_openWallet',
			params: 2
		}),
		//new web3._extend.Method({
		//	name: 'deriveAccount',
		//	call: 'personal_deriveAccount',
		//	params: 3
		//}),
		//new web3._extend.Method({
		//	name: 'signTransaction',
		//	call: 'personal_signTransaction',
		//	params: 2,
		//	inputFormatter: [web3._extend.formatters.inputTransactionFormatter, null]
		//}),
		new web3._extend.Method({
			name: 'transferPtn',
			call: 'personal_transferPtn',
			params: 5,
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`
