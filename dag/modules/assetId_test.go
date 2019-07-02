package modules

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDType16_String(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.String())
	t.Log(NewPTNAsset().String())
}

func TestAssetId_GetDecimal(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.GetDecimal())
}

func TestIDType16_Str(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.String())
	t.Logf("PTN hex:%#x", ptn.AssetId.Bytes())
}
func TestAssetId_String(t *testing.T) {
	token := "DEVIN"
	assetId, _, err := String2AssetId(token)
	assert.Nil(t, err)
	t.Logf("%#v", assetId)
	assert.Equal(t, token, assetId.String())
}
func TestAssetIdSlicsJson(t *testing.T) {
	tokenStrs := "[\"PTN\",\"DEVIN+805IERQX6QQ54N1MOB\",\"ABC+I05IERQX6QQ54N1MOB\"]"
	tokens := []AssetId{}
	err := json.Unmarshal([]byte(tokenStrs), &tokens)
	assert.Nil(t, err)
	for _, token := range tokens {
		t.Log(token.String())
	}
	data, err := json.Marshal(tokens)
	assert.Nil(t, err)
	t.Log(string(data))
}
