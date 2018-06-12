package modules

type WitnessList struct {
	Unit string `json:"unit"`
}

type EarnedHeadersCommissionRecipient struct {
	Unit string `json:"unit"`
}

/* SELECT MAX(units.level) AS max_alt_level \n\
FROM units \n\
LEFT JOIN parenthoods ON units.unit=child_unit \n\
LEFT JOIN units AS punits ON parent_unit=punits.unit AND punits.witnessed_level >= units.witnessed_level \n\
WHERE units.unit IN(?) AND punits.unit IS NULL AND ( \n\
	SELECT COUNT(*) \n\
	FROM unit_witnesses \n\
	WHERE unit_witnesses.unit IN(units.unit, units.witness_list_unit) AND unit_witnesses.address IN(?) \n\
)>=?",  */
type UnitWitness struct {
	Unit    string `json:"unit"`
	Address string `json:"address"`
}

func GetWitnessList() *WitnessList{
	w :=new(WitnessList)
	return  w
	
}