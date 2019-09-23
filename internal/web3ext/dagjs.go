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
	Modules["dag"] = Dag_JS
}

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
        new web3._extend.Method({
            name: 'stableUnitNum',
            call: 'dag_stableUnitNum',
            params: 0,
        }),
        new web3._extend.Method({
            name: 'isSynced',
            call: 'dag_isSynced',
            params: 0,
        }),
		new web3._extend.Method({
            name: 'checkHeader',
            call: 'dag_checkHeader',
            params: 1,
        }),
		new web3._extend.Method({
            name: 'rebuildAddrTxIndex',
            call: 'dag_rebuildAddrTxIndex',
            params: 0,
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
