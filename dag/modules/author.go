package modules

type UnitAuthors struct {
	Unit           string `json:"unit"`
	BestParentUnit string `json:"best_parent_unit"`
	WitnessedLevel int64  `json:"witnessed_level"`
}
