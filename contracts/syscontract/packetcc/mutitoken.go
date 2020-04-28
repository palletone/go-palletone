package packetcc

import (
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)


type Tokens struct {
	Amount  uint64 `json:"amount"`  //数量
	Asset   *modules.Asset `json:"asset"`   //资产
}

type TokensJson struct {
	Amount  decimal.Decimal `json:"amount"`  //数量
	Asset   string `json:"asset"`   //资产
}

type PacketMultiToken struct {
	PubKey          []byte         //红包对应的公钥，也是红包的唯一标识
	Creator         common.Address //红包发放人员地址
	Tokens           []*Tokens //红包中的TokenID
	Amount          uint64         //红包总金额
	Count           uint32         //红包数，为0表示可以无限领取
	MinPacketAmount uint64         //单个红包最小额
	MaxPacketAmount uint64         //单个红包最大额,最大额最小额相同，则说明不是随机红包,0则表示完全随机
	ExpiredTime     uint64         //红包过期时间，0表示永不过期
	Remark          string         //红包的备注
	Constant        bool           //是否固定数额
}
type AllJson struct {
	*PacketJson
	*PacketMultiTokenJson
}

type AllJsonAll struct {
	One []*PacketJson
	Multi []*PacketMultiTokenJson
}

type PacketMultiTokenJson struct {
	PubKey          string          //红包对应的公钥，也是红包的唯一标识
	Creator         common.Address  //红包发放人员地址
	Token           []*TokensJson          //红包中的TokenID
	TotalAmount     decimal.Decimal //红包总金额
	PacketCount     uint32          //红包数，为0表示可以无限领取
	MinPacketAmount decimal.Decimal //单个红包最小额
	MaxPacketAmount decimal.Decimal //单个红包最大额,最大额最小额相同，则说明不是随机红包
	ExpiredTime     string          //红包过期时间，0表示永不过期
	Remark          string          //红包的备注
	IsConstant      string          //是否固定数额
	BalanceAmount   decimal.Decimal //红包剩余额度
	BalanceCount    uint32          //红包剩余次数
}


func convertPacketMultiToken2Json(packet *PacketMultiToken, balanceAmount uint64, balanceCount uint32) *PacketMultiTokenJson {
	js := &PacketMultiTokenJson{
		PubKey:          hex.EncodeToString(packet.PubKey),
		Creator:         packet.Creator,
		TotalAmount:     packet.Tokens[0].Asset.DisplayAmount(packet.Amount),
		MinPacketAmount: packet.Tokens[0].Asset.DisplayAmount(packet.MinPacketAmount),
		MaxPacketAmount: packet.Tokens[0].Asset.DisplayAmount(packet.MaxPacketAmount),
		PacketCount:     packet.Count,
		Remark:          packet.Remark,
		IsConstant:      strconv.FormatBool(packet.Constant),
		BalanceAmount:   packet.Tokens[0].Asset.DisplayAmount(balanceAmount),
		BalanceCount:    balanceCount,
	}
	if packet.ExpiredTime != 0 {
		js.ExpiredTime = time.Unix(int64(packet.ExpiredTime), 0).String()
	}
	for i,t := range packet.Tokens {
		js.Token[i].Amount = t.Asset.DisplayAmount(t.Amount)
		js.Token[i].Asset = t.Asset.String()
	}
	return js
}

type AllPacket struct {
	One []*Packet
	Multi []*PacketMultiToken
}

// 支持多 token 的红包
func (p *PacketMgr) SupportMultiToken(stub shim.ChaincodeStubInterface, creator common.Address, invokeTokens []*modules.InvokeTokens, pubKey []byte, count uint32,
	minAmount, maxAmount decimal.Decimal ,expiredTime uint64, remark string, isConstant bool) error {
	// 获取当前 pubKey 对应的红包
	pk, _ := getPacket(stub, pubKey)
	if pk != nil {
		return errors.New("PubKey already exist")
	}
	packet := &PacketMultiToken{
		PubKey:          pubKey,
		Creator:         creator,
		Count:           count,
		MinPacketAmount: invokeTokens[0].Asset.Uint64Amount(minAmount),
		MaxPacketAmount: invokeTokens[0].Asset.Uint64Amount(maxAmount),
		Remark:          remark,
		Constant:        isConstant,
		ExpiredTime:expiredTime,
	}
	a := uint64(0)
	for i,t := range invokeTokens {
		a += t.Amount
		packet.Tokens[i].Amount = t.Amount
		packet.Tokens[i].Asset = t.Asset
	}
	packet.Amount = a

	// 保存红包
	err := savePacketMultiToken(stub, packet)
	if err != nil {
		return err
	}
	// 保存红包的余额及个数
	err = savePacketBalance(stub, pubKey, packet.Amount, packet.Count)
	if err != nil {
		return err
	}
	return nil
}

// 保存红包
func savePacketMultiToken(stub shim.ChaincodeStubInterface, p *PacketMultiToken) error {
	key := PacketPrefix + hex.EncodeToString(p.PubKey)
	value, err := rlp.EncodeToBytes(p)
	if err != nil {
		return err
	}
	return stub.PutState(key, value)
}

// 获取红包
func getPacketMultiToken(stub shim.ChaincodeStubInterface, pubKey []byte) (*PacketMultiToken, error) {
	key := PacketPrefix + hex.EncodeToString(pubKey)
	value, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil,nil
	}
	p := PacketMultiToken{}
	err = rlp.DecodeBytes(value, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// 获取所有红包
func getAllPackets(stub shim.ChaincodeStubInterface) (*AllPacket,error) {
	value, err := stub.GetStateByPrefix(PacketPrefix)
	if err != nil {
		return nil, err
	}
	ap := &AllPacket{}
	for _,pp := range value {
		p := Packet{}
		// 兼容
		err = rlp.DecodeBytes(pp.Value, &p)
		if err != nil {
			pm := PacketMultiToken{}
			// 当前
			err = rlp.DecodeBytes(pp.Value, &pm)
			if err != nil {
				return nil,err
			}
			ap.Multi = append(ap.Multi,&pm)
		}else {
			ap.One = append(ap.One,&p)
		}
	}
	return ap,nil
}
