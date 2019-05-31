// Copyright 2018 PalletOne

// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// package web3ext contains gptn specific web3.js extensions.

package web3ext

var Modules = map[string]string{
	"admin":      Admin_JS,
	"chequebook": Chequebook_JS,
	"clique":     Clique_JS,
	"ptn":        Ptn_JS,
	"dag":        Dag_JS,
	"net":        Net_JS,
	"personal":   Personal_JS,
	"rpc":        RPC_JS,
	"wallet":     Wallet_JS,
	"txpool":     TxPool_JS,
}

const Admin_JS = `
web3._extend({
	property: 'admin',
	methods: [
		new web3._extend.Method({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new web3._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
		new web3._extend.Method({
			name: 'addCorsPeer',
			call: 'admin_addCorsPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'corsPeers',
			call: 'admin_corsPeers',
			params: 1
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new web3._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
		new web3._extend.Property({
			name: 'corsInfo',
			getter: 'admin_corsInfo'
		}),
	]
});
`
const Dag_JS = `
web3._extend({
	property: 'dag',
	methods: [
		new web3._extend.Method({
			name: 'getUnitByNumber',
        	call: 'dag_getUnitByNumber',
        	params: 1,
        	inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getUnitsByIndex',
        	call: 'dag_getUnitsByIndex',
        	params: 3,
        	//inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getUnitSummaryByNumber',
        	call: 'dag_getUnitSummaryByNumber',
        	params: 1,
        	inputFormatter: [null]
		}),
 		new web3._extend.Method({
            name: 'getUnstableUnits',
            call: 'dag_getUnstableUnits',
            params: 0,
        }),
		new web3._extend.Method({
			name: 'getUnitByHash',
       		call: 'dag_getUnitByHash',
        	params: 1,
        	inputFormatter: [null]
		}),
        new web3._extend.Method({
			name: 'getHeaderByHash',
       		call: 'dag_getHeaderByHash',
        	params: 1,
        	inputFormatter: [null]
		}),
        new web3._extend.Method({
			name: 'getHeaderByNumber',
       		call: 'dag_getHeaderByNumber',
        	params: 1,
        	inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getTransaction',
        	call: 'dag_getTransactionByHash',
        	params: 1
		}),
		new web3._extend.Method({
		    name: 'getTxHashByReqId',
		    call: 'dag_getTxHashByReqId',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
            name: 'getHexCommon',
            call: 'dag_getHexCommon',
            params: 1,
        }),
		new web3._extend.Method({
            name: 'getCommon',
            call: 'dag_getCommon',
            params: 1,
        }),
        new web3._extend.Method({
            name: 'getCommonByPrefix',
            call: 'dag_getCommonByPrefix',
            params: 1,
        }),
        new web3._extend.Method({   
            name: 'getTxByHash', 
            call: 'dag_getTxByHash',   
            params: 1,  
            // inputFormatter: [null]
        }),   
 		new web3._extend.Method({   
            name: 'getTxByReqId', 
            call: 'dag_getTxByReqId',   
            params: 1,
        }),   
        new web3._extend.Method({   
            name: 'getTxSearchEntry',  
            call: 'dag_getTxSearchEntry',   
            params: 1,  
            // inputFormatter: [null]
        }), 
        new web3._extend.Method({
            name: 'getTxPoolTxByHash',
            call: 'dag_getTxPoolTxByHash',  
            params: 1,
            // inputFormatter: [null, formatters.inputDefaultBlockNumberFormatter],
            // outputFormatter: utils.toDecimal
        }), 
        new web3._extend.Method({
            name: 'getUtxoEntry',
            call: 'dag_getUtxoEntry',
            params: 1,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getAddrOutput',
            call: 'dag_getAddrOutput',
            params: 1,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getAddrOutpoints',
            call: 'dag_getAddrOutpoints',
            params: 1,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getAddrUtxos',
            call: 'dag_getAddrUtxos',
            params: 1,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getAllUtxos',
            call: 'dag_getAllUtxos',
            params: 0,
            // inputFormatter: [null]
        }),        
        new web3._extend.Method({
            name: 'getTokenInfo',
            call: 'dag_getTokenInfo',
            params: 1,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getAllTokenInfo',
            call: 'dag_getAllTokenInfo',
            params: 0,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'saveTokenInfo',
            call: 'dag_saveTokenInfo',
            params: 3,
            //inputFormatter: [null]
        }),
       
        new web3._extend.Method({
            name: 'getHeadHeaderHash',
            call: 'dag_getHeadHeaderHash',
            params: 0,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getHeadUnitHash',
            call: 'dag_getHeadUnitHash',
            params: 0,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getHeadFastUnitHash',
            call: 'dag_getHeadFastUnitHash',
            params: 0,
            // inputFormatter: [null]
        }),
        new web3._extend.Method({
            name: 'getFastUnitIndex',
            call: 'dag_getFastUnitIndex',
            params: 1,
            // inputFormatter: [null]
        }),
	],
	properties: [
		new web3._extend.Property({
			name: 'headUnitTime',
			getter: 'dag_headUnitTime'
		}),
		new web3._extend.Property({
			name: 'headUnitNum',
			getter: 'dag_headUnitNum'
		}),
		new web3._extend.Property({
			name: 'headUnitHash',
			getter: 'dag_headUnitHash'
		}),
	]
});
`
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
		//new web3._extend.Method({
		//	name: 'batchSign',
		//	call: 'ptn_batchSign',
		//	params: 6
		//}),
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
		
		new web3._extend.Method({
			name: 'setJuryAccount',
        	call: 'ptn_setJuryAccount',
        	params: 2, //address, password string
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'getJuryAccount',
        	call: 'ptn_getJuryAccount',
        	params: 0, //
			inputFormatter: []
		}),
		

		new web3._extend.Method({
			name: 'cmdCreateTransaction',
			call: 'ptn_cmdCreateTransaction',
			params: 4,
			inputFormatter: [null,null,null, null]
		}),
		new web3._extend.Method({
			name: 'createRawTransaction',
			call: 'ptn_createRawTransaction',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'signRawTransaction',
			call: 'ptn_signRawTransaction',
			params: 3,
			inputFormatter: [null,null, null]
		}),
		new web3._extend.Method({
			name: 'sendRawTransaction',
			call: 'ptn_sendRawTransaction',
			params: 1,
			inputFormatter: [null]
		}),

        new web3._extend.Method({
			name: 'getBalance',
			call: 'ptn_getBalance',
			params: 1,
			inputFormatter: [null]
		}),
  		new web3._extend.Method({
			name: 'getTokenTxHistory',
			call: 'ptn_getTokenTxHistory',
			params: 1,
			inputFormatter: [null]
		}),
        new web3._extend.Method({
			name: 'getTransactionsByTxid',
            call: 'ptn_getTransactionsByTxid',
			params: 1,
			inputFormatter: [null]
		}),
        new web3._extend.Method({
			name: 'election',
			call: 'ptn_election',
			params: 1,			
		}),
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
		new web3._extend.Method({
			name: 'ccstartChaincodeContainer',
			call: 'ptn_ccstartChaincodeContainer',
			params: 2,
			inputFormatter: [null,null]
		}),
	],

	properties: [
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'ptn_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(web3._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
	new web3._extend.Property({
			name: 'listSysConfig',
			getter: 'ptn_listSysConfig'
		}),
	]
});
`

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
            name: 'getAddrTxHistory',
            call: 'wallet_getAddrTxHistory',
            params: 1
        }),
		new web3._extend.Method({
			name: 'getAddrUtxos',
			call: 'wallet_getAddrUtxos',
			params: 1
		}),
 		//new web3._extend.Method({
		//	name: 'createPaymentTx',
		//	call: 'wallet_createPaymentTx',
		//	params: 4
		//}),
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
 	]
 });
 `

const Chequebook_JS = `
web3._extend({
	property: 'chequebook',
	methods: [
		new web3._extend.Method({
			name: 'deposit',
			call: 'chequebook_deposit',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Property({
			name: 'balance',
			getter: 'chequebook_balance',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Method({
			name: 'cash',
			call: 'chequebook_cash',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'issue',
			call: 'chequebook_issue',
			params: 2,
			inputFormatter: [null, null]
		}),
	]
});
`

const Clique_JS = `
web3._extend({
	property: 'clique',
	methods: [
		new web3._extend.Method({
			name: 'getSnapshot',
			call: 'clique_getSnapshot',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSnapshotAtHash',
			call: 'clique_getSnapshotAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getSigners',
			call: 'clique_getSigners',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSignersAtHash',
			call: 'clique_getSignersAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'propose',
			call: 'clique_propose',
			params: 2
		}),
		new web3._extend.Method({
			name: 'discard',
			call: 'clique_discard',
			params: 1
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'proposals',
			getter: 'clique_proposals'
		}),
	]
});
`

const Net_JS = `
web3._extend({
	property: 'net',
	methods: [],
	properties: [
		new web3._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
	]
});
`

const RPC_JS = `
web3._extend({
	property: 'rpc',
	methods: [],
	properties: [
		new web3._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const TxPool_JS = `
web3._extend({
	property: 'txpool',
	methods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new web3._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new web3._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = web3._extend.utils.toDecimal(status.pending);
      			status.orphans = web3._extend.utils.toDecimal(status.orphans);
				status.queued = web3._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
		new web3._extend.Property({
			name: 'pending',
			getter: 'txpool_pending'
		}),
		new web3._extend.Property({
			name: 'queue',
			getter: 'txpool_queue'
		}),
	]
});
`
