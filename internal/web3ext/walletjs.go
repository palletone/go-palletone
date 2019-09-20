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
	Modules["wallet"] = Wallet_JS
}

const Wallet_JS = `
 web3._extend({
 	property: 'wallet',
 	methods: [
 	],
 	properties:
 	[
 		new web3._extend.Method({
			name: 'getBalance',
			call: 'wallet_getBalance',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getBalance2',
			call: 'wallet_getBalance2',
			params: 1
		}),
		new web3._extend.Method({
            name: 'getAddrTxHistory',
            call: 'wallet_getAddrTxHistory',
            params: 1
        }),
		new web3._extend.Method({
			name: 'getAddrUtxos',
			call: 'wallet_getAddrUtxos',
			params: 1
		}),
 		new web3._extend.Method({
			name: 'getAddrUtxos2',
			call: 'wallet_getAddrUtxos2',
			params: 1
		}),
		new web3._extend.Method({
			name: 'createRawTransaction',
			call: 'wallet_createRawTransaction',
			params: 4
		}),
        new web3._extend.Method({
		    name: 'sendRawTransaction',
		    call: 'wallet_sendRawTransaction',
		    params: 1
		}),
        new web3._extend.Method({
		    name: 'signRawTransaction',
		    call: 'wallet_signRawTransaction',
		    params: 4
		}),	
        new web3._extend.Method({
		    name: 'sendRlpTransaction',
		    call: 'wallet_sendRlpTransaction',
		    params: 1
		}),
		new web3._extend.Method({
			name: 'sendJsonTransaction',
			call: 'wallet_sendJsonTransaction',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getPtnTestCoin',
			call: 'wallet_getPtnTestCoin',
			params: 5
		}),
		new web3._extend.Method({
			name: 'transferToken',
			call: 'wallet_transferToken',
			params: 8,
			inputFormatter: [null,null,null,null,null,null,null,null]
		}),
		new web3._extend.Method({
			name: 'transferPTN',
			call: 'wallet_transferPtn',
			params: 7,
			inputFormatter: [null,null,null,null,null,null,null]
		}),
		new web3._extend.Method({
			name: 'createProofTransaction',
			call: 'wallet_createProofTransaction',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'createProofOfExistenceTx',
			call: 'wallet_createProofOfExistenceTx',
			params: 5
		}),
        new web3._extend.Method({
			name: 'createTraceability',
			call: 'wallet_createTraceability',
			params: 6
		}),
		new web3._extend.Method({
			name: 'getProofOfExistencesByRef',
			call: 'wallet_getProofOfExistencesByRef',
			params: 1
		}),
        new web3._extend.Method({
			name: 'getProofOfExistencesByAsset',
			call: 'wallet_getProofOfExistencesByAsset',
			params: 1
		}),
        new web3._extend.Method({
			name: 'getFileInfoByTxid',
			call: 'wallet_getFileInfoByTxid',
			params: 1,
			inputFormatter: [null]
		}),
        new web3._extend.Method({
			name: 'getFileInfoByFileHash',
			call: 'wallet_getFileInfoByFileHash',
			params: 1,
			inputFormatter: [null]
		}),	
		new web3._extend.Method({
			name: 'getOneTokenInfo',
			call: 'wallet_getOneTokenInfo',
			params: 1,
			inputFormatter: [null]
		}),	
		new web3._extend.Method({
			name: 'getAllTokenInfo',
			call: 'wallet_getAllTokenInfo',
			params: 1,
			inputFormatter: [null]
		}),	
		new web3._extend.Method({
			name: 'getStxo',
			call: 'wallet_getStxo',
			params: 3
		}),	
		new web3._extend.Method({
			name: 'genCert',
			call: 'wallet_genCert',
			params: 7
		}),
        new web3._extend.Method({
			name: 'revokeCert',
			call: 'wallet_revokeCert',
			params: 3
		}),
        new web3._extend.Method({
            name: 'addBatchTxs',
            call: 'wallet_addBatchTxs',
            params: 1,
        }),
		new web3._extend.Method({
            name: 'aggregateUtxo',
            call: 'wallet_aggregateUtxo',
            params: 2,
        }),
		new web3._extend.Method({
            name: 'getAddrTokenFlow',
            call: 'wallet_getAddrTokenFlow',
            params: 2,
        }),
 	]
 });
 `
