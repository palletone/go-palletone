// ball structure and apis
package modules

import (
	"log"
	"time"
)

type Ball struct {
	Ball               string    `json:"ball"`
	Unit               string    `json:"unit"`
	ParentBalls        []string  `json:"parent_balls"` // 基于父units生成的 球的哈希数组
	IsNonserial        bool      `json:"is_nonserial"`
	SkiplistBalls      []string  `json:"skiplist_balls"`       // 该球的前驱球数组，用于构建跳表
	CreationDate       time.Time `json:"creation_date"`        // 生成时间
	CountPaidWitnesses bool      `json:"count_paid_witnesses"` // 见证节点们已支付
}
type Balls []*Ball

func Getlastball(unit string) *Ball {
	log.Println("hash(unit):", unit)
	ball := new(Ball)
	return ball
}
