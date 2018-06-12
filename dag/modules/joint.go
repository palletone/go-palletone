package modules

import (
	"time"
)

type Joint struct {
	Unit *Unit  `json:"unit"`
	Ball string `json:"ball"`
	// obj skiplist
	Skiplist     []string  `json:"skiplist"`
	Unsigned     string    `json:"unsigned"`
	CreationDate time.Time `json:"creation_date"`
}
