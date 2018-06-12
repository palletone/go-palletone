package modules

type ValidationState struct {
	Sequence             string   `json:"sequence"`
	ArrDoubleSpendInputs []string `json:"arrDoubleSpendInputs"`
	ArrAdditionalQueries []string `json:"arrAdditionalQueries"`
}
