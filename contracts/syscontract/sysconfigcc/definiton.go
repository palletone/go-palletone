package sysconfigcc

import (
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

const symbolsKey = "symbol_"

//one topic
type VoteTopic struct {
	TopicTitle    string
	SelectOptions []string
	SelectMax     uint64
}

//topic support result
type TopicSupports struct {
	TopicTitle  string
	VoteResults []VoteResult
	SelectMax   uint64
	//SelectOptionsNum  uint64
}

type VoteResult struct {
	SelectOption string
	Num          uint64
}

//vote token information
type TokenInfo struct {
	Name        string
	Symbol      string
	CreateAddr  string
	VoteType    byte
	TotalSupply uint64
	VoteEndTime time.Time
	VoteContent []byte
	AssetID     modules.AssetId
}

type SupportResult struct {
	TopicIndex  uint64
	TopicTitle  string
	VoteResults []VoteResult
}

type TokenIDInfo struct {
	IsVoteEnd      bool
	CreateAddr     string
	TotalSupply    uint64
	SupportResults []SupportResult
	AssetID        string
}
