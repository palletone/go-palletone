package modules


type VoteInfo struct {
	VoteType uint8
	//VoteID      uint16
	VoteContent []byte
}

const TYPE_MEDIATOR = 0 // VoteContent = []byte(Common.Address)
