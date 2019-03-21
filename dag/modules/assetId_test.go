package modules

import "testing"

func TestIDType16_String(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.String())
	t.Log(NewPTNAsset().String())
}

func TestIDType16_Str(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.String())
}
