package txscript

import (
	"github.com/palletone/go-palletone/common"
	"strings"
	"github.com/palletone/go-palletone/common/crypto"
)

/**
scriptPubKey could be like
"scriptPubKey":{
            "asm":"OP_DUP OP_HASH160 aba7915d5964406e8a02c3202f1f8a4a63e95c13 OP_EQUALVERIFY OP_CHECKSIG",
            "hex":"76a914aba7915d5964406e8a02c3202f1f8a4a63e95c1388ac",
            "reqSigs":1,
            "type":"pubkeyhash",
            "addresses":[
               "1GedHcxdxq2tab98hqAmREUK9BBYHKznof"
            ]
         }
*/

const (
	T_P2PKH = "pubkeyhash"
	T_P2SH = "scripthash"
)

const (
	OP_0                   = 0x00 // 0
	OP_FALSE               = 0x00 // 0 - AKA OP_0
	OP_DATA_1              = 0x01 // 1
	OP_DATA_2              = 0x02 // 2
	OP_DATA_3              = 0x03 // 3
	OP_DATA_4              = 0x04 // 4
	OP_DATA_5              = 0x05 // 5
	OP_DATA_6              = 0x06 // 6
	OP_DATA_7              = 0x07 // 7
	OP_DATA_8              = 0x08 // 8
	OP_DATA_9              = 0x09 // 9
	OP_DATA_10             = 0x0a // 10
	OP_DATA_11             = 0x0b // 11
	OP_DATA_12             = 0x0c // 12
	OP_DATA_13             = 0x0d // 13
	OP_DATA_14             = 0x0e // 14
	OP_DATA_15             = 0x0f // 15
	OP_DATA_16             = 0x10 // 16
	OP_DATA_17             = 0x11 // 17
	OP_DATA_18             = 0x12 // 18
	OP_DATA_19             = 0x13 // 19
	OP_DATA_20             = 0x14 // 20
	OP_DATA_21             = 0x15 // 21
	OP_DATA_22             = 0x16 // 22
	OP_DATA_23             = 0x17 // 23
	OP_DATA_24             = 0x18 // 24
	OP_DATA_25             = 0x19 // 25
	OP_DATA_26             = 0x1a // 26
	OP_DATA_27             = 0x1b // 27
	OP_DATA_28             = 0x1c // 28
	OP_DATA_29             = 0x1d // 29
	OP_DATA_30             = 0x1e // 30
	OP_DATA_31             = 0x1f // 31
	OP_DATA_32             = 0x20 // 32
	OP_DATA_33             = 0x21 // 33
	OP_DATA_34             = 0x22 // 34
	OP_DATA_35             = 0x23 // 35
	OP_DATA_36             = 0x24 // 36
	OP_DATA_37             = 0x25 // 37
	OP_DATA_38             = 0x26 // 38
	OP_DATA_39             = 0x27 // 39
	OP_DATA_40             = 0x28 // 40
	OP_DATA_41             = 0x29 // 41
	OP_DATA_42             = 0x2a // 42
	OP_DATA_43             = 0x2b // 43
	OP_DATA_44             = 0x2c // 44
	OP_DATA_45             = 0x2d // 45
	OP_DATA_46             = 0x2e // 46
	OP_DATA_47             = 0x2f // 47
	OP_DATA_48             = 0x30 // 48
	OP_DATA_49             = 0x31 // 49
	OP_DATA_50             = 0x32 // 50
	OP_DATA_51             = 0x33 // 51
	OP_DATA_52             = 0x34 // 52
	OP_DATA_53             = 0x35 // 53
	OP_DATA_54             = 0x36 // 54
	OP_DATA_55             = 0x37 // 55
	OP_DATA_56             = 0x38 // 56
	OP_DATA_57             = 0x39 // 57
	OP_DATA_58             = 0x3a // 58
	OP_DATA_59             = 0x3b // 59
	OP_DATA_60             = 0x3c // 60
	OP_DATA_61             = 0x3d // 61
	OP_DATA_62             = 0x3e // 62
	OP_DATA_63             = 0x3f // 63
	OP_DATA_64             = 0x40 // 64
	OP_DATA_65             = 0x41 // 65
	OP_DATA_66             = 0x42 // 66
	OP_DATA_67             = 0x43 // 67
	OP_DATA_68             = 0x44 // 68
	OP_DATA_69             = 0x45 // 69
	OP_DATA_70             = 0x46 // 70
	OP_DATA_71             = 0x47 // 71
	OP_DATA_72             = 0x48 // 72
	OP_DATA_73             = 0x49 // 73
	OP_DATA_74             = 0x4a // 74
	OP_DATA_75             = 0x4b // 75
	OP_PUSHDATA1           = 0x4c // 76
	OP_PUSHDATA2           = 0x4d // 77
	OP_PUSHDATA4           = 0x4e // 78
	OP_1NEGATE             = 0x4f // 79
	OP_RESERVED            = 0x50 // 80
	OP_1                   = 0x51 // 81 - AKA OP_TRUE
	OP_TRUE                = 0x51 // 81
	OP_2                   = 0x52 // 82
	OP_3                   = 0x53 // 83
	OP_4                   = 0x54 // 84
	OP_5                   = 0x55 // 85
	OP_6                   = 0x56 // 86
	OP_7                   = 0x57 // 87
	OP_8                   = 0x58 // 88
	OP_9                   = 0x59 // 89
	OP_10                  = 0x5a // 90
	OP_11                  = 0x5b // 91
	OP_12                  = 0x5c // 92
	OP_13                  = 0x5d // 93
	OP_14                  = 0x5e // 94
	OP_15                  = 0x5f // 95
	OP_16                  = 0x60 // 96
	OP_NOP                 = 0x61 // 97
	OP_VER                 = 0x62 // 98
	OP_IF                  = 0x63 // 99
	OP_NOTIF               = 0x64 // 100
	OP_VERIF               = 0x65 // 101
	OP_VERNOTIF            = 0x66 // 102
	OP_ELSE                = 0x67 // 103
	OP_ENDIF               = 0x68 // 104
	OP_VERIFY              = 0x69 // 105
	OP_RETURN              = 0x6a // 106
	OP_TOALTSTACK          = 0x6b // 107
	OP_FROMALTSTACK        = 0x6c // 108
	OP_2DROP               = 0x6d // 109
	OP_2DUP                = 0x6e // 110
	OP_3DUP                = 0x6f // 111
	OP_2OVER               = 0x70 // 112
	OP_2ROT                = 0x71 // 113
	OP_2SWAP               = 0x72 // 114
	OP_IFDUP               = 0x73 // 115
	OP_DEPTH               = 0x74 // 116
	OP_DROP                = 0x75 // 117
	OP_DUP                 = 0x76 // 118
	OP_NIP                 = 0x77 // 119
	OP_OVER                = 0x78 // 120
	OP_PICK                = 0x79 // 121
	OP_ROLL                = 0x7a // 122
	OP_ROT                 = 0x7b // 123
	OP_SWAP                = 0x7c // 124
	OP_TUCK                = 0x7d // 125
	OP_CAT                 = 0x7e // 126
	OP_SUBSTR              = 0x7f // 127
	OP_LEFT                = 0x80 // 128
	OP_RIGHT               = 0x81 // 129
	OP_SIZE                = 0x82 // 130
	OP_INVERT              = 0x83 // 131
	OP_AND                 = 0x84 // 132
	OP_OR                  = 0x85 // 133
	OP_XOR                 = 0x86 // 134
	OP_EQUAL               = 0x87 // 135
	OP_EQUALVERIFY         = 0x88 // 136
	OP_RESERVED1           = 0x89 // 137
	OP_RESERVED2           = 0x8a // 138
	OP_1ADD                = 0x8b // 139
	OP_1SUB                = 0x8c // 140
	OP_2MUL                = 0x8d // 141
	OP_2DIV                = 0x8e // 142
	OP_NEGATE              = 0x8f // 143
	OP_ABS                 = 0x90 // 144
	OP_NOT                 = 0x91 // 145
	OP_0NOTEQUAL           = 0x92 // 146
	OP_ADD                 = 0x93 // 147
	OP_SUB                 = 0x94 // 148
	OP_MUL                 = 0x95 // 149
	OP_DIV                 = 0x96 // 150
	OP_MOD                 = 0x97 // 151
	OP_LSHIFT              = 0x98 // 152
	OP_RSHIFT              = 0x99 // 153
	OP_BOOLAND             = 0x9a // 154
	OP_BOOLOR              = 0x9b // 155
	OP_NUMEQUAL            = 0x9c // 156
	OP_NUMEQUALVERIFY      = 0x9d // 157
	OP_NUMNOTEQUAL         = 0x9e // 158
	OP_LESSTHAN            = 0x9f // 159
	OP_GREATERTHAN         = 0xa0 // 160
	OP_LESSTHANOREQUAL     = 0xa1 // 161
	OP_GREATERTHANOREQUAL  = 0xa2 // 162
	OP_MIN                 = 0xa3 // 163
	OP_MAX                 = 0xa4 // 164
	OP_WITHIN              = 0xa5 // 165
	OP_RIPEMD160           = 0xa6 // 166
	OP_SHA1                = 0xa7 // 167
	OP_SHA256              = 0xa8 // 168
	OP_HASH160             = 0xa9 // 169
	OP_HASH256             = 0xaa // 170
	OP_CODESEPARATOR       = 0xab // 171
	OP_CHECKSIG            = 0xac // 172
	OP_CHECKSIGVERIFY      = 0xad // 173
	OP_CHECKMULTISIG       = 0xae // 174
	OP_CHECKMULTISIGVERIFY = 0xaf // 175
	OP_NOP1                = 0xb0 // 176
	OP_NOP2                = 0xb1 // 177
	OP_CHECKLOCKTIMEVERIFY = 0xb1 // 177 - AKA OP_NOP2
	OP_NOP3                = 0xb2 // 178
	OP_CHECKSEQUENCEVERIFY = 0xb2 // 178 - AKA OP_NOP3
	OP_NOP4                = 0xb3 // 179
	OP_NOP5                = 0xb4 // 180
	OP_NOP6                = 0xb5 // 181
	OP_NOP7                = 0xb6 // 182
	OP_NOP8                = 0xb7 // 183
	OP_NOP9                = 0xb8 // 184
	OP_NOP10               = 0xb9 // 185
	OP_UNKNOWN186          = 0xba // 186
	OP_UNKNOWN187          = 0xbb // 187
	OP_UNKNOWN188          = 0xbc // 188
	OP_UNKNOWN189          = 0xbd // 189
	OP_UNKNOWN190          = 0xbe // 190
	OP_UNKNOWN191          = 0xbf // 191
	OP_UNKNOWN192          = 0xc0 // 192
	OP_UNKNOWN193          = 0xc1 // 193
	OP_UNKNOWN194          = 0xc2 // 194
	OP_UNKNOWN195          = 0xc3 // 195
	OP_UNKNOWN196          = 0xc4 // 196
	OP_UNKNOWN197          = 0xc5 // 197
	OP_UNKNOWN198          = 0xc6 // 198
	OP_UNKNOWN199          = 0xc7 // 199
	OP_UNKNOWN200          = 0xc8 // 200
	OP_UNKNOWN201          = 0xc9 // 201
	OP_UNKNOWN202          = 0xca // 202
	OP_UNKNOWN203          = 0xcb // 203
	OP_UNKNOWN204          = 0xcc // 204
	OP_UNKNOWN205          = 0xcd // 205
	OP_UNKNOWN206          = 0xce // 206
	OP_UNKNOWN207          = 0xcf // 207
	OP_UNKNOWN208          = 0xd0 // 208
	OP_UNKNOWN209          = 0xd1 // 209
	OP_UNKNOWN210          = 0xd2 // 210
	OP_UNKNOWN211          = 0xd3 // 211
	OP_UNKNOWN212          = 0xd4 // 212
	OP_UNKNOWN213          = 0xd5 // 213
	OP_UNKNOWN214          = 0xd6 // 214
	OP_UNKNOWN215          = 0xd7 // 215
	OP_UNKNOWN216          = 0xd8 // 216
	OP_UNKNOWN217          = 0xd9 // 217
	OP_UNKNOWN218          = 0xda // 218
	OP_UNKNOWN219          = 0xdb // 219
	OP_UNKNOWN220          = 0xdc // 220
	OP_UNKNOWN221          = 0xdd // 221
	OP_UNKNOWN222          = 0xde // 222
	OP_UNKNOWN223          = 0xdf // 223
	OP_UNKNOWN224          = 0xe0 // 224
	OP_UNKNOWN225          = 0xe1 // 225
	OP_UNKNOWN226          = 0xe2 // 226
	OP_UNKNOWN227          = 0xe3 // 227
	OP_UNKNOWN228          = 0xe4 // 228
	OP_UNKNOWN229          = 0xe5 // 229
	OP_UNKNOWN230          = 0xe6 // 230
	OP_UNKNOWN231          = 0xe7 // 231
	OP_UNKNOWN232          = 0xe8 // 232
	OP_UNKNOWN233          = 0xe9 // 233
	OP_UNKNOWN234          = 0xea // 234
	OP_UNKNOWN235          = 0xeb // 235
	OP_UNKNOWN236          = 0xec // 236
	OP_UNKNOWN237          = 0xed // 237
	OP_UNKNOWN238          = 0xee // 238
	OP_UNKNOWN239          = 0xef // 239
	OP_UNKNOWN240          = 0xf0 // 240
	OP_UNKNOWN241          = 0xf1 // 241
	OP_UNKNOWN242          = 0xf2 // 242
	OP_UNKNOWN243          = 0xf3 // 243
	OP_UNKNOWN244          = 0xf4 // 244
	OP_UNKNOWN245          = 0xf5 // 245
	OP_UNKNOWN246          = 0xf6 // 246
	OP_UNKNOWN247          = 0xf7 // 247
	OP_UNKNOWN248          = 0xf8 // 248
	OP_UNKNOWN249          = 0xf9 // 249
	OP_SMALLINTEGER        = 0xfa // 250 - bitcoin core internal
	OP_PUBKEYS             = 0xfb // 251 - bitcoin core internal
	OP_UNKNOWN252          = 0xfc // 252
	OP_PUBKEYHASH          = 0xfd // 253 - bitcoin core internal
	OP_PUBKEY              = 0xfe // 254 - bitcoin core internal
	OP_INVALIDOPCODE       = 0xff // 255 - bitcoin core internal
)

type ScriptPubKey struct {
	Asm     []byte
	Hex     []byte //could be like 50 bytes
	RegSigs uint16 // signature num
	Type    string
	Address common.Address // utxo holder address
}

/**
从P2SH（多签）或者P2PKH（单签）
P2PKH Example: OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
P2SH Example: OP_HASH160 <scripthash> OP_EQUAL
 */
func ExtractPkScriptAddrs(pkScript []byte) (*ScriptPubKey, error) {
	spk := ScriptPubKey{}
	if len(pkScript)==25 {
		spk.Type = T_P2PKH
		pubkeyBytes := pkScript[3:23]
		pubkey := crypto.ToECDSAPub(pubkeyBytes)
		spk.Address = crypto.PubkeyToAddress(*pubkey)
	}


	//reg := regexp.MustCompile(`OP_DUP OP_HASH160 ([\w]+) OP_EQUALVERIFY OP_CHECKSIG`)
	//shReq := regexp.MustCompile(`OP_HASH160 ([\w]+) OP_EQUAL`)
	//// now just decode P2PKH
	//results := reg.FindStringSubmatch(string(pkScript))
	//if len(results)==2 {
	//	spk.Type = T_P2PKH
	//	pubkey := crypto.ToECDSAPub([]byte(results[0]))
	//	if pubkey==nil {return nil, fmt.Errorf("Get pubkey from bytes error.")}
	//	spk.Address = crypto.PubkeyToAddress(*pubkey)
	//
	//} else {
	//	results = shReq.FindStringSubmatch(string(pkScript))
	//	if len(results)==2 {
	//		spk.Type = T_P2SH
	//		spk.Address = crypto.ScriptHashToAddress([]byte(results[1]))
	//	} else {
	//		return &spk, fmt.Errorf("Extract PkScript Addrs error")
	//	}
	//}

	return &spk, nil
}

/**
生成P2PKH脚本
 */
func PayToPubkeyHashScript(pubkey []byte) []byte {
	// copy data
	pkScript := make([]byte, 5+len(pubkey))
	pkScript[0] = OP_DUP
	pkScript[1] = OP_HASH160
	pkScript[2] = OP_DATA_20
	copy(pkScript[3:], pubkey)
	pkScript[23] = OP_EQUALVERIFY
	pkScript[24] = OP_CHECKSIG
	return pkScript
}

func CheckP2PKHAddress(addr common.Address)  bool{
	prefix := string(addr[:2])
	if strings.Compare(prefix,"P1")!=0 {return false}
	return true
}
