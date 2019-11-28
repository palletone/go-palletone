package certficate

import "testing"

func TestGMGenCert(t *testing.T) {
	var c GmCertInfo
	var emai namesVar
	var ip ipsVar
	c.Address = "P1CkorqRjxs8cQeQ7NQkGcpiLMSSiJYMX1a"
	c.Country = []string{"China"}
	c.Locality = []string{"BeiJing"}
	c.Organization = []string{"PalletOne"}
	err :=emai.Set("lk2684753@163.com")
	err = ip.Set("123.126.106.82")
	c.EmailAddresses = emai
	c.IPAddresses = ip
	certbyte, err := c.GenSMCert()
	if err != nil {
		t.Log(err)
	}
	t.Log(certbyte)
}
