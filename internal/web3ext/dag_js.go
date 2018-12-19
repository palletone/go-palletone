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

const Dag_JS = `
web3._extend({
	property: 'dag',
	methods: [
		new web3._extend.Method({
			name: 'getUnitByNumber',
        	call: 'ptn_getUnitByNumber',
        	params: 1,
        	inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getUnitByHash',
       		call: 'ptn_getUnitByHash',
        	params: 1,
        	inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getTransaction',
        	call: 'ptn_getTransactionByHash',
        	params: 1
		}),
		
	],
	properties: []
});
`
