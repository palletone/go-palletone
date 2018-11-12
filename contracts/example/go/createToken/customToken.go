package createToken

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)


type Token struct {
	tokenID uint64 //inner identity
	Holder common.Address
	extraMsg []byte //global unique identity
}
type CustomToken struct  {
	tokenName string
	tokenSymbol modules.IDType16
	owner common.Address
	totalSupply  uint64
	currentIdx uint64
	Members map[common.Address][]uint64
	inventory map[uint64]Token

	//optional field
	transFee uint
	expiryTimeStamp uint64
	Decimals uint
}
//IsOwner(const):Anyone
func (ct *CustomToken)IsOwner(address common.Address)bool{
	return ct.owner == address
}
//ChangeTotalSupply(danger):Owner
func (ct *CustomToken)ChangeTotalSupply(address common.Address,amount uint64)bool{
	if !ct.IsOwner(address){
		return false
	}
	if amount < ct.currentIdx {
		return false
	}
	ct.totalSupply = amount
	return true
}
//ChangeOwner(danger):Owner
func (ct *CustomToken)ChangeOwner(address common.Address, des common.Address)bool {
	if !ct.IsOwner(address)&& des != address{
		return false
	}
	ct.owner = des
	return true
}
//OwnerOf(const):AnyOne
func (ct *CustomToken)OwnerOf(address common.Address,TokenID uint64)common.Address {
	return ct.inventory[TokenID].Holder
}
//TotalSupply(const):AnyOne
func (ct *CustomToken)TotalSupply(address common.Address) uint64{
	return ct.totalSupply
}
//Name(const):AnyOne
func (ct *CustomToken)Name(address common.Address)string{
	return ct.tokenName
}
//Symbol(const):AnyOne
func (ct *CustomToken)Symbol(address common.Address)modules.IDType16{
	return ct.tokenSymbol
}

func(ct *CustomToken)balanceOf(account common.Address)[]uint64 {
	return ct.Members[account]
}


//Transfer(from common.Address , to common.Address, tokens []TokenID )
//transfer(to common.Address, tokens []TokenID)
//transferFrom(from common.Address, to common.Address, tokens []TokenID)

type CustomTokenConstructor struct {

}

func (ctor *CustomTokenConstructor) CreateToken(contractName string, tokenSymbol string, ownerAddress common.Address, recipientAddress common.Address) {

}
