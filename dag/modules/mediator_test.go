package modules

import (
	"encoding/json"

	"testing"
)

func TestMediatorJsonConvert(t *testing.T) {
	js := `{"account":"P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N","initPubKey":"34k7awAPrFTS8cvbr498EFcgMRBzhkdk3S8guurUashd6jZL3QwBM2qs16gnXVSc7R4cgXwMyvQTGq7AXGrURVz42qV7MXh6tyFtgkaZuj3zEnzE5D2VZ9EYUNRxbQy88YqTjBjXkaQZNhoCcQtMGARYDFnra9vUtLfz82K3AaPwVGe","node":"pnode://514058fa1a1949a044b0537b6f788b0672abc081cd2f85b9923b9123641663d79be0c2bea1bb0af704eeb0265e39228254688453c35bdb5d55b1c89f561134ca@119.3.213.46:30303","applyInfo":"{公司名称:北京矢链科技有限公司，宣传网站:http://www.vector.link}"}`
	mco := MediatorCreateOperation{}
	err := json.Unmarshal([]byte(js), &mco)
	t.Logf("%#v", mco)
	t.Log(err)
}
