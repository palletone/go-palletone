package modules

import (
	"time"
)

type Joint struct {
	Unit *Unit `json:"unit"`
	// obj skiplist

	Unsigned     string    `json:"unsigned"`
	CreationDate time.Time `json:"creation_date"`
}
