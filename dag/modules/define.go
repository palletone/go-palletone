package modules

const (
	FREEUNITS     = "free_units"
	UNSTABLEUNITS = "unstable_units"
	STABLEUNITS   = "stable_units"
)

var (
	FreeUnitslist FreeUnitsList
	Unstableunits UnStableUnitsList
	Stableunits   StableUnitslist
)

type FreeUnits struct {
	Key  string
	Hash string
}

type FreeUnitsList []*FreeUnits

func (list FreeUnitsList) Len() int { return len(list) }
func (list FreeUnitsList) Less(i, j int) bool {
	if list[i].Hash < list[j].Hash {
		return true
	}
	return false
}
func (list FreeUnitsList) Swap(i, j int) {
	var temp *FreeUnits = list[i]
	list[i] = list[j]
	list[j] = temp
}
