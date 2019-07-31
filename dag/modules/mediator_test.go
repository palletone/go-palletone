package modules

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/core"
)

func TestMediatorJsonConvert(t *testing.T) {
	oldMi := NewOldMediatorCreateOp()
	oldMi.AddStr = "P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"
	oldMi.InitPubKey = "34k7awAPrFTS8cvbr498EFcgMRBzhkdk3S8guurUashd6jZL3QwBM2qs16gnXVSc7R4cgXwMyvQTGq7AXGrURVz42qV7MXh6tyFtgkaZuj3zEnzE5D2VZ9EYUNRxbQy88YqTjBjXkaQZNhoCcQtMGARYDFnra9vUtLfz82K3AaPwVGe"
	oldMi.Node = "pnode://514058fa1a1949a044b0537b6f788b0672abc081cd2f85b9923b9123641663d79be0c2bea1bb0af704eeb0265e39228254688453c35bdb5d55b1c89f561134ca@119.3.213.46:30303"
	oldMi.ApplyInfo = "北京矢链科技有限公司; http://www.vector.link"

	jsData, err := json.Marshal(oldMi)
	t.Logf("%#v", string(jsData))
	t.Log(err)

	mco := &MediatorCreateOperation{}
	err = json.Unmarshal(jsData, mco)
	t.Logf("%#v, MediatorInfoBase:%#v, MediatorApplyInfo:%#v", mco, mco.MediatorInfoBase, mco.MediatorApplyInfo)
	t.Log(err)
}

func TestMediatorRlpConvert(t *testing.T) {
	oldMi := NewOldMediatorCreateOp()
	oldMi.AddStr = "P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"
	oldMi.InitPubKey = "34k7awAPrFTS8cvbr498EFcgMRBzhkdk3S8guurUashd6jZL3QwBM2qs16gnXVSc7R4cgXwMyvQTGq7AXGrURVz42qV7MXh6tyFtgkaZuj3zEnzE5D2VZ9EYUNRxbQy88YqTjBjXkaQZNhoCcQtMGARYDFnra9vUtLfz82K3AaPwVGe"
	oldMi.Node = "pnode://514058fa1a1949a044b0537b6f788b0672abc081cd2f85b9923b9123641663d79be0c2bea1bb0af704eeb0265e39228254688453c35bdb5d55b1c89f561134ca@119.3.213.46:30303"
	oldMi.ApplyInfo = "北京矢链科技有限公司; http://www.vector.link"

	rlpData, err := rlp.EncodeToBytes(oldMi)
	t.Logf("%#v", string(rlpData))
	t.Log(err)

	mco := &MediatorCreateOperation{}
	err = rlp.DecodeBytes(rlpData, mco)
	t.Logf("%#v, MediatorInfoBase:%#v, MediatorApplyInfo:%#v", mco, mco.MediatorInfoBase, mco.MediatorApplyInfo)
	t.Log(err)
}

type OldMediatorCreateOperation struct {
	*core.MediatorInfoBase
	*OldMediatorApplyInfo
}

func NewOldMediatorCreateOp() *OldMediatorCreateOperation {
	return &OldMediatorCreateOperation{
		MediatorInfoBase:     &core.MediatorInfoBase{},
		OldMediatorApplyInfo: &OldMediatorApplyInfo{},
	}
}

type OldMediatorApplyInfo struct {
	ApplyInfo string `json:"applyInfo"` // 节点信息描述
}
