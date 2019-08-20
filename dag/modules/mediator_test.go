package modules

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/core"
)

func TestMediatorJsonConvert(t *testing.T) {
	oldMi := NewMediatorCreateArgs100()
	oldMi.AddStr = "P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"
	oldMi.InitPubKey = "34k7awAPrFTS8cvbr498EFcgMRBzhkdk3S8guurUashd6jZL3QwBM2qs16gnXVSc7R4cgXwMyvQTGq7AXGrURVz" +
		"42qV7MXh6tyFtgkaZuj3zEnzE5D2VZ9EYUNRxbQy88YqTjBjXkaQZNhoCcQtMGARYDFnra9vUtLfz82K3AaPwVGe"
	oldMi.Node = "pnode://514058fa1a1949a044b0537b6f788b0672abc081cd2f85b9923b9123641663d79be0c2bea1bb0a" +
		"f704eeb0265e39228254688453c35bdb5d55b1c89f561134ca@119.3.213.46:30303"
	oldMi.ApplyInfo = "北京矢链科技有限公司; http://www.vector.link"

	jsData, err := json.Marshal(oldMi)
	t.Logf("%#v", string(jsData))
	t.Log(err)

	mco := &MediatorCreateArgs{}
	err = json.Unmarshal(jsData, mco)
	t.Logf("%#v, \n MediatorInfoBase:%#v, \n MediatorApplyInfo:%#v", mco,
		mco.MediatorInfoBase, mco.MediatorApplyInfo)
	t.Log(err)
}

func TestMediatorRlpConvert(t *testing.T) {
	oldMi := NewMediatorCreateArgs100()
	oldMi.AddStr = "P1NzevLMVCFJKWr4KAcHxyyh9xXaVU8yv3N"
	oldMi.InitPubKey = "34k7awAPrFTS8cvbr498EFcgMRBzhkdk3S8guurUashd6jZL3QwBM2qs16gnXVSc7R4cgXwMyvQTGq7AX" +
		"GrURVz42qV7MXh6tyFtgkaZuj3zEnzE5D2VZ9EYUNRxbQy88YqTjBjXkaQZNhoCcQtMGARYDFnra9vUtLfz82K3AaPwVGe"
	oldMi.Node = "pnode://514058fa1a1949a044b0537b6f788b0672abc081cd2f85b9923b9123641663d79be" +
		"0c2bea1bb0af704eeb0265e39228254688453c35bdb5d55b1c89f561134ca@119.3.213.46:30303"
	oldMi.ApplyInfo = "北京矢链科技有限公司; http://www.vector.link"

	rlpData, err := rlp.EncodeToBytes(oldMi)
	t.Logf("%#v", string(rlpData))
	t.Log(err)

	mco := &MediatorCreateArgs{}
	err = rlp.DecodeBytes(rlpData, mco)
	t.Logf("%#v, \n MediatorInfoBase:%#v, \n MediatorApplyInfo:%#v", mco,
		mco.MediatorInfoBase, mco.MediatorApplyInfo)
	t.Log(err)
}

type MediatorCreateArgs100 struct {
	*core.MediatorInfoBase
	*MediatorApplyInfo100
}

func NewMediatorCreateArgs100() *MediatorCreateArgs100 {
	return &MediatorCreateArgs100{
		MediatorInfoBase:     &core.MediatorInfoBase{},
		MediatorApplyInfo100: &MediatorApplyInfo100{},
	}
}

type MediatorApplyInfo100 struct {
	ApplyInfo string `json:"applyInfo"` // 节点信息描述
}

func TestMediatorUpdate(t *testing.T) {
	var args MediatorUpdateArgs
	args.AddStr = "P1xxx"
	name := "某节点"
	args.Name = &name
	t.Logf("%#v, \n account: %v \n logo: %v \n name: %v", args, args.AddStr, args.Logo, *args.Name)

	jsonData, err := json.Marshal(&args)
	t.Log(string(jsonData))
	t.Log(err)

	var args1 MediatorUpdateArgs
	err = json.Unmarshal(jsonData, &args1)
	t.Logf("%#v, \n account: %v \n logo: %v \n name: %v", args1, args1.AddStr, args1.Logo, *args1.Name)
	t.Log(err)

	rlpData, err := rlp.EncodeToBytes(&args)
	t.Log(string(rlpData))
	t.Log(err)

	var args2 MediatorUpdateArgs
	err = rlp.DecodeBytes(rlpData, &args2)
	t.Logf("%#v, \n account: %v \n logo: %v \n name: %v", args2, args2.AddStr, args2.Logo, *args2.Name)
	t.Log(err)
}
