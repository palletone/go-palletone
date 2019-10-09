package list

import "testing"

func TestChainCodeList(t *testing.T) {
	cc1 := &CCInfo{
		Id:      []byte("cc1test"),
		Name:    "cc1",
		Path:    "/root",
		Version: "v1",
		SysCC:   false,
	}
	cc2 := &CCInfo{
		Id:      []byte("cc2test"),
		Name:    "cc2",
		Path:    "/root",
		Version: "v2",
		SysCC:   false,
	}
	SetChaincode("cid", 1, cc1)
	SetChaincode("cid", 1, cc2)

	cl, err := GetChaincodeList("cid")
	if err != nil {
		t.Error("not find chainlist")
	}
	t.Logf("vsersion=%d", cl.Version)
	DelChaincode("cid", "cc1", "v1")

	for k, v := range cl.CClist {
		t.Logf("----%s:%v", k, *v)
	}

	cc, err := GetChaincode("cid", []byte("cc2test"), "")
	if err != nil {
		t.Error("not find chainlist")
	}
	t.Logf("%v", *cc)

	for k, v := range cl.CClist {
		t.Logf("%s:%v", k, *v)
	}
}
