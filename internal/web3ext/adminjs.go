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
	Modules["admin"] = Admin_JS
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
			name: 'addTrustedPeer',
			call: 'admin_addTrustedPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'removeTrustedPeer',
			call: 'admin_removeTrustedPeer',
			params: 1
		}),
		//new web3._extend.Method({
		//	name: 'exportChain',
		//	call: 'admin_exportChain',
		//	params: 1,
		//	inputFormatter: [null]
		//}),
		//new web3._extend.Method({
		//	name: 'sleepBlocks',
		//	call: 'admin_sleepBlocks',
		//	params: 2
		//}),
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
		//new web3._extend.Method({
		//	name: 'startWS',
		//	call: 'admin_startWS',
		//	params: 4,
		//	inputFormatter: [null, null, null, null]
		//}),
		//new web3._extend.Method({
		//	name: 'stopWS',
		//	call: 'admin_stopWS'
		//}),
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
