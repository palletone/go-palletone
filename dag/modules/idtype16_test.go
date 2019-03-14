package modules

import "testing"

func TestIDType16_String(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.String())
	t.Log(NewPTNAsset().String())
}

func TestIDType16_Str(t *testing.T) {
	ptn := NewPTNAsset()
	t.Log(ptn.AssetId.Str())
}
func TestIDType16_StringFriendly(t *testing.T) {
	uid := &IDType16{0xff, 0xff, 0xff, 0xff}
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_Sequence))
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_Uuid))
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_UserDefine))

}
