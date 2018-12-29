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
 * @author PalletOne core developer YiRan <dev@pallet.one>
 * @date 2018
 */

package vote

// 0.default vote result is the index of the option from list
// 1.If the option is specified by the voter, set Option null
// 2.Expected vote result:[]byte
//type VoteInitiatePayload struct {
//	Title       string        //vote title
//	Option      []string      //vote option list.
//	BallotChain uint64        //vote chain id
//	BallotType  IDType16      //vote asset id
//	BallotCost  big.Int       //token cost
//	ExpiredTime time.Duration //duration of voting
//}

//VotePayload YiRan@
// Mode == 0 [ Replace ] :replace all
// Mode == 1 [ Edit    ] :Replace the addresses in the first half of the account's votes addresses to refer to the addresses in the second half.
// Mode == 2 [ Delete  ] :Delete the addresses from account's votes addresses

type VoteInfo struct {
	//投票内容，JSON格式，如果是Mediator选举，则为string数组的JSON
	//如果不想投任何一个Mediator，则将string数组清空
	VoteType uint8
	//是Mediator选举则为0
	Contents []byte
	//Mode     uint8
}

// system vote 0-9
const TypeMediator = 0 // VoteContent = []byte(Common.Address)
//const TYPE_ADDRESS = 1  // VoteContent = []byte(Common.Address)
// 10-19
//const TYPE_LEN_2 = 10
//const TYPE_LEN_4 = 11
//const TYPE_LEN_8 = 12
//const TYPE_LEN_16 = 13
//const TYPE_LEN_32 = 14
//const TYPE_LEN_64 = 15
//const TYPE_LEN_128 = 16
//const TYPE_LEN_256 = 17
//const TYPE_LEN_512 = 18
//const TYPE_LEN_1024 = 19
// vote detail 200 - 239
//const TYPE_BOOL_TRUE = 200
//const TYPE_BOOL_FALSE = 201
//const TYPE_CHAR_A = 202
//const TYPE_CHAR_B = 203
//const TYPE_CHAR_C = 204
//const TYPE_CHAR_D = 205
//const TYPE_CHAR_E = 206
//const TYPE_CHAR_F = 207
//const TYPE_CHAR_G = 208
//const TYPE_CHAR_H = 209
//const TYPE_CHAR_I = 210
//const TYPE_CHAR_J = 211
//const TYPE_CHAR_K = 212
//const TYPE_CHAR_L = 213
//const TYPE_CHAR_M = 214
//const TYPE_CHAR_N = 215
//const TYPE_CHAR_O = 216
//const TYPE_CHAR_P = 217
//const TYPE_CHAR_Q = 218
//const TYPE_CHAR_R = 219
//const TYPE_CHAR_S = 220
//const TYPE_CHAR_T = 221
//const TYPE_CHAR_U = 222
//const TYPE_CHAR_V = 223
//const TYPE_CHAR_W = 224
//const TYPE_CHAR_X = 225
//const TYPE_CHAR_Y = 226
//const TYPE_CHAR_Z = 227
//const TYPE_NUM_0 = 228
//const TYPE_NUM_1 = 229
//const TYPE_NUM_2 = 230
//const TYPE_NUM_3 = 231
//const TYPE_NUM_4 = 232
//const TYPE_NUM_5 = 233
//const TYPE_NUM_6 = 234
//const TYPE_NUM_7 = 235
//const TYPE_NUM_8 = 236
//const TYPE_NUM_9 = 237
//const TYPE_NULL = 238
// action 240- 255
//const TYPE_VOTEING = 254
//const TYPE_CREATEVOTE = 255
