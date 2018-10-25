package modules

//type Vote struct {
//	Result []byte // vote for some person or some thing
//}

type VoteInfo struct {
	VoteType uint8
	//VoteID      uint16
	VoteContent []byte
}

const TYPE_MEDIATOR = 0 //VoteContent = []byte(Common.Address)
