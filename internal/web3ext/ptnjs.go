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
	Modules["ptn"] = Ptn_JS
}

const Ptn_JS = `
web3._extend({
	property: 'ptn',
	methods: [
		//new web3._extend.Method({
		//	name: 'sign',
		//	call: 'ptn_sign',
		//	params: 2,
		//	inputFormatter: [web3._extend.formatters.inputAddressFormatter, null]
		//}),
		new web3._extend.Method({
			name: 'addressBalanceStatistics',
			call: 'ptn_addressBalanceStatistics',
			params: 2
		}),
		new web3._extend.Method({
			name: 'encodeTx',
			call: 'ptn_encodeTx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'decodeTx',
			call: 'ptn_decodeTx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'decodeJsonTx',
			call: 'ptn_decodeJsonTx',
			params: 1
		}),
		//new web3._extend.Method({
		//	name: 'resend',
		//	call: 'ptn_resend',
		//	params: 3,
		//	inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		//}),
		//new web3._extend.Method({
		//	name: 'signTransaction',
		//	call: 'ptn_signTransaction',
		//	params: 1,
		//	inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		//}),
		//new web3._extend.Method({
		//	name: 'submitTransaction',
		//	call: 'ptn_submitTransaction',
		//	params: 1,
		//	inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		//}),
		//new web3._extend.Method({
		//	name: 'getRawTransaction',
		//	call: 'ptn_getRawTransactionByHash',
		//	params: 1
		//}),
		//new web3._extend.Method({
		//	name: 'getRawTransactionFromBlock',
		//	call: function(args) {
		//		return (web3._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'ptn_getRawTransactionByBlockHashAndIndex' : 'ptn_getRawTransactionByBlockNumberAndIndex';
		//	},
		//	params: 2,
		//	inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, web3._extend.utils.toHex]
		//}),
		
		//new web3._extend.Method({
		//	name: 'setJuryAccount',
        	//call: 'ptn_setJuryAccount',
        	//params: 2, //address, password string
		//	inputFormatter: [null, null]
		//}),
		//new web3._extend.Method({
		//	name: 'getJuryAccount',
        	//call: 'ptn_getJuryAccount',
        	//params: 0, //
		//	inputFormatter: []
		//}),
		

		//new web3._extend.Method({
		//	name: 'cmdCreateTransaction',
		//	call: 'ptn_cmdCreateTransaction',
		//	params: 4,
		//	inputFormatter: [null,null,null, null]
		//}),
		//new web3._extend.Method({
		//	name: 'createRawTransaction',
		//	call: 'ptn_createRawTransaction',
		//	params: 1,
		//	inputFormatter: [null]
		//}),
		//new web3._extend.Method({
		//	name: 'signRawTransaction',
		//	call: 'ptn_signRawTransaction',
		//	params: 3,
		//	inputFormatter: [null,null, null]
		//}),
		//new web3._extend.Method({
		//	name: 'sendRawTransaction',
		//	call: 'ptn_sendRawTransaction',
		//	params: 1,
		//	inputFormatter: [null]
		//}),
         //       new web3._extend.Method({
		//	name: 'sendJsonTransaction',
		//	call: 'ptn_sendJsonTransaction',
		//	params: 1,
		//	inputFormatter: [null]
		//}),

		//new web3._extend.Method({
		//	name: 'getBalance',
		//	call: 'ptn_getBalance',
		//	params: 1,
		//	inputFormatter: [null]
		//}),
  		new web3._extend.Method({
			name: 'getTokenTxHistory',
			call: 'ptn_getTokenTxHistory',
			params: 1,
			inputFormatter: [null]
		}),
		//new web3._extend.Method({
		//	name: 'getTransactionsByTxid',
         //   call: 'ptn_getTransactionsByTxid',
		//	params: 1,
		//	inputFormatter: [null]
		//}),
		//new web3._extend.Method({
		//	name: 'election',
		//	call: 'ptn_election',
		//	params: 1,			
		//}),
		new web3._extend.Method({
			name: 'getProofTxInfoByHash',
			call: 'ptn_getProofTxInfoByHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'proofTransactionByHash',
			call: 'ptn_proofTransactionByHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'proofTransactionByRlptx',
			call: 'ptn_proofTransactionByRlptx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'syncUTXOByAddr',
			call: 'ptn_syncUTXOByAddr',
			params: 1
		}),
		//new web3._extend.Method({
		//	name: 'ccstartChaincodeContainer',
		//	call: 'ptn_ccstartChaincodeContainer',
		//	params: 2,
		//	inputFormatter: [null,null]
		//}),
		new web3._extend.Method({
			name: 'listSysConfig',
			call: 'ptn_listSysConfig'
			params: 0,
		}),
		new web3._extend.Method({
			name: 'getPledge',
			call: 'ptn_getPledge'
			params: 1,
		}),
	],

	properties: [
		//new web3._extend.Property({
		//	name: 'pendingTransactions',
		//	getter: 'ptn_pendingTransactions',
		//	outputFormatter: function(txs) {
		//		var formatted = [];
		//		for (var i = 0; i < txs.length; i++) {
		//			formatted.push(web3._extend.formatters.outputTransactionFormatter(txs[i]));
		//			formatted[i].blockHash = null;
		//		}
		//		return formatted;
		//	}
		//}),
		new web3._extend.Property({
			name: 'getChainParameters',
			getter: 'ptn_getChainParameters'
		}),
		new web3._extend.Property({
			name: 'syncing',
			getter: 'ptn_syncing'
		}),
	]
});
`
